package keeper

import (
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetTxFeeRecord(ctx sdk.Context, inTxHash string, record types.TxFeeRecord) {
	ctx.KVStore(k.storeKey).Set(types.GetTxFeeRecordKey(inTxHash), k.cdc.MustMarshal(&record))
}

func (k Keeper) GetTxFeeRecord(ctx sdk.Context, inTxHash string) *types.TxFeeRecord {
	bytes := ctx.KVStore(k.storeKey).Get(types.GetTxFeeRecordKey(inTxHash))

	if len(bytes) == 0 {
		return nil
	}

	var feeRecord types.TxFeeRecord
	k.cdc.MustUnmarshal(bytes, &feeRecord)

	return &feeRecord
}
