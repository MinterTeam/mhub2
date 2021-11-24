package keeper

import (
	"math/big"
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestEthereumEventProcessor_DetectMaliciousSupply(t *testing.T) {
	input := CreateTestEnv(t)
	eep := ExternalEventProcessor{keeper: input.Mhub2Keeper, bankKeeper: input.BankKeeper}

	// set supply to maximum value
	var testBigInt big.Int
	testBigInt.SetBit(new(big.Int), 256, 1).Sub(&testBigInt, big.NewInt(1))
	bigCoinAmount := sdktypes.NewIntFromBigInt(&testBigInt)

	err := eep.DetectMaliciousSupply(input.Context, "stake", bigCoinAmount)
	require.Error(t, err, "didn't error out on too much added supply")
}
