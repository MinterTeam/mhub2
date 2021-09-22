package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

func TestAddToOutgoingPool(t *testing.T) {
	input := CreateTestEnv(t)
	ctx := input.Context
	tokenInfos := input.GravityKeeper.GetTokenInfos(ctx).TokenInfos
	var (
		mySender, _         = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
		myReceiver          = common.HexToAddress("0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7")
		myTokenContractAddr = common.HexToAddress(tokenInfos[0].ExternalTokenId)
		tokenId             = tokenInfos[0].Id
	)
	// mint some voucher first
	allVouchers := sdk.Coins{types.NewExternalToken(99999, tokenId, myTokenContractAddr.String()).HubCoin(testDenomResolver)}
	err := input.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	input.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	require.NoError(t, fundAccount(ctx, input.BankKeeper, mySender, allVouchers))

	// when
	input.AddSendToEthTxsToPool(t, ctx, chainId, tokenId, myTokenContractAddr.String(), mySender, myReceiver, 2, 3, 2, 1)

	// then
	var got []*types.SendToExternal
	input.GravityKeeper.IterateUnbatchedSendToExternals(ctx, chainId, func(tx *types.SendToExternal) bool {
		got = append(got, tx)
		return false
	})

	exp := []*types.SendToExternal{
		types.NewSendToExternalTx(2, chainId, tokenId, myTokenContractAddr.String(), mySender, myReceiver, 101, 3, 0, "#"),
		types.NewSendToExternalTx(3, chainId, tokenId, myTokenContractAddr.String(), mySender, myReceiver, 102, 2, 0, "#"),
		types.NewSendToExternalTx(1, chainId, tokenId, myTokenContractAddr.String(), mySender, myReceiver, 100, 2, 0, "#"),
		types.NewSendToExternalTx(4, chainId, tokenId, myTokenContractAddr.String(), mySender, myReceiver, 103, 1, 0, "#"),
	}

	require.Equal(t, exp, got)
	require.EqualValues(t, exp[0], got[0])
	require.EqualValues(t, exp[1], got[1])
	require.EqualValues(t, exp[2], got[2])
	require.EqualValues(t, exp[3], got[3])
	require.Len(t, got, 4)
}
