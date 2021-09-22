package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

// createSendToExternal
// - checks a counterpart denominator exists for the given voucher type
// - burns the voucher for transfer amount and fees
// - persists an OutgoingTx
// - adds the TX to the `available` TX pool via a second index
func (k Keeper) createSendToExternal(ctx sdk.Context, chainId types.ChainID, sender sdk.AccAddress, counterpartReceiver string, amount sdk.Coin, fee sdk.Coin, valCommission sdk.Coin, txHash string) (uint64, error) {
	totalAmount := amount.Add(fee).Add(valCommission)
	totalInVouchers := sdk.Coins{totalAmount}

	tokenInfo, err := k.DenomToTokenInfoLookup(ctx, chainId, totalAmount.Denom)
	if err != nil {
		return 0, err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, totalInVouchers); err != nil {
		return 0, err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, totalInVouchers); err != nil {
		panic(err)
	}

	// get next tx id from keeper
	nextID := k.incrementLastSendToExternalIDKey(ctx, chainId)

	convertedAmount := k.ConvertToExternalValue(ctx, chainId, tokenInfo.ExternalTokenId, amount.Amount)
	convertedFee := k.ConvertToExternalValue(ctx, chainId, tokenInfo.ExternalTokenId, fee.Amount)
	convertedValCommission := k.ConvertToExternalValue(ctx, chainId, tokenInfo.ExternalTokenId, valCommission.Amount)

	// set the outgoing tx in the pool index
	k.setUnbatchedSendToExternal(ctx, chainId, &types.SendToExternal{
		Id:                nextID,
		Sender:            sender.String(),
		ExternalRecipient: counterpartReceiver,
		Token:             types.NewSDKIntExternalToken(convertedAmount, tokenInfo.Id, tokenInfo.ExternalTokenId),
		Fee:               types.NewSDKIntExternalToken(convertedFee, tokenInfo.Id, tokenInfo.ExternalTokenId),
		ValCommission:     types.NewSDKIntExternalToken(convertedValCommission, tokenInfo.Id, tokenInfo.ExternalTokenId),
		ChainId:           chainId.String(),
		TxHash:            txHash,
	})

	return nextID, nil
}

// cancelSendToExternal
// - checks that the provided tx actually exists
// - deletes the unbatched tx from the pool
// - issues the tokens back to the sender
func (k Keeper) cancelSendToExternal(ctx sdk.Context, chainId types.ChainID, id uint64, s string) error {
	sender, _ := sdk.AccAddressFromBech32(s)

	var send *types.SendToExternal
	for _, ste := range k.getUnbatchedSendToExternals(ctx, chainId) {
		if ste.Id == id {
			send = ste
		}
	}
	if send == nil {
		// NOTE: this case will also be hit if the transaction is in a batch
		return sdkerrors.Wrap(types.ErrInvalid, "id not found in send to external pool")
	}

	if sender.String() != send.Sender {
		return fmt.Errorf("can't cancel a message you didn't send")
	}

	totalToRefund := send.Token.HubCoin(func(id uint64) (string, error) {
		info, err := k.TokenIdToTokenInfoLookup(ctx, id)
		if err != nil {
			return "", err
		}

		return info.Denom, nil
	})
	totalToRefund.Amount = totalToRefund.Amount.Add(send.Fee.Amount)
	totalToRefund.Amount = k.ConvertFromExternalValue(ctx, chainId, send.Token.ExternalTokenId, totalToRefund.Amount)

	totalToRefundCoins := sdk.NewCoins(totalToRefund)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, totalToRefundCoins); err != nil {
		return sdkerrors.Wrapf(err, "mint vouchers coins: %s", totalToRefundCoins)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, totalToRefundCoins); err != nil {
		return sdkerrors.Wrap(err, "sending coins from module account")
	}

	k.deleteUnbatchedSendToExternal(ctx, chainId, send.Id, send.Fee)
	return nil
}

func (k Keeper) setUnbatchedSendToExternal(ctx sdk.Context, chainId types.ChainID, ste *types.SendToExternal) {
	ctx.KVStore(k.storeKey).Set(types.MakeSendToExternalKey(chainId, ste.Id, ste.Fee), k.cdc.MustMarshal(ste))
}

func (k Keeper) deleteUnbatchedSendToExternal(ctx sdk.Context, chainId types.ChainID, id uint64, fee types.ExternalToken) {
	ctx.KVStore(k.storeKey).Delete(types.MakeSendToExternalKey(chainId, id, fee))
}

func (k Keeper) iterateUnbatchedSendToExternalsByCoin(ctx sdk.Context, chainId types.ChainID, externalTokenId string, cb func(external *types.SendToExternal) bool) {
	iter := prefix.NewStore(ctx.KVStore(k.storeKey), bytes.Join([][]byte{{types.SendToExternalKey}, chainId.Bytes(), []byte(externalTokenId)}, []byte{})).ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var ste types.SendToExternal
		k.cdc.MustUnmarshal(iter.Value(), &ste)
		if cb(&ste) {
			break
		}
	}
}

func (k Keeper) IterateUnbatchedSendToExternals(ctx sdk.Context, chainId types.ChainID, cb func(*types.SendToExternal) bool) {
	iter := prefix.NewStore(ctx.KVStore(k.storeKey), append([]byte{types.SendToExternalKey}, chainId.Bytes()...)).ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var ste types.SendToExternal
		k.cdc.MustUnmarshal(iter.Value(), &ste)
		if cb(&ste) {
			break
		}
	}
}

func (k Keeper) getUnbatchedSendToExternals(ctx sdk.Context, chainId types.ChainID) []*types.SendToExternal {
	var out []*types.SendToExternal
	k.IterateUnbatchedSendToExternals(ctx, chainId, func(ste *types.SendToExternal) bool {
		out = append(out, ste)
		return false
	})
	return out
}

func (k Keeper) incrementLastSendToExternalIDKey(ctx sdk.Context, chainId types.ChainID) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte{types.LastSendToExternalIDKey}, chainId.Bytes()...)
	bz := store.Get(key)
	var id uint64 = 0
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	newId := id + 1
	bz = sdk.Uint64ToBigEndian(newId)
	store.Set(key, bz)
	return newId
}
