package mhub2_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/MinterTeam/mhub2/module/x/mhub2"
	"github.com/MinterTeam/mhub2/module/x/mhub2/keeper"
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

var chainId = types.ChainID("ethereum")

var testDenomResolver = func(id uint64) (string, error) {
	return "hub", nil
}

func TestSignerSetTxCreationIfNotAvailable(t *testing.T) {
	input, ctx := keeper.SetupFiveValChain(t)
	gravityKeeper := input.GravityKeeper

	// BeginBlocker should set a new validator set if not available
	mhub2.BeginBlocker(ctx, gravityKeeper)
	otx := gravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeSignerSetTxKey(chainId, 1))
	require.NotNil(t, otx)
	_, ok := otx.(*types.SignerSetTx)
	require.True(t, ok)
	require.True(t, len(gravityKeeper.GetSignerSetTxs(ctx, chainId)) == 1)
}

func TestSignerSetTxCreationUponUnbonding(t *testing.T) {
	input, ctx := keeper.SetupFiveValChain(t)
	gravityKeeper := input.GravityKeeper
	gravityKeeper.CreateSignerSetTx(ctx, chainId)

	input.Context = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// begin unbonding
	sh := staking.NewHandler(input.StakingKeeper)
	undelegateMsg := keeper.NewTestMsgUnDelegateValidator(keeper.ValAddrs[0], keeper.StakingAmount)
	sh(input.Context, undelegateMsg)

	// Run the staking endblocker to ensure signer set tx is set in state
	staking.EndBlocker(input.Context, input.StakingKeeper)
	mhub2.BeginBlocker(input.Context, gravityKeeper)

	require.EqualValues(t, 2, gravityKeeper.GetLatestSignerSetTxNonce(ctx, chainId))
}

func TestSignerSetTxSlashing_SignerSetTxCreated_Before_ValidatorBonded(t *testing.T) {
	//	Don't slash validators if signer set tx is created before he is bonded.

	input, ctx := keeper.SetupFiveValChain(t)
	pk := input.GravityKeeper
	params := input.GravityKeeper.GetParams(ctx)

	signerSet := pk.CreateSignerSetTx(ctx, chainId)
	height := uint64(ctx.BlockHeight()) - (params.SignedSignerSetTxsWindow + 1)
	signerSet.Height = height
	pk.SetOutgoingTx(ctx, chainId, signerSet)

	mhub2.EndBlocker(ctx, pk)

	// ensure that the  validator who is bonded after signer set tx is created is not slashed
	val := input.StakingKeeper.Validator(ctx, keeper.ValAddrs[0])
	require.False(t, val.IsJailed())
}

func TestSignerSetTxSlashing_SignerSetTxCreated_After_ValidatorBonded(t *testing.T) {
	return // todo: enable when slashing will be enabled
	//	Slashing Conditions for Bonded Validator

	input, ctx := keeper.SetupFiveValChain(t)
	pk := input.GravityKeeper
	params := input.GravityKeeper.GetParams(ctx)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(params.SignedSignerSetTxsWindow) + 2)
	signerSet := pk.CreateSignerSetTx(ctx, chainId)
	height := uint64(ctx.BlockHeight()) - (params.SignedSignerSetTxsWindow + 1)
	signerSet.Height = height
	pk.SetOutgoingTx(ctx, chainId, signerSet)

	for i, val := range keeper.ValAddrs {
		if i == 0 {
			continue
		}
		pk.SetExternalSignature(ctx, chainId, &types.SignerSetTxConfirmation{signerSet.Nonce, keeper.AccAddrs[i].String(), []byte("dummysig")}, val)
	}

	mhub2.EndBlocker(ctx, pk)

	// ensure that the  validator who is bonded before signer set tx is created is slashed
	val := input.StakingKeeper.Validator(ctx, keeper.ValAddrs[0])
	require.True(t, val.IsJailed())

	// ensure that the  validator who attested the signer set tx is not slashed.
	val = input.StakingKeeper.Validator(ctx, keeper.ValAddrs[1])
	require.False(t, val.IsJailed())

}

func TestSignerSetTxSlashing_UnbondingValidator_UnbondWindow_NotExpired(t *testing.T) {
	//	Slashing Conditions for Unbonding Validator

	//  Create 5 validators
	input, ctx := keeper.SetupFiveValChain(t)
	// val := input.StakingKeeper.Validator(ctx, keeper.ValAddrs[0])
	// fmt.Println("val1  tokens", val.GetTokens().ToDec())

	gravityKeeper := input.GravityKeeper
	params := input.GravityKeeper.GetParams(ctx)

	// Define slashing variables
	validatorStartHeight := ctx.BlockHeight()                                                             // 0
	signerSetTxHeight := validatorStartHeight + 1                                                         // 1
	valUnbondingHeight := signerSetTxHeight + 1                                                           // 2
	signerSetTxSlashedAt := signerSetTxHeight + int64(params.SignedSignerSetTxsWindow)                    // 11
	validatorUnbondingWindowExpiry := valUnbondingHeight + int64(params.UnbondSlashingSignerSetTxsWindow) // 17
	currentBlockHeight := signerSetTxSlashedAt + 1                                                        // 12

	require.True(t, signerSetTxSlashedAt < currentBlockHeight)
	require.True(t, signerSetTxHeight < validatorUnbondingWindowExpiry)

	// Create signer set tx request
	ctx = ctx.WithBlockHeight(signerSetTxHeight)
	vs := gravityKeeper.CreateSignerSetTx(ctx, chainId)
	vs.Height = uint64(signerSetTxHeight)
	vs.Nonce = uint64(signerSetTxHeight)
	gravityKeeper.SetOutgoingTx(ctx, chainId, vs)

	// Start Unbonding validators
	// Validator-1  Unbond slash window is not expired. if not attested, slash
	// Validator-2  Unbond slash window is not expired. if attested, don't slash
	input.Context = ctx.WithBlockHeight(valUnbondingHeight)
	sh := staking.NewHandler(input.StakingKeeper)
	undelegateMsg1 := keeper.NewTestMsgUnDelegateValidator(keeper.ValAddrs[0], keeper.StakingAmount)
	sh(input.Context, undelegateMsg1)
	undelegateMsg2 := keeper.NewTestMsgUnDelegateValidator(keeper.ValAddrs[1], keeper.StakingAmount)
	sh(input.Context, undelegateMsg2)

	for i, val := range keeper.ValAddrs {
		if i == 0 {
			// don't sign with first validator
			continue
		}
		gravityKeeper.SetExternalSignature(ctx, chainId, &types.SignerSetTxConfirmation{vs.Nonce, keeper.EthAddrs[i].Hex(), []byte("dummySig")}, val)
	}
	staking.EndBlocker(input.Context, input.StakingKeeper)

	ctx = ctx.WithBlockHeight(currentBlockHeight)
	mhub2.EndBlocker(ctx, gravityKeeper)

	// Assertions
	val1 := input.StakingKeeper.Validator(ctx, keeper.ValAddrs[0])
	require.True(t, val1.IsJailed())
	fmt.Println("val1  tokens", val1.GetTokens().ToDec())
	// check if tokens are slashed for val1.

	val2 := input.StakingKeeper.Validator(ctx, keeper.ValAddrs[1])
	require.True(t, val2.IsJailed())
	fmt.Println("val2  tokens", val2.GetTokens().ToDec())
	// check if tokens shouldn't be slashed for val2.
}

func TestBatchSlashing(t *testing.T) {
	return // todo: enable when slashing will be enabled

	input, ctx := keeper.SetupFiveValChain(t)
	gravityKeeper := input.GravityKeeper
	params := gravityKeeper.GetParams(ctx)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(params.SignedBatchesWindow) + 2)

	// First store a batch
	batch := &types.BatchTx{
		BatchNonce:      1,
		Transactions:    []*types.SendToExternal{},
		ExternalTokenId: keeper.TokenContractAddrs[0],
		Height:          uint64(ctx.BlockHeight() - int64(params.SignedBatchesWindow+1)),
	}
	gravityKeeper.SetOutgoingTx(ctx, chainId, batch)

	for i, val := range keeper.ValAddrs {
		if i == 0 {
			// don't sign with first validator
			continue
		}
		if i == 1 {
			// don't sign with 2nd validator. set val bond height > batch block height
			validator := input.StakingKeeper.Validator(ctx, keeper.ValAddrs[i])
			valConsAddr, _ := validator.GetConsAddr()
			valSigningInfo := slashingtypes.ValidatorSigningInfo{StartHeight: int64(batch.Height + 1)}
			input.SlashingKeeper.SetValidatorSigningInfo(ctx, valConsAddr, valSigningInfo)
			continue
		}
		gravityKeeper.SetExternalSignature(ctx, chainId, &types.BatchTxConfirmation{
			BatchNonce:      batch.BatchNonce,
			ExternalTokenId: keeper.TokenContractAddrs[0],
			ExternalSigner:  keeper.EthAddrs[i].String(),
			Signature:       []byte("dummysig"),
		}, val)
	}

	mhub2.EndBlocker(ctx, gravityKeeper)

	// ensure that the  validator is jailed and slashed
	require.True(t, input.StakingKeeper.Validator(ctx, keeper.ValAddrs[0]).IsJailed())

	// ensure that the 2nd  validator is not jailed and slashed
	require.False(t, input.StakingKeeper.Validator(ctx, keeper.ValAddrs[1]).IsJailed())

	// Ensure that the last slashed signer set tx nonce is set properly
	require.Equal(t, input.GravityKeeper.GetLastSlashedOutgoingTxBlockHeight(ctx, chainId), batch.Height)
}

func TestSignerSetTxEmission(t *testing.T) {
	input, ctx := keeper.SetupFiveValChain(t)
	gravityKeeper := input.GravityKeeper

	// Store a validator set with a power change as the most recent validator set
	sstx := gravityKeeper.CreateSignerSetTx(ctx, chainId)
	delta := float64(types.ExternalSigners(sstx.Signers).TotalPower()) * 0.05
	sstx.Signers[0].Power = uint64(float64(sstx.Signers[0].Power) - delta/2)
	sstx.Signers[1].Power = uint64(float64(sstx.Signers[1].Power) + delta/2)
	gravityKeeper.SetOutgoingTx(ctx, chainId, sstx)

	// BeginBlocker should set a new validator set
	mhub2.BeginBlocker(ctx, gravityKeeper)
	require.NotNil(t, gravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeSignerSetTxKey(chainId, 2)))
	require.EqualValues(t, 2, len(gravityKeeper.GetSignerSetTxs(ctx, chainId)))
}

func TestSignerSetTxSetting(t *testing.T) {
	input, ctx := keeper.SetupFiveValChain(t)
	gk := input.GravityKeeper
	gk.CreateSignerSetTx(ctx, chainId)
	require.EqualValues(t, 1, len(gk.GetSignerSetTxs(ctx, chainId)))
}

/// Test batch timeout
func TestBatchTxTimeout(t *testing.T) {
	input, ctx := keeper.SetupFiveValChain(t)
	gravityKeeper := input.GravityKeeper
	params := gravityKeeper.GetParams(ctx)
	tokenInfos := input.GravityKeeper.GetTokenInfos(ctx).TokenInfos
	var (
		now             = time.Now().UTC()
		mySender, _     = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver      = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		tokenId         = tokenInfos[0].Id
		externalTokenId = tokenInfos[0].ExternalTokenId
		allVouchers     = sdk.NewCoins(types.NewExternalToken(99999, tokenId, externalTokenId).HubCoin(testDenomResolver))
	)

	require.Greater(t, params.AverageBlockTime, uint64(0))
	require.Greater(t, params.AverageEthereumBlockTime, uint64(0))

	// mint some vouchers first
	require.NoError(t, input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers))
	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, fundAccount(ctx, input.BankKeeper, mySender, allVouchers))

	// add some TX to the pool
	input.AddSendToEthTxsToPool(t, ctx, chainId, tokenId, externalTokenId, mySender, myReceiver, 2, 3, 2, 1, 5, 6)

	// when
	ctx = ctx.WithBlockTime(now).WithBlockHeight(250)

	// check that we can make a batch without first setting an ethereum block height
	b1 := gravityKeeper.BuildBatchTx(ctx, chainId, externalTokenId, 2)
	require.Equal(t, b1.Timeout, uint64(0))

	gravityKeeper.SetLastObservedExternalBlockHeight(ctx, chainId, 500)

	b2 := gravityKeeper.BuildBatchTx(ctx, chainId, externalTokenId, 2)
	// this is exactly block 500 plus twelve hours
	require.Equal(t, b2.Timeout, uint64(504))

	// make sure the batches got stored in the first place
	gotFirstBatch := input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b1.ExternalTokenId, b1.BatchNonce))
	require.NotNil(t, gotFirstBatch)
	gotSecondBatch := input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b2.ExternalTokenId, b2.BatchNonce))
	require.NotNil(t, gotSecondBatch)

	// when, way into the future
	ctx = ctx.WithBlockTime(now).WithBlockHeight(9)

	b3 := gravityKeeper.BuildBatchTx(ctx, chainId, externalTokenId, 2)

	mhub2.BeginBlocker(ctx, gravityKeeper)

	// this had a timeout of zero should be deleted.
	gotFirstBatch = input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b1.ExternalTokenId, b1.BatchNonce))
	require.Nil(t, gotFirstBatch)
	// make sure the end blocker does not delete these, as the block height has not officially
	// been updated by a relay event
	gotSecondBatch = input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b2.ExternalTokenId, b2.BatchNonce))
	require.NotNil(t, gotSecondBatch)
	gotThirdBatch := input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b3.ExternalTokenId, b3.BatchNonce))
	require.NotNil(t, gotThirdBatch)

	gravityKeeper.SetLastObservedExternalBlockHeight(ctx, chainId, 5000)
	mhub2.BeginBlocker(ctx, gravityKeeper)

	// make sure the end blocker does delete these, as we've got a new Ethereum block height
	gotFirstBatch = input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b1.ExternalTokenId, b1.BatchNonce))
	require.Nil(t, gotFirstBatch)
	gotSecondBatch = input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b2.ExternalTokenId, b2.BatchNonce))
	require.Nil(t, gotSecondBatch)
	gotThirdBatch = input.GravityKeeper.GetOutgoingTx(ctx, chainId, types.MakeBatchTxKey(chainId, b3.ExternalTokenId, b3.BatchNonce))
	require.NotNil(t, gotThirdBatch)
}

func fundAccount(ctx sdk.Context, bankKeeper types.BankKeeper, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, types.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, amounts)
}
