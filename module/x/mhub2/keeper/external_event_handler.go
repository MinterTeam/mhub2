package keeper

import (
	"fmt"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

// ExternalEventProcessor processes `accepted` EthereumEvents
type ExternalEventProcessor struct {
	keeper     Keeper
	bankKeeper types.BankKeeper
}

func (a ExternalEventProcessor) DetectMaliciousSupply(ctx sdk.Context, denom string, amount sdk.Int) (err error) {
	currentSupply := a.keeper.bankKeeper.GetSupply(ctx, denom)
	newSupply := new(big.Int).Add(currentSupply.Amount.BigInt(), amount.BigInt())
	if newSupply.BitLen() > 256 {
		return sdkerrors.Wrapf(types.ErrSupplyOverflow, "malicious supply of %s detected", denom)
	}

	return nil
}

// Handle is the entry point for ExternalEvent processing
func (a ExternalEventProcessor) Handle(ctx sdk.Context, chainId types.ChainID, eve types.ExternalEvent) (err error) {
	switch event := eve.(type) {
	case *types.TransferToChainEvent:
		if event.ReceiverChainId == "HUB" {
			receiver, err := sdk.AccAddressFromHex(event.ExternalReceiver)
			if err != nil {
				return err
			}

			return a.Handle(ctx, chainId, &types.SendToHubEvent{
				EventNonce:     event.EventNonce,
				ExternalCoinId: event.ExternalCoinId,
				Amount:         event.Amount.Add(event.Fee),
				Sender:         event.Sender,
				CosmosReceiver: receiver.String(),
				ExternalHeight: event.ExternalHeight,
				TxHash:         event.TxHash,
			})
		}

		tempReceiver, _ := sdk.AccAddressFromBech32("hub1xhaedvjeu88p5hrpgyugyy7stflm3nmqa0jhjc") // todo
		err := a.Handle(ctx, chainId, &types.SendToHubEvent{
			EventNonce:     event.EventNonce,
			ExternalCoinId: event.ExternalCoinId,
			Amount:         event.Amount.Add(event.Fee),
			Sender:         event.Sender,
			CosmosReceiver: tempReceiver.String(),
			ExternalHeight: event.ExternalHeight,
			TxHash:         event.TxHash,
		})
		if err != nil {
			panic(err)
		}

		senderChainTokenInfo, err := a.keeper.ExternalIdToTokenInfoLookup(ctx, chainId, event.ExternalCoinId)
		if err != nil {
			return err
		}

		receiverChainTokenInfo, err := a.keeper.DenomToTokenInfoLookup(ctx, types.ChainID(event.ReceiverChainId), senderChainTokenInfo.Denom)
		if err != nil {
			return err
		}

		// TODO: check decimals logic
		totalAmount := event.Amount.Add(event.Fee)
		commissionValue := a.keeper.GetCommissionForHolder(ctx, event.Sender, receiverChainTokenInfo.Commission).Mul(totalAmount.ToDec()).TruncateInt()
		fee := sdk.NewCoin(receiverChainTokenInfo.Denom, event.Fee)
		commission := sdk.NewCoin(receiverChainTokenInfo.Denom, commissionValue)
		amount := sdk.NewCoin(receiverChainTokenInfo.Denom, event.Amount).Sub(commission)

		txID, err := a.keeper.createSendToExternal(ctx, types.ChainID(event.ReceiverChainId), tempReceiver, event.ExternalReceiver, amount, fee, commission, event.TxHash, chainId, event.Sender)
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvents([]sdk.Event{
			sdk.NewEvent(
				types.EventTypeBridgeWithdrawalReceived,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyContract, a.keeper.getBridgeContractAddress(ctx)),
				sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(a.keeper.getBridgeChainID(ctx)))),
				sdk.NewAttribute(types.AttributeKeyOutgoingTXID, strconv.Itoa(int(txID))),
				sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprint(txID)),
			),
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(types.AttributeKeyOutgoingTXID, fmt.Sprint(txID)),
			),
		})

		return nil

	case *types.SendToHubEvent:
		tokenInfo, err := a.keeper.ExternalIdToTokenInfoLookup(ctx, chainId, event.ExternalCoinId)
		if err != nil {
			return err
		}
		addr, _ := sdk.AccAddressFromBech32(event.CosmosReceiver)
		convertedAmount := a.keeper.ConvertFromExternalValue(ctx, chainId, event.ExternalCoinId, event.Amount)
		coins := sdk.Coins{sdk.NewCoin(tokenInfo.Denom, convertedAmount)}

		if err := a.DetectMaliciousSupply(ctx, tokenInfo.Denom, convertedAmount); err != nil {
			return err
		}

		if err := a.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			return sdkerrors.Wrapf(err, "mint vouchers coins: %s", coins)
		}

		if err := a.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
			return err
		}
		a.keeper.AfterSendToHubEvent(ctx, *event)
		a.keeper.SetTxStatus(ctx, chainId, event.TxHash, types.TX_STATUS_DEPOSIT_RECEIVED, "")

		return nil

	case *types.BatchExecutedEvent:
		a.keeper.batchTxExecuted(ctx, chainId, event.ExternalCoinId, event.BatchNonce, event.TxHash, event.FeePaid, event.FeePayer)

		a.keeper.AfterBatchExecutedEvent(ctx, *event)
		return nil

	case *types.ContractCallExecutedEvent:
		a.keeper.AfterContractCallExecutedEvent(ctx, *event)
		return nil

	case *types.SignerSetTxExecutedEvent:
		// TODO here we should check the contents of the validator set against
		// the store, if they differ we should take some action to indicate to the
		// user that bridge highjacking has occurred
		a.keeper.setLastObservedSignerSetTx(ctx, chainId, types.SignerSetTx{
			Nonce:   event.SignerSetTxNonce,
			Signers: event.Members,
		})
		a.keeper.AfterSignerSetExecutedEvent(ctx, *event)
		return nil

	default:
		return sdkerrors.Wrapf(types.ErrInvalid, "event type: %T", event)
	}
}
