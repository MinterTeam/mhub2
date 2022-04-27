package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

// InitGenesis starts a chain from a genesis state
func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {
	k.setParams(ctx, *data.Params)
	k.SetTokenInfos(ctx, data.TokenInfos)

	for _, externalState := range data.ExternalStates {
		chainId := types.ChainID(externalState.ChainId)

		k.setOutgoingSequence(ctx, chainId, externalState.Sequence)

		// reset pool transactions in state
		for _, tx := range externalState.UnbatchedSendToExternalTxs {
			k.setUnbatchedSendToExternal(ctx, chainId, tx)
		}

		// reset external event vote records in state
		for _, evr := range externalState.ExternalEventVoteRecords {
			event, err := types.UnpackEvent(evr.Event)
			if err != nil {
				panic("couldn't cast to event")
			}
			if err := event.Validate(chainId); err != nil {
				panic("invalid event in genesis")
			}
			k.setExternalEventVoteRecord(ctx, chainId, event.GetEventNonce(), event.Hash(), evr)
		}

		// reset last observed event nonce
		k.setLastObservedEventNonce(ctx, chainId, externalState.LastObservedEventNonce)

		// reset attestation state of all validators
		for _, eventVoteRecord := range externalState.ExternalEventVoteRecords {
			event, _ := types.UnpackEvent(eventVoteRecord.Event)
			for _, vote := range eventVoteRecord.Votes {
				val, err := sdk.ValAddressFromBech32(vote)
				if err != nil {
					panic(err)
				}
				last := k.getLastEventNonceByValidator(ctx, chainId, val)
				if event.GetEventNonce() > last {
					k.setLastEventNonceByValidator(ctx, chainId, val, event.GetEventNonce())
				}
			}
		}

		// reset delegate keys in state
		for _, keys := range externalState.DelegateKeys {
			if err := keys.ValidateBasic(); err != nil {
				panic("Invalid delegate key in Genesis!")
			}

			val, _ := sdk.ValAddressFromBech32(keys.ValidatorAddress)
			orch, _ := sdk.AccAddressFromBech32(keys.OrchestratorAddress)
			eth := common.HexToAddress(keys.ExternalAddress)

			// set the orchestrator address
			k.SetOrchestratorValidatorAddress(ctx, chainId, val, orch)
			// set the ethereum address
			k.setValidatorExternalAddress(ctx, chainId, val, common.HexToAddress(keys.ExternalAddress))
			k.setExternalOrchestratorAddress(ctx, chainId, eth, orch)
		}

		// reset outgoing txs in state
		for _, ota := range externalState.OutgoingTxs {
			otx, err := types.UnpackOutgoingTx(ota)
			if err != nil {
				panic("invalid outgoing tx any in genesis file")
			}
			k.SetOutgoingTx(ctx, chainId, otx)
		}

		// reset signatures in state
		for _, confa := range externalState.Confirmations {
			conf, err := types.UnpackConfirmation(confa)
			if err != nil {
				panic("invalid etheruem signature in genesis")
			}
			// TODO: not currently an easy way to get the validator address from the
			// external address here. once we implement the third index for keys
			// this will be easy.
			k.SetExternalSignature(ctx, chainId, conf, sdk.ValAddress{})
		}
	}
}

// ExportGenesis exports all the state needed to restart the chain
// from the current state of the chain
func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	params := k.GetParams(ctx)
	chains := k.GetChains(ctx)
	tokenInfos := k.GetTokenInfos(ctx)
	state := types.GenesisState{
		Params:     &params,
		TokenInfos: tokenInfos,
	}

	for _, chainId := range chains {
		var (
			delegates    = k.getDelegateKeys(ctx, chainId)
			lastobserved = k.GetLastObservedEventNonce(ctx, chainId)
		)

		state.ExternalStates = append(state.ExternalStates, &types.ExternalState{
			ChainId:                chainId.String(),
			DelegateKeys:           delegates,
			LastObservedEventNonce: lastobserved,
			Sequence:               k.getOutgoingSequence(ctx, chainId),
		})
	}

	return state
}
