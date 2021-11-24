package keeper

import (
	"context"

	"github.com/MinterTeam/mhub2/module/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const gweiInEth = 1e9

var _ types.QueryServer = Keeper{}

func (k Keeper) Holders(context context.Context, _ *types.QueryHoldersRequest) (*types.QueryHoldersResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	return &types.QueryHoldersResponse{Holders: k.GetHolders(ctx)}, nil
}

func (k Keeper) Prices(context context.Context, _ *types.QueryPricesRequest) (*types.QueryPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	return &types.QueryPricesResponse{Prices: k.GetPrices(ctx)}, nil
}

func (k Keeper) CurrentEpoch(context context.Context, _ *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	currentEpoch := types.Epoch{
		Nonce: k.GetCurrentEpoch(ctx),
	}

	att := k.GetAttestation(ctx, currentEpoch.Nonce, &types.MsgPriceClaim{})
	votes := att.GetVotes()
	for _, valaddr := range votes {
		validator, _ := sdk.ValAddressFromBech32(valaddr)
		oracle := sdk.AccAddress(validator).String()
		priceClaim := k.GetPriceClaim(ctx, oracle, currentEpoch.Nonce).(*types.GenericClaim).GetPriceClaim()
		holdersChaim := k.GetHoldersClaim(ctx, oracle, currentEpoch.Nonce).(*types.GenericClaim).GetHoldersClaim()
		currentEpoch.Votes = append(currentEpoch.Votes, &types.Vote{
			Oracle:       oracle,
			PriceClaim:   priceClaim,
			HoldersClaim: holdersChaim,
		})
	}

	return &types.QueryCurrentEpochResponse{Epoch: &currentEpoch}, nil
}

func (k Keeper) EthFee(context context.Context, _ *types.QueryEthFeeRequest) (*types.QueryEthFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	gasPrice, err := k.GetTokenPrice(ctx, "eth/gas")
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gas price")
	}

	ethPrice, err := k.GetTokenPrice(ctx, "eth")
	if err != nil {
		return nil, sdkerrors.Wrap(err, "eth price")
	}

	return &types.QueryEthFeeResponse{
		Min:  gasPrice.Mul(ethPrice).MulInt64(int64(k.GetMinSingleWithdrawGas(ctx))).QuoInt64(gweiInEth).TruncateInt(),
		Fast: gasPrice.Mul(ethPrice).MulInt64(int64(k.GetMinBatchGas(ctx))).QuoInt64(gweiInEth).TruncateInt(),
	}, nil
}

func (k Keeper) BscFee(context context.Context, _ *types.QueryBscFeeRequest) (*types.QueryBscFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(context)

	gasPrice, err := k.GetTokenPrice(ctx, "bsc/gas")
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gas price")
	}

	bnbPrice, err := k.GetTokenPrice(ctx, "bnb")
	if err != nil {
		return nil, sdkerrors.Wrap(err, "eth price")
	}

	return &types.QueryBscFeeResponse{
		Min:  gasPrice.Mul(bnbPrice).MulInt64(int64(k.GetMinSingleWithdrawGas(ctx))).QuoInt64(gweiInEth).TruncateInt(),
		Fast: gasPrice.Mul(bnbPrice).MulInt64(int64(k.GetMinBatchGas(ctx))).QuoInt64(gweiInEth).TruncateInt(),
	}, nil
}
