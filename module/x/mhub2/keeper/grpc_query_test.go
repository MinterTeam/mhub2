package keeper

import (
	"testing"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/bytes"
)

func TestKeeper_Params(t *testing.T) {
	env := CreateTestEnv(t)
	ctx := sdk.WrapSDKContext(env.Context)
	gk := env.Mhub2Keeper

	req := &types.ParamsRequest{}
	res, err := gk.Params(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestKeeper_LatestSignerSetTx(t *testing.T) {
	t.Run("read before there's anything in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper

		req := &types.LatestSignerSetTxRequest{chainId.String()}
		res, err := gk.LatestSignerSetTx(sdk.WrapSDKContext(ctx), req)
		require.Error(t, err)
		require.Nil(t, res)
	})
	t.Run("read after there's something in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper
		{ // setup
			sstx := gk.CreateSignerSetTx(env.Context, chainId)
			require.NotNil(t, sstx)
		}
		{ // validate
			req := &types.LatestSignerSetTxRequest{chainId.String()}
			res, err := gk.LatestSignerSetTx(sdk.WrapSDKContext(ctx), req)
			require.NoError(t, err)
			require.NotNil(t, res)
		}
	})
}

func TestKeeper_SignerSetTx(t *testing.T) {
	t.Run("read after there's something in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper

		var signerSetNonce uint64
		{ // setup
			sstx := gk.CreateSignerSetTx(env.Context, chainId)
			require.NotNil(t, sstx)
			signerSetNonce = sstx.Nonce
		}
		{ // validate
			req := &types.SignerSetTxRequest{SignerSetNonce: signerSetNonce, ChainId: chainId.String()}
			res, err := gk.SignerSetTx(sdk.WrapSDKContext(ctx), req)
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, res.SignerSet)
		}
	})
}

func TestKeeper_BatchTx(t *testing.T) {
	t.Run("read after there's something in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper

		const (
			batchNonce    = 55
			tokenContract = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
		)

		{ // setup
			gk.SetOutgoingTx(ctx, chainId, &types.BatchTx{
				BatchNonce:      batchNonce,
				Timeout:         1000,
				Transactions:    nil,
				ExternalTokenId: tokenContract,
				Height:          100,
			})
		}
		{ // validate
			req := &types.BatchTxRequest{
				BatchNonce:      batchNonce,
				ExternalTokenId: tokenContract,
				ChainId:         chainId.String(),
			}

			res, err := gk.BatchTx(sdk.WrapSDKContext(ctx), req)
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, res.Batch)
		}
	})
}

func TestKeeper_ContractCallTx(t *testing.T) {
	t.Run("read after there's something in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper

		const (
			invalidationNonce = 100
			invalidationScope = "an-invalidation-scope"
		)

		{ // setup
			gk.SetOutgoingTx(ctx, chainId, &types.ContractCallTx{
				InvalidationNonce: invalidationNonce,
				InvalidationScope: bytes.HexBytes(invalidationScope),
			})
		}
		{ // validate
			req := &types.ContractCallTxRequest{
				InvalidationNonce: invalidationNonce,
				InvalidationScope: bytes.HexBytes(invalidationScope),
				ChainId:           chainId.String(),
			}

			res, err := gk.ContractCallTx(sdk.WrapSDKContext(ctx), req)
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, res.LogicCall)
		}
	})
}

func TestKeeper_SignerSetTxs(t *testing.T) {
	t.Run("read after there's something in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper

		{ // setup
			require.NotNil(t, gk.CreateSignerSetTx(env.Context, chainId))
			require.NotNil(t, gk.CreateSignerSetTx(env.Context, chainId))
		}
		{ // validate
			req := &types.SignerSetTxsRequest{ChainId: chainId.String()}
			res, err := gk.SignerSetTxs(sdk.WrapSDKContext(ctx), req)
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Len(t, res.SignerSets, 2)
		}
	})
}

func TestKeeper_BatchTxs(t *testing.T) {
	t.Run("read after there's something in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper

		{ // setup
			gk.SetOutgoingTx(ctx, chainId, &types.BatchTx{
				BatchNonce:      1000,
				Timeout:         1000,
				Transactions:    nil,
				ExternalTokenId: "1",
				Height:          1000,
			})
			gk.SetOutgoingTx(ctx, chainId, &types.BatchTx{
				BatchNonce:      1001,
				Timeout:         1000,
				Transactions:    nil,
				ExternalTokenId: "1",
				Height:          1001,
			})
		}
		{ // validate
			req := &types.BatchTxsRequest{ChainId: chainId.String()}
			got, err := gk.BatchTxs(sdk.WrapSDKContext(ctx), req)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got.Batches, 2)
		}
	})
}

func TestKeeper_ContractCallTxs(t *testing.T) {
	t.Run("read after there's something in state", func(t *testing.T) {
		env := CreateTestEnv(t)
		ctx := env.Context
		gk := env.Mhub2Keeper

		{ // setup
			gk.SetOutgoingTx(ctx, chainId, &types.ContractCallTx{
				InvalidationNonce: 5,
				InvalidationScope: []byte("an-invalidation-scope"),
				// TODO
			})
			gk.SetOutgoingTx(ctx, chainId, &types.ContractCallTx{
				InvalidationNonce: 6,
				InvalidationScope: []byte("an-invalidation-scope"),
			})
		}
		{ // validate
			req := &types.ContractCallTxsRequest{ChainId: chainId.String()}
			got, err := gk.ContractCallTxs(sdk.WrapSDKContext(ctx), req)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got.Calls, 2)
		}
	})
}

// TODO ensure coverage for:
// ContractCallTx(context.Context, *ContractCallTxRequest) (*ContractCallTxResponse, error)
// ContractCallTxs(context.Context, *ContractCallTxsRequest) (*ContractCallTxsResponse, error)

// SignerSetTxConfirmations(context.Context, *SignerSetTxConfirmationsRequest) (*SignerSetTxConfirmationsResponse, error)
// BatchTxConfirmations(context.Context, *BatchTxConfirmationsRequest) (*BatchTxConfirmationsResponse, error)
// ContractCallTxConfirmations(context.Context, *ContractCallTxConfirmationsRequest) (*ContractCallTxConfirmationsResponse, error)

// UnsignedSignerSetTxs(context.Context, *UnsignedSignerSetTxsRequest) (*UnsignedSignerSetTxsResponse, error)
// UnsignedBatchTxs(context.Context, *UnsignedBatchTxsRequest) (*UnsignedBatchTxsResponse, error)
// UnsignedContractCallTxs(context.Context, *UnsignedContractCallTxsRequest) (*UnsignedContractCallTxsResponse, error)

// BatchTxFees(context.Context, *BatchTxFeesRequest) (*BatchTxFeesResponse, error)
// ERC20ToDenom(context.Context, *ERC20ToDenomRequest) (*ERC20ToDenomResponse, error)
// DenomToERC20(context.Context, *DenomToERC20Request) (*DenomToERC20Response, error)
// BatchedSendToEthereums(context.Context, *BatchedSendToEthereumsRequest) (*BatchedSendToEthereumsResponse, error)
// UnbatchedSendToEthereums(context.Context, *UnbatchedSendToEthereumsRequest) (*UnbatchedSendToEthereumsResponse, error)
// DelegateKeysByValidator(context.Context, *DelegateKeysByValidatorRequest) (*DelegateKeysByValidatorResponse, error)
// DelegateKeysByEthereumSigner(context.Context, *DelegateKeysByEthereumSignerRequest) (*DelegateKeysByEthereumSignerResponse, error)
// DelegateKeysByOrchestrator(context.Context, *DelegateKeysByOrchestratorRequest) (*DelegateKeysByOrchestratorResponse, error)
