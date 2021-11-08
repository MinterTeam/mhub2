package mhub2

import (
	"sort"
	"time"

	"github.com/MinterTeam/mhub2/module/x/mhub2/keeper"
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker is called at the beginning of every block
// NOTE: begin blocker also emits events which are helpful for
// clients listening to the chain and creating transactions
// based on the events (i.e. orchestrators)
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	for _, chainId := range k.GetChains(ctx) {
		cleanupTimedOutBatchTxs(ctx, chainId, k)
		cleanupTimedOutContractCallTxs(ctx, chainId, k)
		createSignerSetTxs(ctx, chainId, k)
		createBatchTxs(ctx, chainId, k)
		pruneSignerSetTxs(ctx, chainId, k)
	}
}

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	for _, chainId := range k.GetChains(ctx) {
		outgoingTxSlashing(ctx, chainId, k)
		eventVoteRecordTally(ctx, chainId, k)
		refundExpiredTxs(ctx, chainId, k)
	}
}

func refundExpiredTxs(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	if ctx.BlockHeight()%3 == 0 { // todo: make the period configurable ?
		k.IterateUnbatchedSendToExternals(ctx, chainId, func(ste *types.SendToExternal) bool {
			if time.Unix(int64(ste.CreatedAt), 0).Add(time.Hour).Before(ctx.BlockTime()) { // todo: make timeout configurable
				k.OnOutgoingTransactionTimeouts(ctx, chainId, ste.Id, ste.Sender)
			}
			return false
		})
	}
}

func createBatchTxs(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	if ctx.BlockHeight()%10 == 0 { // todo: make the period configurable ?
		coinIds := map[string]bool{}
		k.IterateUnbatchedSendToExternals(ctx, chainId, func(ste *types.SendToExternal) bool {
			coinIds[ste.Token.ExternalTokenId] = true
			return false
		})

		var ids []string // todo: sort
		for k := range coinIds {
			ids = append(ids, k)
		}

		for _, id := range ids {
			// NOTE: this doesn't emit events which would be helpful for client processes
			k.BuildBatchTx(ctx, chainId, id, 100)
		}
	}
}

func createSignerSetTxs(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	// Auto signerset tx creation.
	// 1. If there are no signer set requests, create a new one.
	// 2. If there is at least one validator who started unbonding in current block. (we persist last unbonded block height in hooks.go)
	//      This will make sure the unbonding validator has to provide an ethereum signature to a new signer set tx
	//	    that excludes him before he completely Unbonds.  Otherwise he will be slashed
	// 3. If power change between validators of Current signer set and latest signer set request is > 5%
	latestSignerSetTx := k.GetLatestSignerSetTx(ctx, chainId)
	if latestSignerSetTx == nil {
		k.CreateSignerSetTx(ctx, chainId)
		return
	}

	lastUnbondingHeight := k.GetLastUnbondingBlockHeight(ctx)
	blockHeight := uint64(ctx.BlockHeight())
	powerDiff := k.CurrentSignerSet(ctx).PowerDiff(latestSignerSetTx.Signers)

	shouldCreate := (lastUnbondingHeight == blockHeight) || (powerDiff > 0.05) // todo?
	k.Logger(ctx).Info(
		"considering signer set tx creation",
		"blockHeight", blockHeight,
		"lastUnbondingHeight", lastUnbondingHeight,
		"latestSignerSetTx.Nonce", latestSignerSetTx.Nonce,
		"powerDiff", powerDiff,
		"chainId", chainId,
		"shouldCreate", shouldCreate,
	)

	if shouldCreate {
		k.CreateSignerSetTx(ctx, chainId)
	}
}

func pruneSignerSetTxs(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	params := k.GetParams(ctx)
	// Validator set pruning
	// prune all validator sets with a nonce less than the
	// last observed nonce, they can't be submitted any longer
	//
	// Only prune valsets after the signed valsets window has passed
	// so that slashing can occur the block before we remove them
	lastObserved := k.GetLastObservedSignerSetTx(ctx, chainId)
	currentBlock := uint64(ctx.BlockHeight())
	tooEarly := currentBlock < params.SignedSignerSetTxsWindow
	if lastObserved != nil && !tooEarly {
		earliestToPrune := currentBlock - params.SignedSignerSetTxsWindow
		for _, set := range k.GetSignerSetTxs(ctx, chainId) {
			if set.Nonce < lastObserved.Nonce && set.Height < earliestToPrune {
				k.DeleteOutgoingTx(ctx, chainId, set.GetStoreIndex(chainId))
			}
		}
	}
}

// Iterate over all attestations currently being voted on in order of nonce and
// "Observe" those who have passed the threshold. Break the loop once we see
// an attestation that has not passed the threshold
func eventVoteRecordTally(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	attmap := k.GetExternalEventVoteRecordMapping(ctx, chainId)

	// We make a slice with all the event nonces that are in the attestation mapping
	keys := make([]uint64, 0, len(attmap))
	for k := range attmap {
		keys = append(keys, k)
	}
	// Then we sort it
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// This iterates over all keys (event nonces) in the attestation mapping. Each value contains
	// a slice with one or more attestations at that event nonce. There can be multiple attestations
	// at one event nonce when validators disagree about what event happened at that nonce.
	for _, nonce := range keys {
		// This iterates over all attestations at a particular event nonce.
		// They are ordered by when the first attestation at the event nonce was received.
		// This order is not important.
		for _, att := range attmap[nonce] {
			// We check if the event nonce is exactly 1 higher than the last attestation that was
			// observed. If it is not, we just move on to the next nonce. This will skip over all
			// attestations that have already been observed.
			//
			// Once we hit an event nonce that is one higher than the last observed event, we stop
			// skipping over this conditional and start calling tryAttestation (counting votes)
			// Once an attestation at a given event nonce has enough votes and becomes observed,
			// every other attestation at that nonce will be skipped, since the lastObservedEventNonce
			// will be incremented.
			//
			// Then we go to the next event nonce in the attestation mapping, if there is one. This
			// nonce will once again be one higher than the lastObservedEventNonce.
			// If there is an attestation at this event nonce which has enough votes to be observed,
			// we skip the other attestations and move on to the next nonce again.
			// If no attestation becomes observed, when we get to the next nonce, every attestation in
			// it will be skipped. The same will happen for every nonce after that.
			if nonce == uint64(k.GetLastObservedEventNonce(ctx, chainId))+1 {
				k.TryEventVoteRecord(ctx, chainId, att)
			}
		}
	}
}

// cleanupTimedOutBatchTxs deletes batches that have passed their expiration on Ethereum
// keep in mind several things when modifying this function
// A) unlike nonces timeouts are not monotonically increasing, meaning batch 5 can have a later timeout than batch 6
//    this means that we MUST only cleanup a single batch at a time
// B) it is possible for ethereumHeight to be zero if no events have ever occurred, make sure your code accounts for this
// C) When we compute the timeout we do our best to estimate the Ethereum block height at that very second. But what we work with
//    here is the Ethereum block height at the time of the last Deposit or Withdraw to be observed. It's very important we do not
//    project, if we do a slowdown on ethereum could cause a double spend. Instead timeouts will *only* occur after the timeout period
//    AND any deposit or withdraw has occurred to update the Ethereum block height.
func cleanupTimedOutBatchTxs(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	externalHeight := k.GetLastObservedExternalBlockHeight(ctx, chainId).ExternalHeight
	k.IterateOutgoingTxsByType(ctx, chainId, types.BatchTxPrefixByte, func(key []byte, otx types.OutgoingTx) bool {
		btx, _ := otx.(*types.BatchTx)

		if btx.Timeout < externalHeight {
			k.CancelBatchTx(ctx, chainId, btx.ExternalTokenId, btx.BatchNonce)
		}

		return false
	})
}

// cleanupTimedOutBatchTxs deletes logic calls that have passed their expiration on Ethereum
// keep in mind several things when modifying this function
// A) unlike nonces timeouts are not monotonically increasing, meaning call 5 can have a later timeout than batch 6
//    this means that we MUST only cleanup a single call at a time
// B) it is possible for ethereumHeight to be zero if no events have ever occurred, make sure your code accounts for this
// C) When we compute the timeout we do our best to estimate the Ethereum block height at that very second. But what we work with
//    here is the Ethereum block height at the time of the last Deposit or Withdraw to be observed. It's very important we do not
//    project, if we do a slowdown on ethereum could cause a double spend. Instead timeouts will *only* occur after the timeout period
//    AND any deposit or withdraw has occurred to update the Ethereum block height.
func cleanupTimedOutContractCallTxs(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	externalHeight := k.GetLastObservedExternalBlockHeight(ctx, chainId).ExternalHeight
	k.IterateOutgoingTxsByType(ctx, chainId, types.ContractCallTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
		cctx, _ := otx.(*types.ContractCallTx)
		if cctx.Timeout < externalHeight {
			k.DeleteOutgoingTx(ctx, chainId, cctx.GetStoreIndex(chainId))
		}
		return true
	})
}

func outgoingTxSlashing(ctx sdk.Context, chainId types.ChainID, k keeper.Keeper) {
	params := k.GetParams(ctx)
	maxHeight := uint64(0)
	if uint64(ctx.BlockHeight()) > params.SignedBatchesWindow {
		maxHeight = uint64(ctx.BlockHeight()) - params.SignedBatchesWindow
	} else {
		return
	}

	usotxs := k.GetUnSlashedOutgoingTxs(ctx, chainId, maxHeight)
	if len(usotxs) == 0 {
		return
	}

	// get signing info for each validator
	type valInfo struct {
		val   stakingtypes.Validator
		exist bool
		sigs  slashingtypes.ValidatorSigningInfo
		cons  sdk.ConsAddress
	}

	var valInfos []valInfo

	for _, val := range k.StakingKeeper.GetBondedValidatorsByPower(ctx) {
		consAddr, _ := val.GetConsAddr()
		sigs, exist := k.SlashingKeeper.GetValidatorSigningInfo(ctx, consAddr)
		valInfos = append(valInfos, valInfo{val, exist, sigs, consAddr})
	}

	var unbondingValInfos []valInfo

	blockTime := ctx.BlockTime().Add(k.StakingKeeper.GetParams(ctx).UnbondingTime)
	blockHeight := ctx.BlockHeight()
	unbondingValIterator := k.StakingKeeper.ValidatorQueueIterator(ctx, blockTime, blockHeight)
	defer unbondingValIterator.Close()

	// All unbonding validators
	for ; unbondingValIterator.Valid(); unbondingValIterator.Next() {
		unbondingValidators := k.GetUnbondingvalidators(unbondingValIterator.Value())
		for _, valAddr := range unbondingValidators.Addresses {
			addr, _ := sdk.ValAddressFromBech32(valAddr)
			validator, _ := k.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addr))
			valConsAddr, _ := validator.GetConsAddr()
			valSigningInfo, exist := k.SlashingKeeper.GetValidatorSigningInfo(ctx, valConsAddr)
			unbondingValInfos = append(unbondingValInfos, valInfo{validator, exist, valSigningInfo, valConsAddr})
		}
	}

	for _, otx := range usotxs {
		// SLASH BONDED VALIDATORS who didn't sign batch txs
		signatures := k.GetExternalSignatures(ctx, chainId, otx.GetStoreIndex(chainId))
		for _, valInfo := range valInfos {
			// Don't slash validators who joined after outgoingtx is created
			if valInfo.exist && valInfo.sigs.StartHeight < int64(otx.GetCosmosHeight()) {
				if _, ok := signatures[valInfo.val.GetOperator().String()]; !ok {
					if !valInfo.val.IsJailed() {
						k.StakingKeeper.Slash(
							ctx,
							valInfo.cons,
							ctx.BlockHeight(),
							valInfo.val.ConsensusPower(k.PowerReduction),
							params.SlashFractionBatch,
						)
						k.StakingKeeper.Jail(ctx, valInfo.cons)
					}
				}
			}
		}

		if sstx, ok := otx.(*types.SignerSetTx); ok {
			for _, valInfo := range unbondingValInfos {
				// Only slash validators who joined after valset is created and they are unbonding and UNBOND_SLASHING_WINDOW didn't pass
				if valInfo.exist && valInfo.sigs.StartHeight < int64(sstx.Nonce) && valInfo.val.IsUnbonding() && sstx.Height < uint64(valInfo.val.UnbondingHeight)+params.UnbondSlashingSignerSetTxsWindow {
					// Check if validator has confirmed valset or not
					if _, found := signatures[valInfo.val.GetOperator().String()]; !found {
						if !valInfo.val.IsJailed() {
							// TODO: do we want to slash jailed validators
							k.StakingKeeper.Slash(
								ctx,
								valInfo.cons,
								ctx.BlockHeight(),
								valInfo.val.ConsensusPower(k.PowerReduction),
								params.SlashFractionSignerSetTx,
							)
							k.StakingKeeper.Jail(ctx, valInfo.cons)
						}
					}
				}
			}
		}

		// then we set the latest slashed outgoing tx block
		k.SetLastSlashedOutgoingTxBlockHeight(ctx, chainId, otx.GetCosmosHeight())
	}
}
