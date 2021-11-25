package command

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

const TypeSendToEth = "send_to_eth"
const TypeSendToHub = "send_to_hub"
const TypeSendToBsc = "send_to_bsc"

type Command struct {
	Type      string `json:"type"`
	Recipient string `json:"recipient"`
	Fee       string `json:"fee"`
}

func (cmd Command) Validate(amount sdk.Int) error {
	switch cmd.Type {
	case TypeSendToEth, TypeSendToBsc:
		if !common.IsHexAddress(cmd.Recipient) {
			return errors.New("wrong recipient")
		}
	case TypeSendToHub:
		if _, err := sdk.AccAddressFromBech32(cmd.Recipient); err != nil {
			return err
		}
	default:
		return errors.New("wrong type")
	}

	fee, ok := sdk.NewIntFromString(cmd.Fee)
	if !ok {
		return errors.New("incorrect fee")
	}

	if amount.Sub(amount.QuoRaw(100)).LTE(fee) {
		return errors.New("incorrect fee")
	}

	return nil
}
