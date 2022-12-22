package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	errors2 "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

const HubDecimals = 18

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	StakingKeeper          types.StakingKeeper
	ExternalEventProcessor interface {
		Handle(sdk.Context, types.ChainID, types.ExternalEvent) error
	}

	storeKey       sdk.StoreKey
	paramSpace     paramtypes.Subspace
	cdc            codec.Codec
	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	SlashingKeeper types.SlashingKeeper
	oracleKeeper   types.OracleKeeper
	PowerReduction sdk.Int
	hooks          types.MhubHooks
}

// NewKeeper returns a new instance of the mhub2 keeper
func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	accKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	slashingKeeper types.SlashingKeeper,
	oracleKeeper types.OracleKeeper,
	powerReduction sdk.Int,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	k := Keeper{
		cdc:            cdc,
		paramSpace:     paramSpace,
		storeKey:       storeKey,
		accountKeeper:  accKeeper,
		oracleKeeper:   oracleKeeper,
		bankKeeper:     bankKeeper,
		SlashingKeeper: slashingKeeper,
		PowerReduction: powerReduction,
	}

	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

/////////////////////////////
//     SignerSetTxNonce    //
/////////////////////////////

// incrementLatestSignerSetTxNonce sets the latest valset nonce
func (k Keeper) incrementLatestSignerSetTxNonce(ctx sdk.Context, chainId types.ChainID) uint64 {
	current := k.GetLatestSignerSetTxNonce(ctx, chainId)
	next := current + 1
	ctx.KVStore(k.storeKey).Set(append([]byte{types.LatestSignerSetTxNonceKey}, chainId.Bytes()...), sdk.Uint64ToBigEndian(next))
	return next
}

// GetLatestSignerSetTxNonce returns the latest valset nonce
func (k Keeper) GetLatestSignerSetTxNonce(ctx sdk.Context, chainId types.ChainID) uint64 {
	if bz := ctx.KVStore(k.storeKey).Get(append([]byte{types.LatestSignerSetTxNonceKey}, chainId.Bytes()...)); bz != nil {
		return binary.BigEndian.Uint64(bz)
	}
	return 0
}

// GetLatestSignerSetTx returns the latest validator set in state
func (k Keeper) GetLatestSignerSetTx(ctx sdk.Context, chainId types.ChainID) *types.SignerSetTx {
	key := types.MakeSignerSetTxKey(chainId, k.GetLatestSignerSetTxNonce(ctx, chainId))
	otx := k.GetOutgoingTx(ctx, chainId, key)
	out, _ := otx.(*types.SignerSetTx)
	return out
}

//////////////////////////////
// LastUnbondingBlockHeight //
//////////////////////////////

// setLastUnbondingBlockHeight sets the last unbonding block height
func (k Keeper) setLastUnbondingBlockHeight(ctx sdk.Context, unbondingBlockHeight uint64) {
	ctx.KVStore(k.storeKey).Set([]byte{types.LastUnBondingBlockHeightKey}, sdk.Uint64ToBigEndian(unbondingBlockHeight))
}

// GetLastUnbondingBlockHeight returns the last unbonding block height
func (k Keeper) GetLastUnbondingBlockHeight(ctx sdk.Context) uint64 {
	if bz := ctx.KVStore(k.storeKey).Get([]byte{types.LastUnBondingBlockHeightKey}); len(bz) == 0 {
		return 0
	} else {
		return binary.BigEndian.Uint64(bz)
	}
}

///////////////////////////////
//     EXTERNAL SIGNATURES   //
///////////////////////////////

// getExternalSignature returns a valset confirmation by a nonce and validator address
func (k Keeper) getExternalSignature(ctx sdk.Context, chainId types.ChainID, storeIndex []byte, validator sdk.ValAddress) []byte {
	return ctx.KVStore(k.storeKey).Get(types.MakeExternalSignatureKey(chainId, storeIndex, validator))
}

// SetExternalSignature sets a valset confirmation
func (k Keeper) SetExternalSignature(ctx sdk.Context, chainId types.ChainID, sig types.ExternalTxConfirmation, val sdk.ValAddress) []byte {
	key := types.MakeExternalSignatureKey(chainId, sig.GetStoreIndex(chainId), val)
	ctx.KVStore(k.storeKey).Set(key, sig.GetSignature())
	return key
}

// GetExternalSignatures returns all external signatures for a given outgoing tx by store index
func (k Keeper) GetExternalSignatures(ctx sdk.Context, chainId types.ChainID, storeIndex []byte) map[string][]byte {
	var signatures = make(map[string][]byte)
	k.iterateExternalSignatures(ctx, chainId, storeIndex, func(val sdk.ValAddress, h []byte) bool {
		signatures[val.String()] = h
		return false
	})
	return signatures
}

// iterateExternalSignatures iterates through all valset confirms by nonce in ASC order
func (k Keeper) iterateExternalSignatures(ctx sdk.Context, chainId types.ChainID, storeIndex []byte, cb func(sdk.ValAddress, []byte) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), bytes.Join([][]byte{{types.ExternalSignatureKey}, chainId.Bytes(), storeIndex}, []byte{}))
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		if cb(iter.Key(), iter.Value()) {
			break
		}
	}
}

/////////////////////////
//  ORC -> VAL ADDRESS //
/////////////////////////

// SetOrchestratorValidatorAddress sets the Orchestrator key for a given validator.
func (k Keeper) SetOrchestratorValidatorAddress(ctx sdk.Context, chainId types.ChainID, val sdk.ValAddress, orchAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.MakeOrchestratorValidatorAddressKey(chainId, orchAddr)

	store.Set(key, val.Bytes())
}

// GetOrchestratorValidatorAddress returns the validator key associated with an
// orchestrator key.
func (k Keeper) GetOrchestratorValidatorAddress(ctx sdk.Context, chainId types.ChainID, orchAddr sdk.AccAddress) sdk.ValAddress {
	store := ctx.KVStore(k.storeKey)
	key := types.MakeOrchestratorValidatorAddressKey(chainId, orchAddr)

	return store.Get(key)
}

////////////////////////
// VAL -> ETH ADDRESS //
////////////////////////

// setValidatorExternalAddress sets the ethereum address for a given validator
func (k Keeper) setValidatorExternalAddress(ctx sdk.Context, chainId types.ChainID, valAddr sdk.ValAddress, ethAddr common.Address) {
	store := ctx.KVStore(k.storeKey)
	key := types.MakeValidatorExternalAddressKey(chainId, valAddr)

	store.Set(key, ethAddr.Bytes())
}

// GetValidatorExternalAddress returns the eth address for a given mhub2 validator.
func (k Keeper) GetValidatorExternalAddress(ctx sdk.Context, chainId types.ChainID, valAddr sdk.ValAddress) common.Address {
	store := ctx.KVStore(k.storeKey)
	key := types.MakeValidatorExternalAddressKey(chainId, valAddr)

	return common.BytesToAddress(store.Get(key))
}

func (k Keeper) getValidatorsByExternalAddress(ctx sdk.Context, chainId types.ChainID, ethAddr common.Address) (vals []sdk.ValAddress) {
	iter := ctx.KVStore(k.storeKey).Iterator(nil, nil)

	for ; iter.Valid(); iter.Next() {
		if common.BytesToAddress(iter.Value()) == ethAddr {
			valBs := bytes.TrimPrefix(iter.Key(), []byte{types.ValidatorExternalAddressKey})
			if !bytes.HasPrefix(valBs, chainId.Bytes()) {
				continue
			}
			valBs = bytes.TrimPrefix(valBs, chainId.Bytes())

			val := sdk.ValAddress(valBs)
			vals = append(vals, val)
		}
	}

	return
}

////////////////////////
// ETH -> ORC ADDRESS //
////////////////////////

// setExternalOrchestratorAddress sets the eth orch addr mapping
func (k Keeper) setExternalOrchestratorAddress(ctx sdk.Context, chainId types.ChainID, ethAddr common.Address, orch sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.MakeExternalOrchestratorAddressKey(chainId, ethAddr)

	store.Set(key, orch.Bytes())
}

// GetExternalOrchestratorAddress gets the orch address for a given eth address
func (k Keeper) GetExternalOrchestratorAddress(ctx sdk.Context, chainId types.ChainID, ethAddr common.Address) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	key := types.MakeExternalOrchestratorAddressKey(chainId, ethAddr)

	return store.Get(key)
}

func (k Keeper) getExternalAddressesByOrchestrator(ctx sdk.Context, chainId types.ChainID, orch sdk.AccAddress) (ethAddrs []common.Address) {
	iter := ctx.KVStore(k.storeKey).Iterator(nil, nil)

	for ; iter.Valid(); iter.Next() {
		if sdk.AccAddress(iter.Value()).String() == orch.String() {
			ethBs := bytes.TrimPrefix(iter.Key(), []byte{types.ExternalOrchestratorAddressKey})
			if !bytes.HasPrefix(ethBs, chainId.Bytes()) {
				continue
			}
			ethBs = bytes.TrimPrefix(ethBs, chainId.Bytes())

			ethAddr := common.BytesToAddress(ethBs)
			ethAddrs = append(ethAddrs, ethAddr)
		}
	}

	return
}

// CreateSignerSetTx gets the current signer set from the staking keeper, increments the nonce,
// creates the signer set tx object, emits an event and sets the signer set in state
func (k Keeper) CreateSignerSetTx(ctx sdk.Context, chainId types.ChainID) *types.SignerSetTx {
	nonce := k.incrementLatestSignerSetTxNonce(ctx, chainId)
	currSignerSet := k.CurrentSignerSet(ctx, chainId)
	newSignerSetTx := types.NewSignerSetTx(nonce, uint64(ctx.BlockHeight()), currSignerSet)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMultisigUpdateRequest,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyContract, k.getBridgeContractAddress(ctx)),
			sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.getBridgeChainID(ctx)))),
			sdk.NewAttribute(types.AttributeKeySignerSetNonce, fmt.Sprint(nonce)),
		),
	)
	k.SetOutgoingTx(ctx, chainId, newSignerSetTx)
	k.Logger(ctx).Info(
		"SignerSetTx created",
		"nonce", newSignerSetTx.Nonce,
		"height", newSignerSetTx.Height,
		"signers", len(newSignerSetTx.Signers),
	)
	return newSignerSetTx
}

// CurrentSignerSet gets powers from the store and normalizes them
// into an integer percentage with a resolution of uint32 Max meaning
// a given validators 'mhub2 power' is computed as
// Cosmos power / total cosmos power = x / uint32 Max
// where x is the voting power on the mhub2 contract. This allows us
// to only use integer division which produces a known rounding error
// from truncation equal to the ratio of the validators
// Cosmos power / total cosmos power ratio, leaving us at uint32 Max - 1
// total voting power. This is an acceptable rounding error since floating
// point may cause consensus problems if different floating point unit
// implementations are involved.
func (k Keeper) CurrentSignerSet(ctx sdk.Context, chainId types.ChainID) types.ExternalSigners {
	validators := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	externalSigners := make([]*types.ExternalSigner, 0)
	var totalPower uint64
	for _, validator := range validators {
		val := validator.GetOperator()

		p := uint64(k.StakingKeeper.GetLastValidatorPower(ctx, val))

		if extAddr := k.GetValidatorExternalAddress(ctx, chainId, val); extAddr.Hex() != "0x0000000000000000000000000000000000000000" {
			es := &types.ExternalSigner{Power: p, ExternalAddress: extAddr.Hex()}
			externalSigners = append(externalSigners, es)
			totalPower += p
		}
	}
	// normalize power values
	for i := range externalSigners {
		externalSigners[i].Power = sdk.NewUint(externalSigners[i].Power).MulUint64(math.MaxUint32).QuoUint64(totalPower).Uint64()
	}

	return externalSigners
}

// GetSignerSetTxs returns all the signer set txs from the store
func (k Keeper) GetSignerSetTxs(ctx sdk.Context, chainId types.ChainID) (out []*types.SignerSetTx) {
	k.IterateOutgoingTxsByType(ctx, chainId, types.SignerSetTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
		sstx, _ := otx.(*types.SignerSetTx)
		out = append(out, sstx)
		return false
	})
	return
}

/////////////////////////////
//       PARAMETERS        //
/////////////////////////////

// GetParams returns the parameters from the store
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}

// setParams sets the parameters in the store
func (k Keeper) setParams(ctx sdk.Context, ps types.Params) {
	k.paramSpace.SetParamSet(ctx, &ps)
}

// getBridgeContractAddress returns the bridge contract address on ETH
func (k Keeper) getBridgeContractAddress(ctx sdk.Context) string {
	var a string
	k.paramSpace.Get(ctx, types.ParamsStoreKeyBridgeContractAddress, &a)
	return a
}

// getBridgeChainID returns the chain id of the ETH chain we are running against
func (k Keeper) getBridgeChainID(ctx sdk.Context) uint64 {
	var a uint64
	k.paramSpace.Get(ctx, types.ParamsStoreKeyBridgeContractChainID, &a)
	return a
}

func (k Keeper) GetChains(ctx sdk.Context) []types.ChainID {
	params := k.GetParams(ctx)

	var chains []types.ChainID
	for _, chain := range params.GetChains() {
		chains = append(chains, types.ChainID(chain))
	}

	return chains
}

// getGravityID returns the GravityID the GravityID is essentially a salt value
// for bridge signatures, provided each chain running Mhub2 has a unique ID
// it won't be possible to play back signatures from one bridge onto another
// even if they share a validator set.
//
// The lifecycle of the GravityID is that it is set in the Genesis file
// read from the live chain for the contract deployment, once a Mhub2 contract
// is deployed the GravityID CAN NOT BE CHANGED. Meaning that it can't just be the
// same as the chain id since the chain id may be changed many times with each
// successive chain in charge of the same bridge
func (k Keeper) getGravityID(ctx sdk.Context) string {
	var a string
	k.paramSpace.Get(ctx, types.ParamsStoreKeyGravityID, &a)
	return a
}

// getDelegateKeys iterates both the EthAddress and Orchestrator address indexes to produce
// a vector of MsgDelegateKeys entries containing all the delgate keys for state
// export / import. This may seem at first glance to be excessively complicated, why not combine
// the EthAddress and Orchestrator address indexes and simply iterate one thing? The answer is that
// even though we set the Eth and Orchestrator address in the same place we use them differently we
// always go from Orchestrator address to Validator address and from validator address to Ethereum address
// we want to keep looking up the validator address for various reasons, so a direct Orchestrator to Ethereum
// address mapping will mean having to keep two of the same data around just to provide lookups.
//
// For the time being this will serve
func (k Keeper) getDelegateKeys(ctx sdk.Context, chainId types.ChainID) (out []*types.MsgDelegateKeys) {
	store := ctx.KVStore(k.storeKey)
	iter := prefix.NewStore(store, append([]byte{types.ValidatorExternalAddressKey}, chainId.Bytes()...)).Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		out = append(out, &types.MsgDelegateKeys{
			ValidatorAddress: sdk.ValAddress(iter.Key()).String(),
			ExternalAddress:  common.BytesToAddress(iter.Value()).Hex(),
			ChainId:          chainId.String(),
		})
	}
	iter.Close()

	for _, msg := range out {
		msg.EthSignature = []byte{0}
		msg.OrchestratorAddress = k.GetExternalOrchestratorAddress(ctx, chainId, common.HexToAddress(msg.ExternalAddress)).String()
	}

	// we iterated over a map, so now we have to sort to ensure the
	// output here is deterministic, eth address chosen for no particular
	// reason
	sort.Slice(out[:], func(i, j int) bool {
		return out[i].ExternalAddress < out[j].ExternalAddress
	})

	return out
}

func (k Keeper) getNonces(ctx sdk.Context, chainId types.ChainID) (out []*types.Nonce) {
	store := ctx.KVStore(k.storeKey)
	iter := prefix.NewStore(store, append([]byte{types.LastEventNonceByValidatorKey}, chainId.Bytes()...)).Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		out = append(out, &types.Nonce{
			ValidatorAddress: sdk.ValAddress(iter.Key()).String(),
			LastEventNonce:   binary.BigEndian.Uint64(iter.Value()),
		})
	}
	iter.Close()

	// we iterated over a map, so now we have to sort to ensure the
	// output here is deterministic, eth address chosen for no particular
	// reason
	sort.Slice(out[:], func(i, j int) bool {
		return out[i].ValidatorAddress < out[j].ValidatorAddress
	})

	return out
}

// GetUnbondingvalidators returns UnbondingValidators.
// Adding here in mhub2 keeper as cdc is available inside endblocker.
func (k Keeper) GetUnbondingvalidators(unbondingVals []byte) stakingtypes.ValAddresses {
	unbondingValidators := stakingtypes.ValAddresses{}
	k.cdc.MustUnmarshal(unbondingVals, &unbondingValidators)
	return unbondingValidators
}

/////////////////
// OUTGOING TX //
/////////////////

func (k Keeper) GetOutgoingTx(ctx sdk.Context, chainId types.ChainID, storeIndex []byte) (out types.OutgoingTx) {
	if err := k.cdc.UnmarshalInterface(ctx.KVStore(k.storeKey).Get(types.MakeOutgoingTxKey(chainId, storeIndex)), &out); err != nil {
		panic(err)
	}
	return out
}

func (k Keeper) SetOutgoingTx(ctx sdk.Context, chainId types.ChainID, outgoing types.OutgoingTx) {
	outgoing.SetSequence(k.incrementOutgoingSequence(ctx, chainId))

	any, err := types.PackOutgoingTx(outgoing)
	if err != nil {
		panic(err)
	}
	ctx.KVStore(k.storeKey).Set(
		types.MakeOutgoingTxKey(chainId, outgoing.GetStoreIndex(chainId)),
		k.cdc.MustMarshal(any),
	)
}

// DeleteOutgoingTx deletes a given outgoingtx
func (k Keeper) DeleteOutgoingTx(ctx sdk.Context, chainId types.ChainID, storeIndex []byte) {
	ctx.KVStore(k.storeKey).Delete(types.MakeOutgoingTxKey(chainId, storeIndex))
}

func (k Keeper) PaginateOutgoingTxsByType(ctx sdk.Context, chainId types.ChainID, pageReq *query.PageRequest, prefixByte byte, cb func(key []byte, outgoing types.OutgoingTx) bool) (*query.PageResponse, error) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.MakeOutgoingTxKey(chainId, []byte{prefixByte}))

	return query.FilteredPaginate(prefixStore, pageReq, func(key []byte, value []byte, accumulate bool) (bool, error) {
		if !accumulate {
			return false, nil
		}

		var any cdctypes.Any
		k.cdc.MustUnmarshal(value, &any)
		var otx types.OutgoingTx
		if err := k.cdc.UnpackAny(&any, &otx); err != nil {
			panic(err)
		}
		if accumulate {
			return cb(key, otx), nil
		}

		return false, nil
	})
}

// IterateOutgoingTxsByType iterates over a specific type of outgoing transaction denoted by the chosen prefix byte
func (k Keeper) IterateOutgoingTxsByType(ctx sdk.Context, chainId types.ChainID, prefixByte byte, cb func(key []byte, outgoing types.OutgoingTx) (stop bool)) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.MakeOutgoingTxKey(chainId, []byte{prefixByte}))
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var any cdctypes.Any
		k.cdc.MustUnmarshal(iter.Value(), &any)
		var otx types.OutgoingTx
		if err := k.cdc.UnpackAny(&any, &otx); err != nil {
			panic(err)
		}
		if cb(iter.Key(), otx) {
			break
		}
	}
}

// iterateOutgoingTxs iterates over a specific type of outgoing transaction denoted by the chosen prefix byte
func (k Keeper) iterateOutgoingTxs(ctx sdk.Context, chainId types.ChainID, cb func(key []byte, outgoing types.OutgoingTx) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), append([]byte{types.OutgoingTxKey}, chainId.Bytes()...))
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var any cdctypes.Any
		k.cdc.MustUnmarshal(iter.Value(), &any)
		var otx types.OutgoingTx
		if err := k.cdc.UnpackAny(&any, &otx); err != nil {
			panic(err)
		}
		if cb(iter.Key(), otx) {
			break
		}
	}
}

// GetLastObservedSignerSetTx retrieves the last observed validator set from the store
func (k Keeper) GetLastObservedSignerSetTx(ctx sdk.Context, chainId types.ChainID) *types.SignerSetTx {
	key := append([]byte{types.LastObservedSignerSetKey}, chainId.Bytes()...)
	if val := ctx.KVStore(k.storeKey).Get(key); val != nil {
		var out types.SignerSetTx
		k.cdc.MustUnmarshal(val, &out)
		return &out
	}
	return nil
}

// setLastObservedSignerSetTx updates the last observed validator set in the stor e
func (k Keeper) setLastObservedSignerSetTx(ctx sdk.Context, chainId types.ChainID, signerSet types.SignerSetTx) {
	key := append([]byte{types.LastObservedSignerSetKey}, chainId.Bytes()...)
	ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshal(&signerSet))
}

// CreateContractCallTx xxx
func (k Keeper) CreateContractCallTx(ctx sdk.Context, chainId types.ChainID, invalidationNonce uint64, invalidationScope tmbytes.HexBytes,
	payload []byte, tokens []types.ExternalToken, fees []types.ExternalToken) *types.ContractCallTx {
	params := k.GetParams(ctx)

	newContractCallTx := &types.ContractCallTx{
		InvalidationNonce: invalidationNonce,
		InvalidationScope: invalidationScope,
		Address:           k.getBridgeContractAddress(ctx),
		Payload:           payload,
		Timeout:           params.TargetEthTxTimeout,
		Tokens:            tokens,
		Fees:              fees,
		Height:            uint64(ctx.BlockHeight()),
	}

	var tokenString []string
	for _, token := range tokens {
		tokenString = append(tokenString, token.String())
	}

	var feeString []string
	for _, fee := range fees {
		feeString = append(feeString, fee.String())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMultisigUpdateRequest,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyContract, k.getBridgeContractAddress(ctx)),
			sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.getBridgeChainID(ctx)))),
			sdk.NewAttribute(types.AttributeKeyContractCallInvalidationNonce, fmt.Sprint(invalidationNonce)),
			sdk.NewAttribute(types.AttributeKeyContractCallInvalidationScope, fmt.Sprint(invalidationScope)),
			sdk.NewAttribute(types.AttributeKeyContractCallPayload, string(payload)),
			sdk.NewAttribute(types.AttributeKeyContractCallTokens, strings.Join(tokenString, "|")),
			sdk.NewAttribute(types.AttributeKeyContractCallFees, strings.Join(feeString, "|")),
			sdk.NewAttribute(types.AttributeKeyEthTxTimeout, strconv.FormatUint(params.TargetEthTxTimeout, 10)),
		),
	)
	k.SetOutgoingTx(ctx, chainId, newContractCallTx)
	k.Logger(ctx).Info(
		"ContractCallTx created",
		"invalidation_nonce", newContractCallTx.InvalidationNonce,
		"invalidation_scope", newContractCallTx.InvalidationScope,
		// todo: fill out all fields
	)
	return newContractCallTx
}

func (k Keeper) GetTokenInfos(ctx sdk.Context) *types.TokenInfos {
	out := &types.TokenInfos{}
	if err := k.cdc.Unmarshal(ctx.KVStore(k.storeKey).Get([]byte{types.TokenInfosKey}), out); err != nil {
		panic(err)
	}

	return out
}

func (k Keeper) SetTokenInfos(ctx sdk.Context, tokenInfos *types.TokenInfos) {
	ctx.KVStore(k.storeKey).Set([]byte{types.TokenInfosKey}, k.cdc.MustMarshal(tokenInfos))
}

var ErrTokenNotFound = errors.New("token not found")

func (k Keeper) DenomToTokenInfoLookup(ctx sdk.Context, chainId types.ChainID, denom string) (*types.TokenInfo, error) {
	for _, info := range k.GetTokenInfos(ctx).TokenInfos {
		if info.Denom == denom && info.ChainId == chainId.String() {
			return info, nil
		}
	}

	return nil, errors2.Wrap(ErrTokenNotFound, fmt.Sprintf("chainId:%s denom:%s", chainId, denom))
}

func (k Keeper) TokenIdToTokenInfoLookup(ctx sdk.Context, tokenId uint64) (*types.TokenInfo, error) {
	for _, info := range k.GetTokenInfos(ctx).TokenInfos {
		if info.Id == tokenId {
			return info, nil
		}
	}

	return nil, errors2.Wrap(ErrTokenNotFound, fmt.Sprintf("tokenId:%d", tokenId))
}

func (k Keeper) ExternalIdToTokenInfoLookup(ctx sdk.Context, chainId types.ChainID, externalId string) (*types.TokenInfo, error) {
	for _, info := range k.GetTokenInfos(ctx).TokenInfos {
		if info.ChainId == chainId.String() && info.ExternalTokenId == externalId {
			return info, nil
		}
	}

	return nil, errors2.Wrap(ErrTokenNotFound, fmt.Sprintf("chainId:%s externalId:%s", chainId, externalId))
}

func (k Keeper) getOutgoingSequence(ctx sdk.Context, chainId types.ChainID) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte{types.OutgoingSequence}, chainId.Bytes()...)
	bz := store.Get(key)
	var id uint64 = 0
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}

	return id
}

func (k Keeper) setOutgoingSequence(ctx sdk.Context, chainId types.ChainID, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte{types.OutgoingSequence}, chainId.Bytes()...)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set(key, bz)
}

func (k Keeper) incrementOutgoingSequence(ctx sdk.Context, chainId types.ChainID) uint64 {
	newId := k.getOutgoingSequence(ctx, chainId) + 1
	k.setOutgoingSequence(ctx, chainId, newId)

	return newId
}

func (k Keeper) ConvertFromExternalValue(ctx sdk.Context, chainId types.ChainID, externalId string, amount sdk.Int) sdk.Int {
	coin, err := k.ExternalIdToTokenInfoLookup(ctx, chainId, externalId)
	if err != nil {
		return amount
	}

	return convertDecimals(coin.ExternalDecimals, HubDecimals, amount)
}

func (k Keeper) ConvertToExternalValue(ctx sdk.Context, chainId types.ChainID, externalId string, amount sdk.Int) sdk.Int {
	coin, err := k.ExternalIdToTokenInfoLookup(ctx, chainId, externalId)
	if err != nil {
		return amount
	}

	return convertDecimals(HubDecimals, coin.ExternalDecimals, amount)
}

func (k Keeper) GetCommissionForHolder(ctx sdk.Context, addresses []string, commission sdk.Dec) sdk.Dec {
	maxValue := sdk.NewInt(0)
	for _, address := range addresses {
		if len(address) > 2 && address[:2] == "0x" {
			address = address[2:]
		}
		maxValue = sdk.MaxInt(k.oracleKeeper.GetHolderValue(ctx, address), maxValue)
	}

	if !maxValue.IsPositive() {
		return commission
	}

	// 1 HUB -10%
	// 2 HUB -20%
	// 4 HUB -30%
	// 8 HUB -40%
	// 16 HUB -50%
	// 32 HUB -60%
	// 64 HUB -70%
	// 128 HUB -80%
	// 256 HUB -90%

	discount1 := convertDecimals(0, 18, sdk.NewInt(1))
	discount2 := convertDecimals(0, 18, sdk.NewInt(2))
	discount4 := convertDecimals(0, 18, sdk.NewInt(4))
	discount8 := convertDecimals(0, 18, sdk.NewInt(8))
	discount16 := convertDecimals(0, 18, sdk.NewInt(16))
	discount32 := convertDecimals(0, 18, sdk.NewInt(32))
	discount128 := convertDecimals(0, 18, sdk.NewInt(128))
	discount256 := convertDecimals(0, 18, sdk.NewInt(256))

	switch {
	case maxValue.GTE(discount256):
		return commission.Sub(commission.MulInt64(90).QuoInt64(100))
	case maxValue.GTE(discount128):
		return commission.Sub(commission.MulInt64(80).QuoInt64(100))
	case maxValue.GTE(discount32):
		return commission.Sub(commission.MulInt64(60).QuoInt64(100))
	case maxValue.GTE(discount16):
		return commission.Sub(commission.MulInt64(50).QuoInt64(100))
	case maxValue.GTE(discount8):
		return commission.Sub(commission.MulInt64(40).QuoInt64(100))
	case maxValue.GTE(discount4):
		return commission.Sub(commission.MulInt64(30).QuoInt64(100))
	case maxValue.GTE(discount2):
		return commission.Sub(commission.MulInt64(20).QuoInt64(100))
	case maxValue.GTE(discount1):
		return commission.Sub(commission.MulInt64(10).QuoInt64(100))
	}

	return commission
}

func (k Keeper) GetColdStorageAddr(ctx sdk.Context, chainId types.ChainID) string {
	switch chainId {
	case "minter":
		return "0x7072558b2b91e62dbed78e9a3453e5c9e01fec5e"
	case "ethereum":
		return "0x58BD8047F441B9D511aEE9c581aEb1caB4FE0b6d"
	case "bsc":
		return "0xbCc2Fa395c6198096855c932f4087cF1377d28EE"
	}

	panic("unknown network")
}

func (k Keeper) ColdStorageTransfer(ctx sdk.Context, c *types.ColdStorageTransferProposal) error {
	chainId := types.ChainID(c.ChainId)
	coldStorageAddr := k.GetColdStorageAddr(ctx, chainId)

	for _, coin := range c.Amount {
		vouchers := sdk.Coins{coin}
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, vouchers); err != nil {
			return errors2.Wrapf(err, "mint vouchers coins: %s", vouchers)
		}

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, types.TempAddress, vouchers); err != nil {
			return errors2.Wrap(err, "transfer vouchers")
		}

		txID, err := k.createSendToExternal(ctx, chainId, types.TempAddress, coldStorageAddr, coin, sdk.NewCoin(coin.Denom, sdk.NewInt(0)), sdk.NewCoin(coin.Denom, sdk.NewInt(0)), fmt.Sprintf("%x", sha256.Sum256(ctx.TxBytes())), "hub", types.TempAddress.String())
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvents([]sdk.Event{
			sdk.NewEvent(
				types.EventTypeBridgeWithdrawalReceived,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyContract, k.getBridgeContractAddress(ctx)),
				sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.getBridgeChainID(ctx)))),
				sdk.NewAttribute(types.AttributeKeyOutgoingTXID, strconv.Itoa(int(txID))),
				sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprint(txID)),
			),
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(types.AttributeKeyOutgoingTXID, fmt.Sprint(txID)),
			),
		})
	}

	return nil
}

func (k Keeper) OnOutgoingTransactionTimeouts(ctx sdk.Context, chainId types.ChainID, txId uint64, sender string) {
	k.cancelSendToExternal(ctx, chainId, txId, sender)
}

func (k Keeper) GetOutgoingTxTimeout(ctx sdk.Context) time.Duration {
	var a uint64
	k.paramSpace.Get(ctx, types.ParamOutgoingTxTimeout, &a)
	return time.Duration(a) * time.Millisecond
}

func (k Keeper) CheckChainID(ctx sdk.Context, id types.ChainID) error {
	for _, c := range k.GetChains(ctx) {
		if c.String() == id.String() {
			return nil
		}
	}

	return errors.New("invalid chain id")
}

func (k Keeper) SetStakingKeeper(keeper types.StakingKeeper) Keeper {
	k.StakingKeeper = keeper
	k.ExternalEventProcessor = ExternalEventProcessor{
		keeper:     k,
		bankKeeper: k.bankKeeper,
	}

	return k
}

func convertDecimals(fromDecimals uint64, toDecimals uint64, amount sdk.Int) sdk.Int {
	if fromDecimals == toDecimals {
		return amount
	}

	to := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(toDecimals)), nil)
	from := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(fromDecimals)), nil)

	result := amount.BigInt()
	result.Mul(result, to)
	result.Div(result, from)

	return sdk.NewIntFromBigInt(result)
}
