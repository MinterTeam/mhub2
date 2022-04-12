package keeper

import (
	"os"
	"path/filepath"
	"sync"

	db "github.com/tendermint/tm-db"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// todo: temporary solution. change to main store when major update will occur
var feeRecordDB = &db.GoLevelDB{}
var o = sync.Once{}

func initFeeRecordDB() {
	dir := os.ExpandEnv(filepath.Join("$HOME", ".mhub2", "data", "fee_record_db.db"))
	var err error
	feeRecordDB, err = db.NewGoLevelDB("fee_record_store", dir)
	if err != nil {
		panic(err)
	}
}

func (k Keeper) SetTxFeeRecord(_ sdk.Context, inTxHash string, record types.TxFeeRecord) {
	o.Do(initFeeRecordDB)
	feeRecordDB.Set(types.GetTxFeeRecordKey(inTxHash), k.cdc.MustMarshal(&record))
}

func (k Keeper) GetTxFeeRecord(_ sdk.Context, inTxHash string) *types.TxFeeRecord {
	o.Do(initFeeRecordDB)
	bytes, _ := feeRecordDB.Get(types.GetTxFeeRecordKey(inTxHash))

	if len(bytes) == 0 {
		return nil
	}

	var feeRecord types.TxFeeRecord
	k.cdc.MustUnmarshal(bytes, &feeRecord)

	return &feeRecord
}
