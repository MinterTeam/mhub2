package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

// EthereumAddrLessThan migrates the Ethereum address less than function
func EthereumAddrLessThan(e, o string) bool {
	return bytes.Compare([]byte(e)[:], []byte(o)[:]) == -1
}

/////////////////////////
//   External Token    //
/////////////////////////

// NewExternalToken returns a new instance of an external token
func NewExternalToken(amount uint64, id uint64, externalId string) ExternalToken {
	return ExternalToken{
		TokenId:         id,
		ExternalTokenId: externalId,
		Amount:          sdk.NewIntFromUint64(amount),
	}
}

func NewSDKIntExternalToken(amount sdk.Int, id uint64, externalId string) ExternalToken {
	return ExternalToken{
		TokenId:         id,
		ExternalTokenId: externalId,
		Amount:          amount,
	}
}

// HubCoin returns the mhub2 representation of the External Token
func (e ExternalToken) HubCoin(denomResolver func(id uint64) (string, error)) sdk.Coin {
	denom, err := denomResolver(e.TokenId)
	if err != nil {
		denom = fmt.Sprintf("token/%d", e.TokenId)
	}

	return sdk.Coin{Amount: e.Amount, Denom: denom}
}

func NewSendToExternalTx(id uint64, chainId ChainID, tokenId uint64, externalTokenId string, sender sdk.AccAddress, recipient common.Address, amount, feeAmount, valCommission uint64, txHash string) *SendToExternal {
	return &SendToExternal{
		Id:                id,
		Sender:            sender.String(),
		ExternalRecipient: recipient.Hex(),
		ChainId:           chainId.String(),
		Token:             NewExternalToken(amount, tokenId, externalTokenId),
		Fee:               NewExternalToken(feeAmount, tokenId, externalTokenId),
		TxHash:            txHash,
		ValCommission:     NewExternalToken(valCommission, tokenId, externalTokenId),
	}
}
