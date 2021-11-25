package keeper

import (
	"bytes"
	"context"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) DiscountForHolder(c context.Context, request *types.DiscountForHolderRequest) (*types.DiscountForHolderResponse, error) {
	return &types.DiscountForHolderResponse{
		Discount: sdk.NewDec(1).Sub(k.GetCommissionForHolder(sdk.UnwrapSDKContext(c), []string{request.Address}, sdk.NewDec(1))),
	}, nil
}

func (k Keeper) TransactionStatus(ctx context.Context, request *types.TransactionStatusRequest) (*types.TransactionStatusResponse, error) {
	return &types.TransactionStatusResponse{Status: k.GetTxStatus(sdk.UnwrapSDKContext(ctx), request.TxHash)}, nil
}

func (k Keeper) TokenInfos(ctx context.Context, _ *types.TokenInfosRequest) (*types.TokenInfosResponse, error) {
	return &types.TokenInfosResponse{List: *k.GetTokenInfos(sdk.UnwrapSDKContext(ctx))}, nil
}

func (k Keeper) Params(c context.Context, _ *types.ParamsRequest) (*types.ParamsResponse, error) {
	params := k.GetParams(sdk.UnwrapSDKContext(c))
	return &types.ParamsResponse{Params: params}, nil
}

func (k Keeper) LastSubmittedExternalEvent(c context.Context, req *types.LastSubmittedExternalEventRequest) (*types.LastSubmittedExternalEventResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	valAddr, err := k.getSignerValidator(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	return &types.LastSubmittedExternalEventResponse{
		EventNonce: k.getLastEventNonceByValidator(ctx, types.ChainID(req.ChainId), valAddr),
	}, nil
}

func (k Keeper) BatchedSendToExternals(c context.Context, req *types.BatchedSendToExternalsRequest) (*types.BatchedSendToExternalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	res := &types.BatchedSendToExternalsResponse{
		ChainId: req.ChainId,
	}

	k.IterateOutgoingTxsByType(ctx, types.ChainID(req.ChainId), types.BatchTxPrefixByte, func(_ []byte, outgoing types.OutgoingTx) bool {
		batchTx := outgoing.(*types.BatchTx)
		for _, ste := range batchTx.Transactions {
			if req.SenderAddress == "" || ste.Sender == req.SenderAddress {
				res.SendToExternals = append(res.SendToExternals, ste)
			}
		}

		return false
	})

	return res, nil
}

func (k Keeper) UnbatchedSendToExternals(c context.Context, req *types.UnbatchedSendToExternalsRequest) (*types.UnbatchedSendToExternalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	res := &types.UnbatchedSendToExternalsResponse{
		ChainId: req.ChainId,
	}

	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), bytes.Join([][]byte{{types.SendToExternalKey}, types.ChainID(req.ChainId).Bytes()}, []byte{}))
	pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var ste types.SendToExternal
		k.cdc.MustUnmarshal(value, &ste)
		if req.SenderAddress == "" || ste.Sender == req.SenderAddress {
			res.SendToExternals = append(res.SendToExternals, &ste)
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	res.Pagination = pageRes

	return res, nil
}

func (k Keeper) DelegateKeysByExternalSigner(c context.Context, req *types.DelegateKeysByExternalSignerRequest) (*types.DelegateKeysByExternalSignerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	ethAddr := common.HexToAddress(req.ExternalSigner)
	orchAddr := k.GetExternalOrchestratorAddress(ctx, ethAddr)
	valAddr := k.GetOrchestratorValidatorAddress(ctx, orchAddr)
	res := &types.DelegateKeysByExternalSignerResponse{
		ValidatorAddress:    valAddr.String(),
		OrchestratorAddress: orchAddr.String(),
	}
	return res, nil
}

func (k Keeper) LatestSignerSetTx(c context.Context, req *types.LatestSignerSetTxRequest) (*types.SignerSetTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), append(append([]byte{types.OutgoingTxKey}, types.ChainID(req.ChainId).Bytes()...), types.SignerSetTxPrefixByte))
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	if !iter.Valid() {
		return nil, status.Errorf(codes.NotFound, "latest signer set not found")
	}

	var any cdctypes.Any
	k.cdc.MustUnmarshal(iter.Value(), &any)

	var otx types.OutgoingTx
	if err := k.cdc.UnpackAny(&any, &otx); err != nil {
		return nil, err
	}
	ss, ok := otx.(*types.SignerSetTx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "couldn't cast to signer set for latest")
	}
	return &types.SignerSetTxResponse{SignerSet: ss}, nil
}

func (k Keeper) SignerSetTx(c context.Context, req *types.SignerSetTxRequest) (*types.SignerSetTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	key := types.MakeSignerSetTxKey(types.ChainID(req.ChainId), req.SignerSetNonce)
	otx := k.GetOutgoingTx(ctx, types.ChainID(req.ChainId), key)
	if otx == nil {
		return &types.SignerSetTxResponse{}, nil
	}

	ss, ok := otx.(*types.SignerSetTx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "couldn't cast to signer set for %d", req.SignerSetNonce)
	}

	return &types.SignerSetTxResponse{SignerSet: ss}, nil
}

func (k Keeper) BatchTx(c context.Context, req *types.BatchTxRequest) (*types.BatchTxResponse, error) {
	res := &types.BatchTxResponse{}

	key := types.MakeBatchTxKey(types.ChainID(req.ChainId), req.ExternalTokenId, req.BatchNonce)
	otx := k.GetOutgoingTx(sdk.UnwrapSDKContext(c), types.ChainID(req.ChainId), key)
	if otx == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no batch tx found for %d %s", req.BatchNonce, req.ExternalTokenId)
	}
	batch, ok := otx.(*types.BatchTx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "couldn't cast to batch tx for %d %s", req.BatchNonce, req.ExternalTokenId)
	}
	res.Batch = batch

	return res, nil
}

func (k Keeper) ContractCallTx(c context.Context, req *types.ContractCallTxRequest) (*types.ContractCallTxResponse, error) {
	key := types.MakeContractCallTxKey(types.ChainID(req.ChainId), req.InvalidationScope, req.InvalidationNonce)
	otx := k.GetOutgoingTx(sdk.UnwrapSDKContext(c), types.ChainID(req.ChainId), key)
	if otx == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no contract call found for %d %s", req.InvalidationNonce, req.InvalidationScope)
	}

	cctx, ok := otx.(*types.ContractCallTx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "couldn't cast to contract call for %d %s", req.InvalidationNonce, req.InvalidationScope)
	}

	return &types.ContractCallTxResponse{LogicCall: cctx}, nil
}

func (k Keeper) SignerSetTxs(c context.Context, req *types.SignerSetTxsRequest) (*types.SignerSetTxsResponse, error) {
	var signers []*types.SignerSetTx
	pageRes, err := k.PaginateOutgoingTxsByType(sdk.UnwrapSDKContext(c), types.ChainID(req.ChainId), req.Pagination, types.SignerSetTxPrefixByte, func(_ []byte, otx types.OutgoingTx) (hit bool) {
		signer, ok := otx.(*types.SignerSetTx)
		if !ok {
			panic(sdkerrors.Wrapf(types.ErrInvalid, "couldn't cast to signer set for %s", otx))
		}
		signers = append(signers, signer)

		return true
	})
	if err != nil {
		return nil, err
	}

	return &types.SignerSetTxsResponse{SignerSets: signers, Pagination: pageRes}, nil
}

func (k Keeper) BatchTxs(c context.Context, req *types.BatchTxsRequest) (*types.BatchTxsResponse, error) {
	var batches []*types.BatchTx
	pageRes, err := k.PaginateOutgoingTxsByType(sdk.UnwrapSDKContext(c), types.ChainID(req.ChainId), req.Pagination, types.BatchTxPrefixByte, func(_ []byte, otx types.OutgoingTx) (hit bool) {
		batch, ok := otx.(*types.BatchTx)
		if !ok {
			panic(sdkerrors.Wrapf(types.ErrInvalid, "couldn't cast to batch tx for %s", otx))
		}
		batches = append(batches, batch)
		return true
	})
	if err != nil {
		return nil, err
	}

	return &types.BatchTxsResponse{Batches: batches, Pagination: pageRes}, nil
}

func (k Keeper) ContractCallTxs(c context.Context, req *types.ContractCallTxsRequest) (*types.ContractCallTxsResponse, error) {
	var calls []*types.ContractCallTx
	pageRes, err := k.PaginateOutgoingTxsByType(sdk.UnwrapSDKContext(c), types.ChainID(req.ChainId), req.Pagination, types.ContractCallTxPrefixByte, func(_ []byte, otx types.OutgoingTx) (hit bool) {
		call, ok := otx.(*types.ContractCallTx)
		if !ok {
			panic(sdkerrors.Wrapf(types.ErrInvalid, "couldn't cast to contract call for %s", otx))
		}
		calls = append(calls, call)
		return true
	})
	if err != nil {
		return nil, err
	}

	return &types.ContractCallTxsResponse{Calls: calls, Pagination: pageRes}, nil
}

func (k Keeper) SignerSetTxConfirmations(c context.Context, req *types.SignerSetTxConfirmationsRequest) (*types.SignerSetTxConfirmationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	key := types.MakeSignerSetTxKey(types.ChainID(req.ChainId), req.SignerSetNonce)

	var out []*types.SignerSetTxConfirmation
	k.iterateExternalSignatures(ctx, types.ChainID(req.ChainId), key, func(val sdk.ValAddress, sig []byte) bool {
		out = append(out, &types.SignerSetTxConfirmation{
			SignerSetNonce: req.SignerSetNonce,
			ExternalSigner: k.GetValidatorExternalAddress(ctx, val).Hex(),
			Signature:      sig,
		})
		return false
	})

	return &types.SignerSetTxConfirmationsResponse{Signatures: out}, nil
}

func (k Keeper) BatchTxConfirmations(c context.Context, req *types.BatchTxConfirmationsRequest) (*types.BatchTxConfirmationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	key := types.MakeBatchTxKey(types.ChainID(req.ChainId), req.ExternalTokenId, req.BatchNonce)

	var out []*types.BatchTxConfirmation
	k.iterateExternalSignatures(ctx, types.ChainID(req.ChainId), key, func(val sdk.ValAddress, sig []byte) bool {
		out = append(out, &types.BatchTxConfirmation{
			ExternalTokenId: req.ExternalTokenId,
			BatchNonce:      req.BatchNonce,
			ExternalSigner:  k.GetValidatorExternalAddress(ctx, val).Hex(),
			Signature:       sig,
		})
		return false
	})
	return &types.BatchTxConfirmationsResponse{Signatures: out}, nil
}

func (k Keeper) ContractCallTxConfirmations(c context.Context, req *types.ContractCallTxConfirmationsRequest) (*types.ContractCallTxConfirmationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	key := types.MakeContractCallTxKey(types.ChainID(req.ChainId), req.InvalidationScope, req.InvalidationNonce)

	var out []*types.ContractCallTxConfirmation
	k.iterateExternalSignatures(ctx, types.ChainID(req.ChainId), key, func(val sdk.ValAddress, sig []byte) bool {
		out = append(out, &types.ContractCallTxConfirmation{
			InvalidationScope: req.InvalidationScope,
			InvalidationNonce: req.InvalidationNonce,
			ExternalSigner:    k.GetValidatorExternalAddress(ctx, val).Hex(),
			Signature:         sig,
		})
		return false
	})
	return &types.ContractCallTxConfirmationsResponse{Signatures: out}, nil
}

func (k Keeper) UnsignedSignerSetTxs(c context.Context, req *types.UnsignedSignerSetTxsRequest) (*types.UnsignedSignerSetTxsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chainId := types.ChainID(req.ChainId)
	val, err := k.getSignerValidator(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	var signerSets []*types.SignerSetTx
	k.IterateOutgoingTxsByType(ctx, chainId, types.SignerSetTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
		sig := k.getExternalSignature(ctx, chainId, otx.GetStoreIndex(chainId), val)
		if len(sig) == 0 { // it's pending
			signerSet, ok := otx.(*types.SignerSetTx)
			if !ok {
				panic(sdkerrors.Wrapf(types.ErrInvalid, "couldn't cast to signer set for %s", otx))
			}
			signerSets = append(signerSets, signerSet)
		}
		return false
	})
	return &types.UnsignedSignerSetTxsResponse{SignerSets: signerSets}, nil
}

func (k Keeper) UnsignedBatchTxs(c context.Context, req *types.UnsignedBatchTxsRequest) (*types.UnsignedBatchTxsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	val, err := k.getSignerValidator(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	var batches []*types.BatchTx
	chainId := types.ChainID(req.ChainId)
	k.IterateOutgoingTxsByType(ctx, chainId, types.BatchTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
		sig := k.getExternalSignature(ctx, chainId, otx.GetStoreIndex(chainId), val)
		if len(sig) == 0 { // it's pending
			batch, ok := otx.(*types.BatchTx)
			if !ok {
				panic(sdkerrors.Wrapf(types.ErrInvalid, "couldn't cast to batch tx for %s", otx))
			}
			batches = append(batches, batch)
		}
		return false
	})
	return &types.UnsignedBatchTxsResponse{Batches: batches}, nil
}

func (k Keeper) UnsignedContractCallTxs(c context.Context, req *types.UnsignedContractCallTxsRequest) (*types.UnsignedContractCallTxsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	val, err := k.getSignerValidator(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	var calls []*types.ContractCallTx
	chainId := types.ChainID(req.ChainId)
	k.IterateOutgoingTxsByType(ctx, chainId, types.ContractCallTxPrefixByte, func(_ []byte, otx types.OutgoingTx) bool {
		sig := k.getExternalSignature(ctx, chainId, otx.GetStoreIndex(chainId), val)
		if len(sig) == 0 { // it's pending
			call, ok := otx.(*types.ContractCallTx)
			if !ok {
				panic(sdkerrors.Wrapf(types.ErrInvalid, "couldn't cast to contract call for %s", otx))
			}
			calls = append(calls, call)
		}
		return false
	})
	return &types.UnsignedContractCallTxsResponse{Calls: calls}, nil
}

func (k Keeper) BatchTxFees(c context.Context, req *types.BatchTxFeesRequest) (*types.BatchTxFeesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	res := &types.BatchTxFeesResponse{}

	// TODO: is this what we want here?
	// Should this calculation return a
	// map[contract_address]fees or something similar?
	k.IterateOutgoingTxsByType(ctx, types.ChainID(req.ChainId), types.BatchTxPrefixByte, func(key []byte, otx types.OutgoingTx) bool {
		btx, _ := otx.(*types.BatchTx)
		for _, tx := range btx.Transactions {
			res.Fees = append(res.Fees, tx.Fee.HubCoin(func(id uint64) (string, error) {
				info, err := k.TokenIdToTokenInfoLookup(ctx, id)
				if err != nil {
					return "", err
				}

				return info.Denom, nil
			}))
		}
		return false
	})

	return res, nil
}

func (k Keeper) ExternalIdToDenom(c context.Context, req *types.ExternalIdToDenomRequest) (*types.ExternalIdToDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	tokenInfo, err := k.ExternalIdToTokenInfoLookup(ctx, types.ChainID(req.ChainId), req.ExternalId)
	if err != nil {
		return nil, err
	}
	res := &types.ExternalIdToDenomResponse{
		Denom: tokenInfo.Denom,
	}
	return res, nil
}

func (k Keeper) DenomToExternalId(c context.Context, req *types.DenomToExternalIdRequest) (*types.DenomToExternalIdResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	tokenInfo, err := k.DenomToTokenInfoLookup(ctx, types.ChainID(req.ChainId), req.Denom)
	if err != nil {
		return nil, err
	}
	res := &types.DenomToExternalIdResponse{
		ExternalId: tokenInfo.ExternalTokenId,
	}
	return res, nil
}

func (k Keeper) DelegateKeysByValidator(c context.Context, req *types.DelegateKeysByValidatorRequest) (*types.DelegateKeysByValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	ethAddr := k.GetValidatorExternalAddress(ctx, valAddr)
	orchAddr := k.GetExternalOrchestratorAddress(ctx, ethAddr)
	res := &types.DelegateKeysByValidatorResponse{
		EthAddress:          ethAddr.Hex(),
		OrchestratorAddress: orchAddr.String(),
	}
	return res, nil
}

func (k Keeper) DelegateKeysByOrchestrator(c context.Context, req *types.DelegateKeysByOrchestratorRequest) (*types.DelegateKeysByOrchestratorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	orchAddr, err := sdk.AccAddressFromBech32(req.OrchestratorAddress)
	if err != nil {
		return nil, err
	}
	valAddr := k.GetOrchestratorValidatorAddress(ctx, orchAddr)
	ethAddr := k.GetValidatorExternalAddress(ctx, valAddr)
	res := &types.DelegateKeysByOrchestratorResponse{
		ValidatorAddress: valAddr.String(),
		ExternalSigner:   ethAddr.Hex(),
	}
	return res, nil
}

func (k Keeper) DelegateKeys(c context.Context, req *types.DelegateKeysRequest) (*types.DelegateKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	delegateKeys := k.getDelegateKeys(ctx)

	res := &types.DelegateKeysResponse{
		DelegateKeys: delegateKeys,
	}
	return res, nil
}
