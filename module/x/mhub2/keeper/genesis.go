package keeper

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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
			if err := event.Validate(); err != nil {
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
			k.SetOrchestratorValidatorAddress(ctx, val, orch)
			// set the ethereum address
			k.setValidatorExternalAddress(ctx, val, common.HexToAddress(keys.ExternalAddress))
			k.setExternalOrchestratorAddress(ctx, eth, orch)
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
		TokenInfos: &tokenInfos,
	}

	for _, chainId := range chains {
		var (
			outgoingTxs              []*cdctypes.Any
			externalTxConfirmations  []*cdctypes.Any
			attmap                   = k.GetExternalEventVoteRecordMapping(ctx, chainId)
			externalEventVoteRecords []*types.ExternalEventVoteRecord
			delegates                = k.getDelegateKeys(ctx)
			lastobserved             = k.GetLastObservedEventNonce(ctx, chainId)
			unbatchedTransfers       = k.getUnbatchedSendToExternals(ctx, chainId)
		)

		// export ethereumEventVoteRecords from state
		for _, atts := range attmap {
			// TODO: set height = 0?
			externalEventVoteRecords = append(externalEventVoteRecords, atts...)
		}

		// export signer set txs and sigs
		k.IterateOutgoingTxsByType(ctx, chainId, types.SignerSetTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
			ota, _ := types.PackOutgoingTx(otx)
			outgoingTxs = append(outgoingTxs, ota)
			sstx, _ := otx.(*types.SignerSetTx)
			k.iterateExternalSignatures(ctx, chainId, sstx.GetStoreIndex(chainId), func(val sdk.ValAddress, sig []byte) bool {
				siga, _ := types.PackConfirmation(&types.SignerSetTxConfirmation{sstx.Nonce, k.GetValidatorExternalAddress(ctx, val).Hex(), sig})
				externalTxConfirmations = append(externalTxConfirmations, siga)
				return false
			})
			return false
		})

		// export batch txs and sigs
		k.IterateOutgoingTxsByType(ctx, chainId, types.BatchTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
			ota, _ := types.PackOutgoingTx(otx)
			outgoingTxs = append(outgoingTxs, ota)
			btx, _ := otx.(*types.BatchTx)
			k.iterateExternalSignatures(ctx, chainId, btx.GetStoreIndex(chainId), func(val sdk.ValAddress, sig []byte) bool {
				siga, _ := types.PackConfirmation(&types.BatchTxConfirmation{btx.ExternalTokenId, btx.BatchNonce, k.GetValidatorExternalAddress(ctx, val).Hex(), sig})
				externalTxConfirmations = append(externalTxConfirmations, siga)
				return false
			})
			return false
		})

		// export contract call txs and sigs
		k.IterateOutgoingTxsByType(ctx, chainId, types.ContractCallTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
			ota, _ := types.PackOutgoingTx(otx)
			outgoingTxs = append(outgoingTxs, ota)
			btx, _ := otx.(*types.ContractCallTx)
			k.iterateExternalSignatures(ctx, chainId, btx.GetStoreIndex(chainId), func(val sdk.ValAddress, sig []byte) bool {
				siga, _ := types.PackConfirmation(&types.ContractCallTxConfirmation{btx.InvalidationScope, btx.InvalidationNonce, k.GetValidatorExternalAddress(ctx, val).Hex(), sig})
				externalTxConfirmations = append(externalTxConfirmations, siga)
				return false
			})
			return false
		})

		state.ExternalStates = append(state.ExternalStates, &types.ExternalState{
			ChainId:                    chainId.String(),
			ExternalEventVoteRecords:   externalEventVoteRecords,
			DelegateKeys:               delegates,
			UnbatchedSendToExternalTxs: unbatchedTransfers,
			LastObservedEventNonce:     lastobserved,
			OutgoingTxs:                outgoingTxs,
			Confirmations:              externalTxConfirmations,
			Sequence:                   k.getOutgoingSequence(ctx, chainId),
		})
	}

	return state
}
