package keeper

import (
	"fmt"
	"math"
	"sort"

	"github.com/MinterTeam/mhub2/module/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AttestationHandler processes `observed` Attestations
type AttestationHandler struct {
	keeper        Keeper
	stakingKeeper types.StakingKeeper
}

// Handle is the entry point for Attestation processing.
func (a AttestationHandler) Handle(ctx sdk.Context, att types.Attestation, claim types.Claim) error {
	switch claim := claim.(type) {
	case *types.MsgPriceClaim:
		votes := att.GetVotes()
		pricesSum := map[string][]sdk.Dec{}

		powers := a.keeper.GetNormalizedValPowers(ctx)

		for _, valaddr := range votes {
			validator, _ := sdk.ValAddressFromBech32(valaddr)
			power := powers[valaddr]

			priceClaim := a.keeper.GetPriceClaim(ctx, sdk.AccAddress(validator).String(), claim.Epoch).(*types.GenericClaim).GetPriceClaim()
			prices := priceClaim.GetPrices()
			for _, item := range prices.List {
				for i := uint64(0); i < power; i++ {
					pricesSum[item.Name] = append(pricesSum[item.Name], item.Value)
				}
			}

			a.keeper.deletePriceClaim(ctx, sdk.AccAddress(validator).String(), att.Epoch)
		}

		var priceNames []string
		for name := range pricesSum {
			priceNames = append(priceNames, name)
		}
		sort.Strings(priceNames)

		prices := types.Prices{}
		for _, name := range priceNames {
			price := pricesSum[name]

			sort.Slice(price, func(i, j int) bool {
				return price[i].LT(price[j])
			})

			var calculatedPrice sdk.Dec
			if len(price)%2 == 0 {
				calculatedPrice = price[len(price)/2].Add(price[len(price)/2-1]).QuoInt64(2) // compute average
			} else {
				calculatedPrice = price[len(price)/2]
			}

			prices.List = append(prices.List, &types.Price{
				Name:  name,
				Value: calculatedPrice,
			})
		}

		a.keeper.storePrices(ctx, &prices)
	case *types.MsgHoldersClaim:
		holdersTally := map[string]uint64{}
		holdersVotes := map[string]*types.Holders{}

		powers := a.keeper.GetNormalizedValPowers(ctx)
		for _, valaddr := range att.GetVotes() {
			validator, _ := sdk.ValAddressFromBech32(valaddr)

			holdersClaim := a.keeper.GetHoldersClaim(ctx, sdk.AccAddress(validator).String(), claim.Epoch).(*types.GenericClaim).GetHoldersClaim()
			hash := fmt.Sprintf("%x", holdersClaim.StabilizedClaimHash())
			holdersVotes[hash] = holdersClaim.Holders
			holdersTally[hash] = holdersTally[hash] + powers[valaddr]
			a.keeper.deleteHoldersClaim(ctx, sdk.AccAddress(validator).String(), att.Epoch)
		}

		// todo: should we iterate this in sorted way?
		for hash, votes := range holdersTally {
			if votes > math.MaxUint16*2/3 {
				a.keeper.storeHolders(ctx, holdersVotes[hash])
				return nil
			}
		}

	default:
		return sdkerrors.Wrapf(types.ErrInvalid, "event type: %s", claim.GetType())
	}
	return nil
}
