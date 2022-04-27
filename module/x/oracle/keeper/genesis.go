package keeper

import (
	"github.com/MinterTeam/mhub2/module/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis starts a chain from a genesis state
func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) {
	k.SetParams(ctx, *data.Params)
	k.setCurrentEpoch(ctx, 1)

	if data.Prices != nil {
		k.storePrices(ctx, data.Prices)
	}

	if data.Holders != nil {
		k.storeHolders(ctx, data.Holders)
	}
}

// ExportGenesis exports all the state needed to restart the chain
// from the current state of the chain
func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	var (
		p            = k.GetParams(ctx)
		attmap       = k.GetAttestationMapping(ctx)
		prices       = k.GetPrices(ctx)
		holders      = k.GetHolders(ctx)
		attestations = []types.Attestation{}
	)

	// export attestations from state
	for _, atts := range attmap {
		// TODO: set height = 0?
		attestations = append(attestations, atts...)
	}

	return types.GenesisState{
		Params:  &p,
		Prices:  prices,
		Holders: holders,
	}
}
