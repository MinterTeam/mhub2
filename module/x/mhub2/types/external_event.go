package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

var (
	_ ExternalEvent = &SendToHubEvent{}
	_ ExternalEvent = &TransferToChainEvent{}
	_ ExternalEvent = &BatchExecutedEvent{}
	_ ExternalEvent = &ContractCallExecutedEvent{}
	_ ExternalEvent = &SignerSetTxExecutedEvent{}
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *ExternalEventVoteRecord) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var event ExternalEvent
	return unpacker.UnpackAny(m.Event, &event)
}

//////////
// Hash //
//////////

func (sthe *SendToHubEvent) Hash() tmbytes.HexBytes {
	rcv, _ := sdk.AccAddressFromBech32(sthe.CosmosReceiver)
	path := bytes.Join(
		[][]byte{
			sdk.Uint64ToBigEndian(sthe.EventNonce),
			[]byte(sthe.ExternalCoinId), // todo: check length ?
			sthe.Amount.BigInt().Bytes(),
			common.Hex2Bytes(sthe.Sender),
			rcv.Bytes(),
			sdk.Uint64ToBigEndian(sthe.ExternalHeight),
		},
		[]byte{},
	)
	hash := sha256.Sum256([]byte(path))
	return hash[:]
}

func (ttce *TransferToChainEvent) Hash() tmbytes.HexBytes {
	path := bytes.Join(
		[][]byte{
			sdk.Uint64ToBigEndian(ttce.EventNonce),
			[]byte(ttce.ExternalCoinId), // todo: check length ?
			ttce.Amount.BigInt().Bytes(),
			common.Hex2Bytes(ttce.Sender),
			[]byte(ttce.ExternalReceiver), // todo: check length ?
			[]byte(ttce.ReceiverChainId),  // todo: check length ?
			sdk.Uint64ToBigEndian(ttce.ExternalHeight),
		},
		[]byte{},
	)
	hash := sha256.Sum256([]byte(path))
	return hash[:]
}

func (bee *BatchExecutedEvent) Hash() tmbytes.HexBytes {
	path := bytes.Join(
		[][]byte{
			[]byte(bee.ExternalCoinId), // todo: check length ?
			sdk.Uint64ToBigEndian(bee.EventNonce),
			sdk.Uint64ToBigEndian(bee.BatchNonce),
			sdk.Uint64ToBigEndian(bee.ExternalHeight),
		},
		[]byte{},
	)
	hash := sha256.Sum256([]byte(path))
	return hash[:]
}

func (ccee *ContractCallExecutedEvent) Hash() tmbytes.HexBytes {
	path := bytes.Join(
		[][]byte{
			sdk.Uint64ToBigEndian(ccee.EventNonce),
			ccee.InvalidationScope,
			sdk.Uint64ToBigEndian(ccee.InvalidationNonce),
			sdk.Uint64ToBigEndian(ccee.ExternalHeight),
		},
		[]byte{},
	)
	hash := sha256.Sum256([]byte(path))
	return hash[:]
}

func (sse *SignerSetTxExecutedEvent) Hash() tmbytes.HexBytes {
	path := bytes.Join(
		[][]byte{
			sdk.Uint64ToBigEndian(sse.EventNonce),
			sdk.Uint64ToBigEndian(sse.SignerSetTxNonce),
			sdk.Uint64ToBigEndian(sse.ExternalHeight),
			ExternalSigners(sse.Members).Hash(),
		},
		[]byte{},
	)
	hash := sha256.Sum256(([]byte(path)))
	return hash[:]
}

//////////////
// Validate //
//////////////

func (stce *SendToHubEvent) Validate() error {
	if stce.EventNonce == 0 {
		return fmt.Errorf("event nonce cannot be 0")
	}
	//if !common.IsHexAddress(stce.TokenContract) {
	//	return sdkerrors.Wrap(ErrInvalid, "ethereum contract address") todo
	//}
	if stce.Amount.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "amount must be positive")
	}
	if !common.IsHexAddress(stce.Sender) {
		return sdkerrors.Wrap(ErrInvalid, "external sender")
	}
	if _, err := sdk.AccAddressFromBech32(stce.CosmosReceiver); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, stce.CosmosReceiver)
	}
	return nil
}

func (ttce *TransferToChainEvent) Validate() error {
	if ttce.EventNonce == 0 {
		return fmt.Errorf("event nonce cannot be 0")
	}
	//if !common.IsHexAddress(stce.TokenContract) {
	//	return sdkerrors.Wrap(ErrInvalid, "ethereum contract address") todo
	//}
	if ttce.Amount.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "amount must be positive")
	}
	if !common.IsHexAddress(ttce.Sender) {
		return sdkerrors.Wrap(ErrInvalid, "external sender")
	}
	if !common.IsHexAddress(ttce.ExternalReceiver) {
		return sdkerrors.Wrap(ErrInvalid, "external receiver")
	}
	return nil
}

func (bee *BatchExecutedEvent) Validate() error {
	if bee.EventNonce == 0 {
		return fmt.Errorf("event nonce cannot be 0")
	}
	//if !common.IsHexAddress(bee.TokenContract) {
	//	return sdkerrors.Wrap(ErrInvalid, "ethereum contract address") todo
	//}
	return nil
}

func (ccee *ContractCallExecutedEvent) Validate() error {
	if ccee.EventNonce == 0 {
		return fmt.Errorf("event nonce cannot be 0")
	}
	return nil
}

func (sse *SignerSetTxExecutedEvent) Validate() error {
	if sse.EventNonce == 0 {
		return fmt.Errorf("event nonce cannot be 0")
	}
	if sse.Members == nil {
		return fmt.Errorf("members cannot be nil")
	}
	for i, member := range sse.Members {
		if err := member.ValidateBasic(); err != nil {
			return fmt.Errorf("ethereum signer %d error: %w", i, err)
		}
	}
	return nil
}
