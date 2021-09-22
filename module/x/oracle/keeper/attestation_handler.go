package keeper

import (
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
		votes := att.GetVotes()
		holdersVotes := map[string]sdk.Int{}

		powers := a.keeper.GetNormalizedValPowers(ctx)

		for _, valaddr := range votes {
			validator, _ := sdk.ValAddressFromBech32(valaddr)
			power := sdk.NewDec(int64(powers[valaddr])).QuoInt64(math.MaxUint16)

			holdersClaim := a.keeper.GetHoldersClaim(ctx, sdk.AccAddress(validator).String(), claim.Epoch).(*types.GenericClaim).GetHoldersClaim()
			for _, item := range holdersClaim.GetHolders().List {
				if _, has := holdersVotes[item.Address]; !has {
					holdersVotes[item.Address] = sdk.NewInt(0)
				}

				holdersVotes[item.Address] = holdersVotes[item.Address].Add(item.Value.ToDec().Mul(power).TruncateInt())
			}
		}

		var addresses []string
		for address := range holdersVotes {
			addresses = append(addresses, address)
		}
		sort.Strings(addresses)

		holders := types.Holders{}
		for _, address := range addresses {
			holders.List = append(holders.List, &types.Holder{
				Address: address,
				Value:   holdersVotes[address],
			})
		}

		a.keeper.storeHolders(ctx, &holders)

	default:
		return sdkerrors.Wrapf(types.ErrInvalid, "event type: %s", claim.GetType())
	}
	return nil
}
