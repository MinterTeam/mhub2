package keeper

import (
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetTxStatus(ctx sdk.Context, inTxHash string, status types.TxStatusType, outTxHash string) {
	newStatusType := status
	if k.GetTxStatus(ctx, inTxHash).Status == types.TX_STATUS_REFUNDED {
		newStatusType = types.TX_STATUS_REFUNDED
	}

	ctx.KVStore(k.storeKey).Set(types.GetTxStatusKey(inTxHash), k.cdc.MustMarshal(&types.TxStatus{
		InTxHash:  inTxHash,
		OutTxHash: outTxHash,
		Status:    newStatusType,
	}))
}

func (k Keeper) GetTxStatus(ctx sdk.Context, inTxHash string) *types.TxStatus {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.GetTxStatusKey(inTxHash))

	if len(bytes) == 0 {
		return &types.TxStatus{
			InTxHash: inTxHash,
			Status:   types.TX_STATUS_NOT_FOUND,
		}
	}

	var status types.TxStatus
	k.cdc.MustUnmarshal(bytes, &status)

	return &status
}
