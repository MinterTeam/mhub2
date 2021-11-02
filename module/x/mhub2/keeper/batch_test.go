package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

var chainId = types.ChainID("ethereum")

func TestBatches(t *testing.T) {
	input := CreateTestEnv(t)
	ctx := input.Context
	tokenInfos := input.GravityKeeper.GetTokenInfos(ctx).TokenInfos
	var (
		now                 = time.Now().UTC()
		mySender, _         = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = tokenInfos[0].ExternalTokenId
		tokenId             = tokenInfos[0].Id
		allVouchers         = sdk.NewCoins(
			types.NewExternalToken(99999, tokenId, myTokenContractAddr).HubCoin(testDenomResolver),
		)
	)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, fundAccount(ctx, input.BankKeeper, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	input.AddSendToEthTxsToPool(t, ctx, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 2, 3, 2, 1)

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	firstBatch := input.GravityKeeper.BuildBatchTx(ctx, chainId, myTokenContractAddr, 2)

	// then batch is persisted
	gotFirstBatch := input.GravityKeeper.GetOutgoingTx(ctx, chainId, firstBatch.GetStoreIndex(chainId))
	require.NotNil(t, gotFirstBatch)

	gfb := gotFirstBatch.(*types.BatchTx)
	expFirstBatch := &types.BatchTx{
		BatchNonce: 1,
		Transactions: []*types.SendToExternal{
			types.NewSendToExternalTx(2, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 101, 3, 0, "#"),
			types.NewSendToExternalTx(3, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 102, 2, 0, "#"),
		},
		ExternalTokenId: myTokenContractAddr,
		Height:          1234567,
		Sequence:        2,
	}

	assert.Equal(t, expFirstBatch.Transactions, gfb.Transactions)

	// and verify remaining available Tx in the pool
	var gotUnbatchedTx []*types.SendToExternal
	input.GravityKeeper.IterateUnbatchedSendToExternals(ctx, chainId, func(tx *types.SendToExternal) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx := []*types.SendToExternal{
		types.NewSendToExternalTx(1, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 100, 2, 0, "#"),
		types.NewSendToExternalTx(4, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 103, 1, 0, "#"),
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// CREATE SECOND, MORE PROFITABLE BATCH
	// ====================================

	// add some more TX to the pool to create a more profitable batch
	input.AddSendToEthTxsToPool(t, ctx, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 4, 5)

	// create the more profitable batch
	ctx = ctx.WithBlockTime(now)
	// tx batch size is 2, so that some of them stay behind
	secondBatch := input.GravityKeeper.BuildBatchTx(ctx, chainId, myTokenContractAddr, 2)

	// check that the more profitable batch has the right txs in it
	expSecondBatch := &types.BatchTx{
		BatchNonce: 2,
		Transactions: []*types.SendToExternal{
			types.NewSendToExternalTx(6, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 101, 5, 0, "#"),
			types.NewSendToExternalTx(5, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 100, 4, 0, "#"),
		},
		ExternalTokenId: myTokenContractAddr,
		Height:          1234567,
		Sequence:        2,
	}

	assert.Equal(t, expSecondBatch, secondBatch)

	// EXECUTE THE MORE PROFITABLE BATCH
	// =================================

	// Execute the batch
	input.GravityKeeper.batchTxExecuted(ctx, chainId, myTokenContractAddr, secondBatch.BatchNonce, "", sdk.NewInt(0), "")

	// check batch has been deleted
	gotSecondBatch := input.GravityKeeper.GetOutgoingTx(ctx, chainId, secondBatch.GetStoreIndex(chainId))
	require.Nil(t, gotSecondBatch)

	// check that txs from first batch have been freed
	gotUnbatchedTx = nil
	input.GravityKeeper.IterateUnbatchedSendToExternals(ctx, chainId, func(tx *types.SendToExternal) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx = []*types.SendToExternal{
		types.NewSendToExternalTx(2, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 101, 3, 0, "#"),
		types.NewSendToExternalTx(3, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 102, 2, 0, "#"),
		types.NewSendToExternalTx(1, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 100, 2, 0, "#"),
		types.NewSendToExternalTx(4, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 103, 1, 0, "#"),
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)
}

// tests that batches work with large token amounts, mostly a duplicate of the above
// tests but using much bigger numbers
func TestBatchesFullCoins(t *testing.T) {
	input := CreateTestEnv(t)
	ctx := input.Context
	tokenInfos := input.GravityKeeper.GetTokenInfos(ctx).TokenInfos
	var (
		now                 = time.Now().UTC()
		mySender, _         = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = tokenInfos[0].ExternalTokenId
		tokenId             = tokenInfos[0].Id
		totalCoins, _       = sdk.NewIntFromString("1500000000000000000000") // 1,500 ETH worth
		oneEth, _           = sdk.NewIntFromString("1000000000000000000")

		allVouchers = sdk.NewCoins(
			types.NewSDKIntExternalToken(totalCoins, tokenId, myTokenContractAddr).HubCoin(testDenomResolver),
		)
	)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, fundAccount(ctx, input.BankKeeper, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	for _, v := range []uint64{20, 300, 25, 10} {
		vAsSDKInt := sdk.NewIntFromUint64(v)
		amount := types.NewSDKIntExternalToken(oneEth.Mul(vAsSDKInt), tokenId, myTokenContractAddr).HubCoin(testDenomResolver)
		fee := types.NewSDKIntExternalToken(oneEth.Mul(vAsSDKInt), tokenId, myTokenContractAddr).HubCoin(testDenomResolver)
		valCommission := types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr).HubCoin(testDenomResolver)

		_, err := input.GravityKeeper.createSendToExternal(ctx, chainId, mySender, myReceiver.Hex(), amount, fee, valCommission, "")
		require.NoError(t, err)
	}

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	firstBatch := input.GravityKeeper.BuildBatchTx(ctx, chainId, myTokenContractAddr, 2)

	// then batch is persisted
	gotFirstBatch := input.GravityKeeper.GetOutgoingTx(ctx, chainId, firstBatch.GetStoreIndex(chainId))
	require.NotNil(t, gotFirstBatch)

	expFirstBatch := &types.BatchTx{
		BatchNonce: 1,
		Transactions: []*types.SendToExternal{
			{
				Id:                2,
				Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(300)), tokenId, myTokenContractAddr),
				Sender:            mySender.String(),
				ExternalRecipient: myReceiver.Hex(),
				Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(300)), tokenId, myTokenContractAddr),
				ChainId:           chainId.String(),
				ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
			},
			{
				Id:                3,
				Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(25)), tokenId, myTokenContractAddr),
				Sender:            mySender.String(),
				ExternalRecipient: myReceiver.Hex(),
				Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(25)), tokenId, myTokenContractAddr),
				ChainId:           chainId.String(),
				ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
			},
		},
		ExternalTokenId: myTokenContractAddr,
		Height:          1234567,
		Sequence:        1,
	}
	assert.Equal(t, expFirstBatch, gotFirstBatch)

	// and verify remaining available Tx in the pool
	var gotUnbatchedTx []*types.SendToExternal
	input.GravityKeeper.IterateUnbatchedSendToExternals(ctx, chainId, func(tx *types.SendToExternal) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx := []*types.SendToExternal{
		{
			Id:                1,
			Sender:            mySender.String(),
			ExternalRecipient: myReceiver.Hex(),
			ChainId:           chainId.String(),
			Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(20)), tokenId, myTokenContractAddr),
			Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(20)), tokenId, myTokenContractAddr),
			TxHash:            "",
			ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
		},
		{
			Id:                4,
			Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(10)), tokenId, myTokenContractAddr),
			Sender:            mySender.String(),
			ExternalRecipient: myReceiver.Hex(),
			Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(10)), tokenId, myTokenContractAddr),
			ChainId:           chainId.String(),
			TxHash:            "",
			ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)

	// CREATE SECOND, MORE PROFITABLE BATCH
	// ====================================

	// add some more TX to the pool to create a more profitable batch
	for _, v := range []uint64{4, 5} {
		vAsSDKInt := sdk.NewIntFromUint64(v)
		amount := types.NewSDKIntExternalToken(oneEth.Mul(vAsSDKInt), tokenId, myTokenContractAddr).HubCoin(testDenomResolver)
		fee := types.NewSDKIntExternalToken(oneEth.Mul(vAsSDKInt), tokenId, myTokenContractAddr).HubCoin(testDenomResolver)
		valCommission := types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr).HubCoin(testDenomResolver)

		_, err := input.GravityKeeper.createSendToExternal(ctx, chainId, mySender, myReceiver.Hex(), amount, fee, valCommission, "")
		require.NoError(t, err)
	}

	// create the more profitable batch
	ctx = ctx.WithBlockTime(now)
	// tx batch size is 2, so that some of them stay behind
	secondBatch := input.GravityKeeper.BuildBatchTx(ctx, chainId, myTokenContractAddr, 2)

	// check that the more profitable batch has the right txs in it
	expSecondBatch := &types.BatchTx{
		BatchNonce: 2,
		Transactions: []*types.SendToExternal{
			{
				Id:                1,
				Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(20)), tokenId, myTokenContractAddr),
				Sender:            mySender.String(),
				ExternalRecipient: myReceiver.Hex(),
				Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(20)), tokenId, myTokenContractAddr),
				ChainId:           chainId.String(),
				TxHash:            "",
				ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
			},
			{
				Id:                4,
				Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(10)), tokenId, myTokenContractAddr),
				Sender:            mySender.String(),
				ExternalRecipient: myReceiver.Hex(),
				Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(10)), tokenId, myTokenContractAddr),
				ChainId:           chainId.String(),
				TxHash:            "",
				ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
			},
		},
		ExternalTokenId: myTokenContractAddr,
		Height:          1234567,
		Sequence:        2,
	}

	assert.Equal(t, expSecondBatch, secondBatch)

	// EXECUTE THE MORE PROFITABLE BATCH
	// =================================

	// Execute the batch
	input.GravityKeeper.batchTxExecuted(ctx, chainId, secondBatch.ExternalTokenId, secondBatch.BatchNonce, "", sdk.NewInt(0), "")

	// check batch has been deleted
	gotSecondBatch := input.GravityKeeper.GetOutgoingTx(ctx, chainId, secondBatch.GetStoreIndex(chainId))
	require.Nil(t, gotSecondBatch)

	// check that txs from first batch have been freed
	gotUnbatchedTx = nil
	input.GravityKeeper.IterateUnbatchedSendToExternals(ctx, chainId, func(tx *types.SendToExternal) bool {
		gotUnbatchedTx = append(gotUnbatchedTx, tx)
		return false
	})
	expUnbatchedTx = []*types.SendToExternal{
		{
			Id:                2,
			Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(300)), tokenId, myTokenContractAddr),
			Sender:            mySender.String(),
			ExternalRecipient: myReceiver.Hex(),
			Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(300)), tokenId, myTokenContractAddr),
			ChainId:           chainId.String(),
			TxHash:            "",
			ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
		},
		{
			Id:                3,
			Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(25)), tokenId, myTokenContractAddr),
			Sender:            mySender.String(),
			ExternalRecipient: myReceiver.Hex(),
			Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(25)), tokenId, myTokenContractAddr),
			ChainId:           chainId.String(),
			TxHash:            "",
			ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
		},
		{
			Id:                6,
			Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(5)), tokenId, myTokenContractAddr),
			Sender:            mySender.String(),
			ExternalRecipient: myReceiver.Hex(),
			Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(5)), tokenId, myTokenContractAddr),
			ChainId:           chainId.String(),
			TxHash:            "",
			ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
		},
		{
			Id:                5,
			Fee:               types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(4)), tokenId, myTokenContractAddr),
			Sender:            mySender.String(),
			ExternalRecipient: myReceiver.Hex(),
			Token:             types.NewSDKIntExternalToken(oneEth.Mul(sdk.NewIntFromUint64(4)), tokenId, myTokenContractAddr),
			ChainId:           chainId.String(),
			TxHash:            "",
			ValCommission:     types.NewSDKIntExternalToken(sdk.NewInt(0), tokenId, myTokenContractAddr),
		},
	}
	assert.Equal(t, expUnbatchedTx, gotUnbatchedTx)
}

func TestPoolTxRefund(t *testing.T) {
	input := CreateTestEnv(t)
	ctx := input.Context
	tokenInfos := input.GravityKeeper.GetTokenInfos(ctx).TokenInfos
	var (
		now                 = time.Now().UTC()
		mySender, _         = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = tokenInfos[0].ExternalTokenId
		tokenId             = tokenInfos[0].Id
		allVouchers         = sdk.NewCoins(
			types.NewExternalToken(414, tokenId, myTokenContractAddr).HubCoin(testDenomResolver),
		)
		myDenom = types.NewExternalToken(1, tokenId, myTokenContractAddr).HubCoin(testDenomResolver).Denom
	)

	// mint some voucher first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, fundAccount(ctx, input.BankKeeper, mySender, allVouchers))

	// CREATE FIRST BATCH
	// ==================

	// add some TX to the pool
	// for i, v := range []uint64{2, 3, 2, 1} {
	// 	amount := types.NewExternalToken(uint64(i+100), myTokenContractAddr).HubCoin()
	// 	fee := types.NewExternalToken(v, myTokenContractAddr).HubCoin()
	// 	_, err := input.GravityKeeper.CreateSendToEthereum(ctx, mySender, myReceiver, amount, fee)
	// 	require.NoError(t, err)
	// }
	input.AddSendToEthTxsToPool(t, ctx, chainId, tokenId, myTokenContractAddr, mySender, myReceiver, 2, 3, 2, 1)

	// when
	ctx = ctx.WithBlockTime(now)

	// tx batch size is 2, so that some of them stay behind
	input.GravityKeeper.BuildBatchTx(ctx, chainId, myTokenContractAddr, 2)

	// try to refund a tx that's in a batch
	err := input.GravityKeeper.cancelSendToExternal(ctx, chainId, 2, mySender.String())
	require.Error(t, err)

	// try to refund a tx that's in the pool
	err = input.GravityKeeper.cancelSendToExternal(ctx, chainId, 4, mySender.String())
	require.NoError(t, err)

	// make sure refund was issued
	balances := input.BankKeeper.GetAllBalances(ctx, mySender)
	require.Equal(t, sdk.NewInt(104), balances.AmountOf(myDenom))
}
