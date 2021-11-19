package keeper

import (
	"context"

	"github.com/MinterTeam/mhub2/module/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) HoldersClaim(c context.Context, msg *types.MsgHoldersClaim) (*types.MsgHoldersClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	orch, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	sval := k.StakingKeeper.Validator(ctx, sdk.ValAddress(orch))
	if sval == nil {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	if k.GetCurrentEpoch(ctx) != msg.GetEpoch() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "wrong epoch")
	}

	// Add the claim to the store
	att, err := k.AddClaim(ctx, msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create attestation")
	}

	// Emit the handle message event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			// TODO: maybe return something better here? is this the right string representation?
			sdk.NewAttribute(types.AttributeKeyAttestationID, string(types.GetAttestationKey(att.Epoch, msg))),
		),
	)

	return &types.MsgHoldersClaimResponse{}, nil
}

func (k msgServer) PriceClaim(c context.Context, msg *types.MsgPriceClaim) (*types.MsgPriceClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	orch, _ := sdk.AccAddressFromBech32(msg.Orchestrator)
	sval := k.StakingKeeper.Validator(ctx, sdk.ValAddress(orch))
	if sval == nil {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "validator")
	}

	if k.GetCurrentEpoch(ctx) != msg.GetEpoch() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "wrong epoch")
	}

	tokenInfos := k.Mhub2keeper.GetTokenInfos(ctx)
	requiredPrices := []string{"eth", "eth/gas"}
	for _, coin := range tokenInfos.TokenInfos {
		requiredPrices = append(requiredPrices, coin.Denom)
	}

	for _, requiredPrice := range requiredPrices {
		found := false
		for _, price := range msg.GetPrices().GetList() {
			if price.GetName() == requiredPrice && price.Value.IsPositive() {
				found = true
				break
			}
		}

		if !found {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "required price not found or malformed: %s", requiredPrice)
		}
	}

	// Add the claim to the store
	att, err := k.AddClaim(ctx, msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create attestation")
	}

	// Emit the handle message event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			// TODO: maybe return something better here? is this the right string representation?
			sdk.NewAttribute(types.AttributeKeyAttestationID, string(types.GetAttestationKey(att.Epoch, msg))),
		),
	)

	return &types.MsgPriceClaimResponse{}, nil
}
