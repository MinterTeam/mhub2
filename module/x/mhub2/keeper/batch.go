package keeper

import (
	"encoding/binary"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BatchTxSize TODO: should we make this a parameter or a a call arg?
const BatchTxSize = 100

// BuildBatchTx starts the following process chain:
//   - find bridged denominator for given voucher type
//   - determine if a an unexecuted batch is already waiting for this token type, if so confirm the new batch would
//     have a higher total fees. If not exit withtout creating a batch
//   - select available transactions from the outgoing transaction pool sorted by fee desc
//   - persist an outgoing batch object with an incrementing ID = nonce
//   - emit an event
func (k Keeper) BuildBatchTx(ctx sdk.Context, chainId types.ChainID, externalTokenId string, maxElements int) *types.BatchTx {
	// if there is a more profitable batch for this token type do not create a new batch
	if lastBatch := k.getLastOutgoingBatchByTokenType(ctx, chainId, externalTokenId); lastBatch != nil {
		if lastBatch.GetFees().GTE(k.getBatchFeesByTokenType(ctx, chainId, externalTokenId, maxElements)) {
			//return nil
		}
	}

	var selectedStes []*types.SendToExternal
	k.iterateUnbatchedSendToExternalsByCoin(ctx, chainId, externalTokenId, func(ste *types.SendToExternal) bool {
		selectedStes = append(selectedStes, ste)
		k.deleteUnbatchedSendToExternal(ctx, chainId, ste.Id, ste.Fee)
		k.SetTxStatus(ctx, ste.TxHash, types.TX_STATUS_BATCH_CREATED, "")
		return len(selectedStes) == maxElements
	})

	batch := &types.BatchTx{
		BatchNonce:      k.incrementLastOutgoingBatchNonce(ctx, chainId),
		Timeout:         k.getBatchTimeoutHeight(ctx, chainId),
		Transactions:    selectedStes,
		ExternalTokenId: externalTokenId,
		Height:          uint64(ctx.BlockHeight()),
	}
	k.SetOutgoingTx(ctx, chainId, batch)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeOutgoingBatch,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprint(batch.BatchNonce)),
	))

	return batch
}

// This gets the batch timeout height in External blocks.
func (k Keeper) getBatchTimeoutHeight(ctx sdk.Context, chainId types.ChainID) uint64 {
	params := k.GetParams(ctx)
	var averageBlockTime uint64
	switch chainId {
	case "ethereum":
		averageBlockTime = params.AverageEthereumBlockTime
	case "bsc":
		averageBlockTime = params.AverageBscBlockTime
	case "metagarden":
		averageBlockTime = 5700
	case "minter":
		averageBlockTime = 5000
	case "hub":
		averageBlockTime = params.AverageBlockTime
	}

	currentCosmosHeight := ctx.BlockHeight()
	// we store the last observed Cosmos and External heights, we do not concern ourselves if these values are zero because
	// no batch can be produced if the last External block height is not first populated by a deposit event.
	heights := k.GetLastObservedExternalBlockHeight(ctx, chainId)
	if heights.CosmosHeight == 0 || heights.ExternalHeight == 0 {
		return 0
	}
	// we project how long it has been in milliseconds since the last External block height was observed
	projectedMillis := (uint64(currentCosmosHeight) - heights.CosmosHeight) * params.AverageBlockTime
	// we convert that projection into the current External height using the average External block time in millis
	projectedCurrentExternalHeight := (projectedMillis / averageBlockTime) + heights.ExternalHeight
	// we convert our target time for block timeouts (lets say 12 hours) into a number of blocks to
	// place on top of our projection of the current External block height.
	blocksToAdd := params.TargetEthTxTimeout / averageBlockTime
	return projectedCurrentExternalHeight + blocksToAdd
}

// batchTxExecuted is run when the Cosmos chain detects that a batch has been executed on External Chain
// It deletes all the transactions in the batch, then cancels all earlier batches
func (k Keeper) batchTxExecuted(ctx sdk.Context, chainId types.ChainID, externalTokenId string, nonce uint64, txHash string, feePaid sdk.Int, feePayer string) {
	otx := k.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, externalTokenId, nonce))
	batchTx, _ := otx.(*types.BatchTx)
	if chainId != "minter" {
		k.IterateOutgoingTxsByType(ctx, chainId, types.BatchTxPrefixByte, func(key []byte, otx types.OutgoingTx) bool {
			// If the iterated batches nonce is lower than the one that was just executed, cancel it
			btx, _ := otx.(*types.BatchTx)
			if (btx.BatchNonce < batchTx.BatchNonce) && (btx.ExternalTokenId == batchTx.ExternalTokenId) {
				k.CancelBatchTx(ctx, chainId, btx.ExternalTokenId, btx.BatchNonce)
			}
			return false
		})
	}
	k.DeleteOutgoingTx(ctx, chainId, batchTx.GetStoreIndex(chainId))

	tokenInfo, err := k.ExternalIdToTokenInfoLookup(ctx, chainId, batchTx.ExternalTokenId)
	if err != nil {
		panic(err)
	}

	totalValCommission := sdk.NewInt64Coin(tokenInfo.Denom, 0)
	totalFee := sdk.NewInt64Coin(tokenInfo.Denom, 0)
	for _, tx := range batchTx.Transactions {
		totalValCommission.Amount = totalValCommission.Amount.Add(tx.ValCommission.Amount)
		totalFee.Amount = totalFee.Amount.Add(tx.Fee.Amount)
		k.SetTxStatus(ctx, tx.TxHash, types.TX_STATUS_BATCH_EXECUTED, txHash)
		k.SetTxFeeRecord(ctx, tx.TxHash, types.TxFeeRecord{
			ValCommission: tx.ValCommission.Amount,
			ExternalFee:   tx.Fee.Amount,
		})
	}

	totalValCommission.Amount = k.ConvertFromExternalValue(ctx, chainId, tokenInfo.ExternalTokenId, totalValCommission.Amount)
	totalFee.Amount = k.ConvertFromExternalValue(ctx, chainId, tokenInfo.ExternalTokenId, totalFee.Amount)

	// pay val's commissions
	if totalValCommission.IsPositive() {
		valset := k.CurrentSignerSet(ctx, "minter")
		var totalPower uint64
		for _, val := range valset {
			totalPower += val.Power
		}

		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{totalValCommission}); err != nil {
			panic(sdkerrors.Wrapf(err, "mint vouchers coins: %s", sdk.Coins{totalValCommission}))
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, types.TempAddress, sdk.Coins{totalValCommission}); err != nil {
			panic(err)
		}

		for _, val := range valset {
			amount := totalValCommission.Amount.Mul(sdk.NewIntFromUint64(val.Power)).Quo(sdk.NewIntFromUint64(totalPower))
			_, err = k.createSendToExternal(ctx, "minter", types.TempAddress, val.ExternalAddress, sdk.NewCoin(tokenInfo.Denom, amount), sdk.NewInt64Coin(tokenInfo.Denom, 0), sdk.NewInt64Coin(tokenInfo.Denom, 0), "#commission", "", "")
			if err != nil {
				panic(err)
			}
		}
	}

	if totalFee.IsPositive() {
		var externalBaseCoin string

		switch chainId {
		case "ethereum":
			externalBaseCoin = "eth"
		case "bsc":
			externalBaseCoin = "bnb"
		case "metagarden":
			externalBaseCoin = "metagarden"
		default:
			return
		}

		amount := feePaid.ToDec().
			Mul(k.oracleKeeper.MustGetTokenPrice(ctx, externalBaseCoin)).
			Quo(k.oracleKeeper.MustGetTokenPrice(ctx, tokenInfo.Denom)).
			MulInt64(150).
			QuoInt64(100).
			TruncateInt()
		fee := sdk.NewCoin(tokenInfo.Denom, amount)
		if fee.IsGTE(totalFee) {
			fee = totalFee
		}

		if fee.IsPositive() {
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{fee}); err != nil {
				panic(sdkerrors.Wrapf(err, "mint vouchers coins: %s", sdk.Coins{fee}))
			}
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, types.TempAddress, sdk.Coins{fee}); err != nil {
				panic(err)
			}

			_, err = k.createSendToExternal(ctx, "minter", types.TempAddress, feePayer, fee, sdk.NewInt64Coin(fee.Denom, 0), sdk.NewInt64Coin(fee.Denom, 0), "#fee", "", "")
			if err != nil {
				panic(err)
			}

			feeLeft := totalFee.Sub(fee)
			if feeLeft.IsPositive() {
				if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.Coins{feeLeft}); err != nil {
					panic(sdkerrors.Wrapf(err, "mint vouchers coins: %s", sdk.Coins{feeLeft}))
				}
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, types.TempAddress, sdk.Coins{feeLeft}); err != nil {
					panic(err)
				}

				averageFeePaid := fee.Amount.QuoRaw(int64(len(batchTx.Transactions)))
				totalGoodFeePaid := sdk.NewInt(0)
				for _, tx := range batchTx.Transactions {
					convertedTxFee := k.ConvertFromExternalValue(ctx, chainId, tokenInfo.ExternalTokenId, tx.Fee.Amount)
					if convertedTxFee.GTE(averageFeePaid) {
						totalGoodFeePaid = totalGoodFeePaid.Add(convertedTxFee)
					}
				}

				for _, tx := range batchTx.Transactions {
					convertedTxFee := k.ConvertFromExternalValue(ctx, chainId, tokenInfo.ExternalTokenId, tx.Fee.Amount)
					if convertedTxFee.LT(averageFeePaid) {
						continue
					}

					toRefund := feeLeft.Amount.Mul(convertedTxFee).Quo(totalGoodFeePaid)
					if tx.RefundChainId != "minter" { // we only can refund fee to minter
						continue
					}

					if toRefund.IsPositive() {
						_, err = k.createSendToExternal(ctx, "minter", types.TempAddress, tx.RefundAddress, sdk.NewCoin(fee.Denom, toRefund), sdk.NewInt64Coin(fee.Denom, 0), sdk.NewInt64Coin(fee.Denom, 0), "#fee", "", "")
						if err != nil {
							panic(err)
						}

						record := k.GetTxFeeRecord(ctx, tx.TxHash)
						record.ExternalFee = record.ExternalFee.Sub(toRefund)
						k.SetTxFeeRecord(ctx, tx.TxHash, *record)
					}
				}
			}
		}
	}
}

// getBatchFeesByTokenType gets the fees the next batch of a given token type would
// have if created. This info is both presented to relayers for the purpose of determining
// when to request batches and also used by the batch creation process to decide not to create
// a new batch
func (k Keeper) getBatchFeesByTokenType(ctx sdk.Context, chainId types.ChainID, externalTokenId string, maxElements int) sdk.Int {
	feeAmount := sdk.ZeroInt()
	i := 0
	k.iterateUnbatchedSendToExternalsByCoin(ctx, chainId, externalTokenId, func(tx *types.SendToExternal) bool {
		feeAmount = feeAmount.Add(tx.Fee.Amount)
		i++
		return i == maxElements
	})

	return feeAmount
}

// GetBatchFeesByTokenType gets the fees the next batch of a given token type would
// have if created. This info is both presented to relayers for the purpose of determining
// when to request batches and also used by the batch creation process to decide not to create
// a new batch
func (k Keeper) GetBatchFeesByTokenType(ctx sdk.Context, chainId types.ChainID, externalTokenId string, maxElements int) sdk.Int {
	feeAmount := sdk.ZeroInt()
	i := 0
	k.iterateUnbatchedSendToExternalsByCoin(ctx, chainId, externalTokenId, func(tx *types.SendToExternal) bool {
		feeAmount = feeAmount.Add(tx.Fee.Amount)
		i++
		return i == maxElements
	})
	return feeAmount
}

// CancelBatchTx releases all TX in the batch and deletes the batch
func (k Keeper) CancelBatchTx(ctx sdk.Context, chainId types.ChainID, externalTokenId string, nonce uint64) {
	if chainId == "minter" {
		panic("CANNOT CANCEL MINTER BATCH")
	}

	otx := k.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, externalTokenId, nonce))
	batch, _ := otx.(*types.BatchTx)

	// free transactions from batch and reindex them
	for _, tx := range batch.Transactions {
		k.setUnbatchedSendToExternal(ctx, chainId, tx)
	}

	// Delete batch since it is finished
	k.DeleteOutgoingTx(ctx, chainId, batch.GetStoreIndex(chainId))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOutgoingBatchCanceled,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprint(nonce)),
		),
	)
}

// getLastOutgoingBatchByTokenType gets the latest outgoing tx batch by token type
func (k Keeper) getLastOutgoingBatchByTokenType(ctx sdk.Context, chainId types.ChainID, externalTokenId string) *types.BatchTx {
	var lastBatch *types.BatchTx = nil
	lastNonce := uint64(0)
	k.IterateOutgoingTxsByType(ctx, chainId, types.BatchTxPrefixByte, func(key []byte, otx types.OutgoingTx) bool {
		btx, _ := otx.(*types.BatchTx)
		if btx.ExternalTokenId == externalTokenId && btx.BatchNonce > lastNonce {
			lastBatch = btx
			lastNonce = btx.BatchNonce
		}
		return false
	})
	return lastBatch
}

// SetLastSlashedOutgoingTxBlockHeight sets the latest slashed Batch block height
func (k Keeper) SetLastSlashedOutgoingTxBlockHeight(ctx sdk.Context, chainId types.ChainID, blockHeight uint64) {
	ctx.KVStore(k.storeKey).Set(append([]byte{types.LastSlashedOutgoingTxBlockKey}, chainId.Bytes()...), sdk.Uint64ToBigEndian(blockHeight))
}

// GetLastSlashedOutgoingTxBlockHeight returns the latest slashed Batch block
func (k Keeper) GetLastSlashedOutgoingTxBlockHeight(ctx sdk.Context, chainId types.ChainID) uint64 {
	if bz := ctx.KVStore(k.storeKey).Get(append([]byte{types.LastSlashedOutgoingTxBlockKey}, chainId.Bytes()...)); bz == nil {
		return 0
	} else {
		return binary.BigEndian.Uint64(bz)
	}
}

func (k Keeper) GetUnSlashedOutgoingTxs(ctx sdk.Context, chainId types.ChainID, maxHeight uint64) (out []types.OutgoingTx) {
	lastSlashed := k.GetLastSlashedOutgoingTxBlockHeight(ctx, chainId)
	k.iterateOutgoingTxs(ctx, chainId, func(key []byte, otx types.OutgoingTx) bool {
		if (otx.GetCosmosHeight() < maxHeight) && (otx.GetCosmosHeight() > lastSlashed) {
			out = append(out, otx)
		}
		return false
	})
	return
}

func (k Keeper) setLastOutgoingBatchNonce(ctx sdk.Context, chainId types.ChainID, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte{types.LastOutgoingBatchNonceKey}, chainId.Bytes()...)
	store.Set(key, sdk.Uint64ToBigEndian(nonce))
}

func (k Keeper) getLastOutgoingBatchNonce(ctx sdk.Context, chainId types.ChainID) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte{types.LastOutgoingBatchNonceKey}, chainId.Bytes()...)
	bz := store.Get(key)
	var id uint64 = 0
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k Keeper) incrementLastOutgoingBatchNonce(ctx sdk.Context, chainId types.ChainID) uint64 {
	newId := k.getLastOutgoingBatchNonce(ctx, chainId) + 1
	k.setLastOutgoingBatchNonce(ctx, chainId, newId)

	return newId
}
