package types

import (
	"fmt"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = &MsgDelegateKeys{}
	_ sdk.Msg = &MsgSendToExternal{}
	_ sdk.Msg = &MsgCancelSendToExternal{}
	_ sdk.Msg = &MsgRequestBatchTx{}
	_ sdk.Msg = &MsgSubmitExternalEvent{}
	_ sdk.Msg = &MsgSubmitExternalTxConfirmation{}

	_ cdctypes.UnpackInterfacesMessage = &MsgSubmitExternalEvent{}
	_ cdctypes.UnpackInterfacesMessage = &MsgSubmitExternalTxConfirmation{}
	_ cdctypes.UnpackInterfacesMessage = &ExternalEventVoteRecord{}
)

// NewMsgDelegateKeys returns a reference to a new MsgDelegateKeys.
func NewMsgDelegateKeys(val sdk.ValAddress, chainId ChainID, orchAddr sdk.AccAddress, ethAddr string, ethSig []byte) *MsgDelegateKeys {
	return &MsgDelegateKeys{
		ValidatorAddress:    val.String(),
		OrchestratorAddress: orchAddr.String(),
		ExternalAddress:     ethAddr,
		EthSignature:        ethSig,
		ChainId:             chainId.String(),
	}
}

// Route should return the name of the module
func (msg *MsgDelegateKeys) Route() string { return RouterKey }

// Type should return the action
func (msg *MsgDelegateKeys) Type() string { return "delegate_keys" }

// ValidateBasic performs stateless checks
func (msg *MsgDelegateKeys) ValidateBasic() (err error) {
	if _, err = sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.ValidatorAddress)
	}
	if _, err = sdk.AccAddressFromBech32(msg.OrchestratorAddress); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.OrchestratorAddress)
	}
	if !common.IsHexAddress(msg.ExternalAddress) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "ethereum address")
	}
	if len(msg.EthSignature) == 0 {
		return ErrEmptyEthSig
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg *MsgDelegateKeys) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg *MsgDelegateKeys) GetSigners() []sdk.AccAddress {
	acc, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sdk.AccAddress(acc)}
}

// Route should return the name of the module
func (msg *MsgSubmitExternalEvent) Route() string { return RouterKey }

// Type should return the action
func (msg *MsgSubmitExternalEvent) Type() string { return "submit_ethereum_event" }

// ValidateBasic performs stateless checks
func (msg *MsgSubmitExternalEvent) ValidateBasic() (err error) {
	if _, err = sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}

	event, err := UnpackEvent(msg.Event)
	if err != nil {
		return err
	}
	return event.Validate(ChainID(msg.ChainId))
}

// GetSignBytes encodes the message for signing
func (msg *MsgSubmitExternalEvent) GetSignBytes() []byte {
	panic(fmt.Errorf("deprecated"))
}

// GetSigners defines whose signature is required
func (msg *MsgSubmitExternalEvent) GetSigners() []sdk.AccAddress {
	// TODO: figure out how to convert between AccAddress and ValAddress properly
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{acc}
}

func (msg *MsgSubmitExternalEvent) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var event ExternalEvent
	return unpacker.UnpackAny(msg.Event, &event)
}

// Route should return the name of the module
func (msg *MsgSubmitExternalTxConfirmation) Route() string { return RouterKey }

// Type should return the action
func (msg *MsgSubmitExternalTxConfirmation) Type() string { return "submit_ethereum_signature" }

// ValidateBasic performs stateless checks
func (msg *MsgSubmitExternalTxConfirmation) ValidateBasic() (err error) {
	if _, err = sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}

	event, err := UnpackConfirmation(msg.Confirmation)

	if err != nil {
		return err
	}

	return event.Validate()
}

// GetSignBytes encodes the message for signing
func (msg *MsgSubmitExternalTxConfirmation) GetSignBytes() []byte {
	panic(fmt.Errorf("deprecated"))
}

// GetSigners defines whose signature is required
func (msg *MsgSubmitExternalTxConfirmation) GetSigners() []sdk.AccAddress {
	// TODO: figure out how to convert between AccAddress and ValAddress properly
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{acc}
}

func (msg *MsgSubmitExternalTxConfirmation) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var sig ExternalTxConfirmation
	return unpacker.UnpackAny(msg.Confirmation, &sig)
}

// NewMsgSendToExternal returns a new MsgSendToEthereum
func NewMsgSendToExternal(sender sdk.AccAddress, destAddress string, send sdk.Coin, bridgeFee sdk.Coin) *MsgSendToExternal {
	return &MsgSendToExternal{
		Sender:            sender.String(),
		ExternalRecipient: destAddress,
		Amount:            send,
		BridgeFee:         bridgeFee,
	}
}

// Route should return the name of the module
func (msg MsgSendToExternal) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSendToExternal) Type() string { return "send_to_eth" }

// ValidateBasic runs stateless checks on the message
// Checks if the Eth address is valid
func (msg MsgSendToExternal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Sender)
	}

	// fee and send must be of the same denom
	// this check is VERY IMPORTANT
	if msg.Amount.Denom != msg.BridgeFee.Denom {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins,
			fmt.Sprintf("fee and amount must be the same type %s != %s", msg.Amount.Denom, msg.BridgeFee.Denom))
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "amount")
	}
	if !msg.BridgeFee.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "fee")
	}
	if !common.IsHexAddress(msg.ExternalRecipient) || len(msg.ExternalRecipient) != common.AddressLength*2+2 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "external address")
	}

	if msg.ExternalRecipient == "0x0000000000000000000000000000000000000000" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot withdraw to zero address")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSendToExternal) GetSignBytes() []byte {
	panic(fmt.Errorf("deprecated"))
}

// GetSigners defines whose signature is required
func (msg MsgSendToExternal) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// NewMsgRequestBatchTx returns a new msgRequestBatch
func NewMsgRequestBatchTx(denom string, signer sdk.AccAddress) *MsgRequestBatchTx {
	return &MsgRequestBatchTx{
		Denom:  denom,
		Signer: signer.String(),
	}
}

// Route should return the name of the module
func (msg MsgRequestBatchTx) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRequestBatchTx) Type() string { return "request_batch" }

// ValidateBasic performs stateless checks
func (msg MsgRequestBatchTx) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return sdkerrors.Wrap(err, "denom is invalid")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgRequestBatchTx) GetSignBytes() []byte {
	panic(fmt.Errorf("deprecated"))
}

// GetSigners defines whose signature is required
func (msg MsgRequestBatchTx) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}

// NewMsgCancelSendToExternal returns a new MsgCancelSendToEthereum
func NewMsgCancelSendToExternal(id uint64, chainId ChainID, orchestrator sdk.AccAddress) *MsgCancelSendToExternal {
	return &MsgCancelSendToExternal{
		Id:      id,
		ChainId: chainId.String(),
		Sender:  orchestrator.String(),
	}
}

// Route should return the name of the module
func (msg MsgCancelSendToExternal) Route() string { return RouterKey }

// Type should return the action
func (msg MsgCancelSendToExternal) Type() string { return "cancel_send_to_external" }

// ValidateBasic performs stateless checks
func (msg MsgCancelSendToExternal) ValidateBasic() error {
	if msg.Id == 0 {
		return sdkerrors.Wrap(ErrInvalid, "Id cannot be 0")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Sender)
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgCancelSendToExternal) GetSignBytes() []byte {
	panic(fmt.Errorf("deprecated"))
}

// GetSigners defines whose signature is required
func (msg MsgCancelSendToExternal) GetSigners() []sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{acc}
}
