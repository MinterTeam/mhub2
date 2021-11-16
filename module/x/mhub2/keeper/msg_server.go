package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
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

func (k msgServer) SetDelegateKeys(c context.Context, msg *types.MsgDelegateKeys) (*types.MsgDelegateKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	orchAddr, err := sdk.AccAddressFromBech32(msg.OrchestratorAddress)
	if err != nil {
		return nil, err
	}

	ethAddr := common.HexToAddress(msg.ExternalAddress)

	// ensure that the validator exists
	if k.Keeper.StakingKeeper.Validator(ctx, valAddr) == nil {
		return nil, sdkerrors.Wrap(stakingtypes.ErrNoValidatorFound, valAddr.String())
	}

	// check if the Ethereum address is currently not used
	validators := k.getValidatorsByExternalAddress(ctx, ethAddr)
	if len(validators) > 0 {
		return nil, sdkerrors.Wrapf(types.ErrDelegateKeys, "ethereum address %s in use", ethAddr)
	}

	// check if the orchestrator address is currently not used
	ethAddrs := k.getExternalAddressesByOrchestrator(ctx, orchAddr)
	if len(ethAddrs) > 0 {
		return nil, sdkerrors.Wrapf(types.ErrDelegateKeys, "orchestrator address %s in use", orchAddr)
	}

	valAccAddr := sdk.AccAddress(valAddr)
	valAccSeq, err := k.accountKeeper.GetSequence(ctx, valAccAddr)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrDelegateKeys, "failed to get sequence for validator account %s", valAccAddr)
	}

	var nonce uint64
	if valAccSeq > 0 {
		nonce = valAccSeq - 1
	}

	signMsgBz := k.cdc.MustMarshal(&types.DelegateKeysSignMsg{
		ValidatorAddress: valAddr.String(),
		// We decrement since we process the message after the ante-handler which
		// increments the nonce.
		Nonce: nonce,
	})

	hash := crypto.Keccak256Hash(signMsgBz).Bytes()

	if err = types.ValidateEthereumSignature(hash, msg.EthSignature, ethAddr); err != nil {
		// TODO: ???
		//return nil, sdkerrors.Wrapf(
		//	types.ErrDelegateKeys,
		//	"failed to validate delegate keys signature for Ethereum address %X; %s",
		//	ethAddr, err,
		//)
	}

	k.SetOrchestratorValidatorAddress(ctx, valAddr, orchAddr)
	k.setValidatorExternalAddress(ctx, valAddr, ethAddr)
	k.setExternalOrchestratorAddress(ctx, ethAddr, orchAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeySetOrchestratorAddr, orchAddr.String()),
			sdk.NewAttribute(types.AttributeKeySetEthereumAddr, ethAddr.Hex()),
			sdk.NewAttribute(types.AttributeKeyValidatorAddr, valAddr.String()),
		),
	)

	return &types.MsgDelegateKeysResponse{}, nil

}

// SubmitTxConfirmation handles MsgSubmitTxConfirmation
func (k msgServer) SubmitTxConfirmation(c context.Context, msg *types.MsgSubmitExternalTxConfirmation) (*types.MsgSubmitTxConfirmationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	chainId := types.ChainID(msg.ChainId)

	confirmation, err := types.UnpackConfirmation(msg.Confirmation)
	if err != nil {
		return nil, err
	}

	val, err := k.getSignerValidator(ctx, msg.Signer)
	if err != nil {
		return nil, err
	}

	otx := k.GetOutgoingTx(ctx, types.ChainID(msg.ChainId), confirmation.GetStoreIndex(chainId))
	if otx == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "couldn't find outgoing tx")
	}

	gravityID := k.getGravityID(ctx)
	checkpoint := otx.GetCheckpoint([]byte(gravityID))

	ethAddress := k.GetValidatorExternalAddress(ctx, val)
	if ethAddress != confirmation.GetSigner() {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "eth address does not match signer eth address")
	}

	if err = types.ValidateEthereumSignature(checkpoint, confirmation.GetSignature(), ethAddress); err != nil {
		// todo
		//return nil, sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf(
		//	"signature verification failed ethAddress %s gravityID %s checkpoint %s typeURL %s signature %s err %s",
		//	ethAddress.Hex(),
		//	gravityID,
		//	hex.EncodeToString(checkpoint),
		//	msg.Confirmation.TypeUrl,
		//	hex.EncodeToString(confirmation.GetSignature()),
		//	err,
		//))
	}

	// TODO: should validators be able to overwrite their signatures?
	if k.getExternalSignature(ctx, types.ChainID(msg.ChainId), confirmation.GetStoreIndex(chainId), val) != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "signature duplicate")
	}

	key := k.SetExternalSignature(ctx, types.ChainID(msg.ChainId), confirmation, val)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeyEthereumSignatureKey, string(key)),
		),
	)

	return &types.MsgSubmitTxConfirmationResponse{}, nil
}

// SubmitExternalEvent handles MsgSubmitExternalEvent
func (k msgServer) SubmitExternalEvent(c context.Context, msg *types.MsgSubmitExternalEvent) (*types.MsgSubmitExternalEventResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	event, err := types.UnpackEvent(msg.Event)
	if err != nil {
		return nil, err
	}

	// return an error if the validator isn't in the active set
	val, err := k.getSignerValidator(ctx, msg.Signer)
	if err != nil {
		return nil, err
	}

	// Add the claim to the store
	_, err = k.recordEventVote(ctx, types.ChainID(msg.ChainId), event, val)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create event vote record")
	}

	// Emit the handle message event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, fmt.Sprintf("%T", event)),
			// TODO: maybe return something better here? is this the right string representation?
			sdk.NewAttribute(types.AttributeKeyEthereumEventVoteRecordID, string(types.MakeExternalEventVoteRecordKey(types.ChainID(msg.ChainId), event.GetEventNonce(), event.Hash()))),
		),
	)

	return &types.MsgSubmitExternalEventResponse{}, nil
}

// SendToExternal handles MsgSendToExternal
func (k msgServer) SendToExternal(c context.Context, msg *types.MsgSendToExternal) (*types.MsgSendToExternalResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	tokenInfo, err := k.DenomToTokenInfoLookup(ctx, types.ChainID(msg.ChainId), msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	commission := k.GetCommissionForHolder(ctx, []string{sender.String(), msg.ExternalRecipient}, tokenInfo.Commission).Mul(msg.Amount.Amount.Add(msg.BridgeFee.Amount).ToDec()).TruncateInt()

	txID, err := k.createSendToExternal(ctx, types.ChainID(msg.ChainId), sender, msg.ExternalRecipient, msg.Amount.SubAmount(commission), msg.BridgeFee, sdk.NewCoin(msg.Amount.Denom, commission), fmt.Sprintf("%x", sha256.Sum256(ctx.TxBytes())), "hub", sender.String())
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents([]sdk.Event{
		sdk.NewEvent(
			types.EventTypeBridgeWithdrawalReceived,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyContract, k.getBridgeContractAddress(ctx)),
			sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.getBridgeChainID(ctx)))),
			sdk.NewAttribute(types.AttributeKeyOutgoingTXID, strconv.Itoa(int(txID))),
			sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprint(txID)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeyOutgoingTXID, fmt.Sprint(txID)),
		),
	})

	return &types.MsgSendToExternalResponse{Id: txID}, nil
}

// RequestBatchTx handles MsgRequestBatchTx
func (k msgServer) RequestBatchTx(c context.Context, msg *types.MsgRequestBatchTx) (*types.MsgRequestBatchTxResponse, error) {
	// TODO: limit this to only orchestrators and validators?
	ctx := sdk.UnwrapSDKContext(c)

	tokenInfo, err := k.DenomToTokenInfoLookup(ctx, types.ChainID(msg.ChainId), msg.Denom)
	if err != nil {
		return nil, err
	}

	batchID := k.BuildBatchTx(ctx, types.ChainID(msg.ChainId), tokenInfo.ExternalTokenId, BatchTxSize)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			//sdk.NewAttribute(types.AttributeKeyContract, tokenContract.Hex()), todo
			sdk.NewAttribute(types.AttributeKeyBatchNonce, fmt.Sprint(batchID.BatchNonce)),
		),
	)

	return &types.MsgRequestBatchTxResponse{}, nil
}

func (k msgServer) CancelSendToExternal(c context.Context, msg *types.MsgCancelSendToExternal) (*types.MsgCancelSendToExternalResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	err := k.Keeper.cancelSendToExternal(ctx, types.ChainID(msg.ChainId), msg.Id, msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents([]sdk.Event{
		sdk.NewEvent(
			types.EventTypeBridgeWithdrawCanceled,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyContract, k.getBridgeContractAddress(ctx)),
			sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.getBridgeChainID(ctx)))),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, msg.Type()),
			sdk.NewAttribute(types.AttributeKeyOutgoingTXID, fmt.Sprint(msg.Id)),
		),
	})

	return &types.MsgCancelSendToExternalResponse{}, nil
}

// getSignerValidator takes an sdk.AccAddress that represents either a validator or orchestrator address and returns
// the assoicated validator address
func (k Keeper) getSignerValidator(ctx sdk.Context, signerString string) (sdk.ValAddress, error) {
	signer, err := sdk.AccAddressFromBech32(signerString)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "signer address")
	}
	var validatorI stakingtypes.ValidatorI
	if validator := k.GetOrchestratorValidatorAddress(ctx, signer); validator == nil {
		validatorI = k.StakingKeeper.Validator(ctx, sdk.ValAddress(signer))
	} else {
		validatorI = k.StakingKeeper.Validator(ctx, validator)
	}

	if validatorI == nil {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "not orchestrator or validator")
	} else if !validatorI.IsBonded() {
		return nil, sdkerrors.Wrap(types.ErrInvalid, fmt.Sprintf("validator is not bonded: %s", validatorI.GetOperator()))
	}

	return validatorI.GetOperator(), nil
}
