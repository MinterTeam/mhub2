package oracle

import (
	"github.com/MinterTeam/mhub2/module/x/oracle/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	// process claims
	if ctx.BlockHeight()%5 == 0 {
		k.ProcessCurrentEpoch(ctx)
	}

	if epoch := k.GetCurrentEpoch(ctx); epoch > 10 {
		k.DeleteOldAttestations(ctx, epoch-10)
	}
}
