package keeper

import (
	"fmt"
	"math"
	"strings"

	"github.com/MinterTeam/mhub2/module/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	StakingKeeper types.StakingKeeper
	Mhub2keeper   types.Mhub2Keeper

	storeKey   sdk.StoreKey // Unexposed key to access store from sdk.Context
	paramSpace paramtypes.Subspace

	cdc        codec.BinaryCodec // The wire codec for binary encoding/decoding.
	bankKeeper types.BankKeeper

	AttestationHandler interface {
		Handle(sdk.Context, types.Attestation, types.Claim) error
	}
}

// NewKeeper returns a new instance of the peggy keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace, stakingKeeper types.StakingKeeper) Keeper {
	k := Keeper{
		cdc:           cdc,
		paramSpace:    paramSpace,
		storeKey:      storeKey,
		StakingKeeper: stakingKeeper,
	}
	k.AttestationHandler = AttestationHandler{
		keeper:        k,
		stakingKeeper: stakingKeeper,
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return k
}

func (k Keeper) SetMhub2Keeper(keeper types.Mhub2Keeper) Keeper {
	k.Mhub2keeper = keeper

	return k
}

func (k Keeper) GetHolders(ctx sdk.Context) *types.Holders {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.CurrentHoldersKey)

	if len(bytes) == 0 {
		return nil
	}

	var holders types.Holders
	k.cdc.MustUnmarshal(bytes, &holders)

	return &holders
}

func (k Keeper) GetPrices(ctx sdk.Context) *types.Prices {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.CurrentPricesKey)

	if len(bytes) == 0 {
		return nil
	}

	var prices types.Prices
	k.cdc.MustUnmarshal(bytes, &prices)

	return &prices
}

func (k Keeper) MustGetTokenPrice(ctx sdk.Context, denom string) sdk.Dec {
	price, err := k.GetTokenPrice(ctx, denom)
	if err != nil {
		panic(err)
	}

	return price
}

func (k Keeper) GetTokenPrice(ctx sdk.Context, denom string) (sdk.Dec, error) {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.CurrentPricesKey)

	if len(bytes) == 0 {
		return sdk.Dec{}, sdkerrors.ErrKeyNotFound
	}

	var prices types.Prices
	k.cdc.MustUnmarshal(bytes, &prices)

	for _, price := range prices.GetList() {
		if price.GetName() == denom {
			return price.Value, nil
		}
	}

	return sdk.Dec{}, sdkerrors.ErrKeyNotFound
}

func (k Keeper) GetCurrentEpoch(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.CurrentEpochKey)

	if len(bytes) == 0 {
		return 0
	}
	return types.UInt64FromBytes(bytes)
}

func (k Keeper) setCurrentEpoch(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.CurrentEpochKey, types.UInt64Bytes(nonce))
}

/////////////////////////////
//       PARAMETERS        //
/////////////////////////////

// GetParams returns the parameters from the store
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}

// SetParams sets the parameters in the store
func (k Keeper) SetParams(ctx sdk.Context, ps types.Params) {
	k.paramSpace.SetParamSet(ctx, &ps)
}

func (k Keeper) GetHolderValue(ctx sdk.Context, address string) sdk.Int {
	for _, item := range k.GetHolders(ctx).GetList() {
		if strings.ToLower(item.GetAddress()) == strings.ToLower(address) {
			return item.Value
		}
	}

	return sdk.NewInt(0)
}

// logger returns a module-specific logger.
func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ProcessCurrentEpoch(ctx sdk.Context) {
	currentEpoch := k.GetCurrentEpoch(ctx)
	k.setCurrentEpoch(ctx, currentEpoch+1)

	{
		claim := &types.MsgPriceClaim{
			Epoch: currentEpoch,
		}
		att := k.GetAttestation(ctx, currentEpoch, claim)
		if att != nil {
			k.tryAttestation(ctx, att, claim)
		}
	}

	{
		claim := &types.MsgHoldersClaim{
			Epoch: currentEpoch,
		}
		att := k.GetAttestation(ctx, currentEpoch, claim)
		if att != nil {
			k.tryAttestation(ctx, att, claim)
		}
	}
}

func (k Keeper) storePrices(ctx sdk.Context, prices *types.Prices) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.CurrentPricesKey, k.cdc.MustMarshal(prices))
}

func (k Keeper) storeHolders(ctx sdk.Context, holders *types.Holders) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.CurrentHoldersKey, k.cdc.MustMarshal(holders))
}

func (k Keeper) GetNormalizedValPowers(ctx sdk.Context) map[string]uint64 {
	validators := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	bridgeValidators := map[string]uint64{}
	var totalPower uint64

	for _, validator := range validators {
		validatorAddress := validator.GetOperator()

		p := uint64(k.StakingKeeper.GetLastValidatorPower(ctx, validatorAddress))
		totalPower += p

		bridgeValidators[validatorAddress.String()] = p
	}

	// normalize power values
	for address, power := range bridgeValidators {
		bridgeValidators[address] = sdk.NewUint(power).MulUint64(math.MaxUint16).QuoUint64(totalPower).Uint64()
	}

	return bridgeValidators
}

// prefixRange turns a prefix into a (start, end) range. The start is the given prefix value and
// the end is calculated by adding 1 bit to the start value. Nil is not allowed as prefix.
// 		Example: []byte{1, 3, 4} becomes []byte{1, 3, 5}
// 				 []byte{15, 42, 255, 255} becomes []byte{15, 43, 0, 0}
//
// In case of an overflow the end is set to nil.
//		Example: []byte{255, 255, 255, 255} becomes nil
// MARK finish-batches: this is where some crazy shit happens
func prefixRange(prefix []byte) ([]byte, []byte) {
	if prefix == nil {
		panic("nil key not allowed")
	}
	// special case: no prefix is whole range
	if len(prefix) == 0 {
		return nil, nil
	}

	// copy the prefix and update last byte
	end := make([]byte, len(prefix))
	copy(end, prefix)
	l := len(end) - 1
	end[l]++

	// wait, what if that overflowed?....
	for end[l] == 0 && l > 0 {
		l--
		end[l]++
	}

	// okay, funny guy, you gave us FFF, no end to this range...
	if l == 0 && end[0] == 0 {
		end = nil
	}
	return prefix, end
}
