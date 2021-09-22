package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ ExternalTxConfirmation = &SignerSetTxConfirmation{}
	_ ExternalTxConfirmation = &ContractCallTxConfirmation{}
	_ ExternalTxConfirmation = &BatchTxConfirmation{}
)

///////////////
// GetSigner //
///////////////

func (u *SignerSetTxConfirmation) GetSigner() common.Address {
	return common.HexToAddress(u.ExternalSigner)
}

func (u *ContractCallTxConfirmation) GetSigner() common.Address {
	return common.HexToAddress(u.ExternalSigner)
}

func (u *BatchTxConfirmation) GetSigner() common.Address {
	return common.HexToAddress(u.ExternalSigner)
}

///////////////////
// GetStoreIndex //
///////////////////

func (sstx *SignerSetTxConfirmation) GetStoreIndex(chainId ChainID) []byte {
	return MakeSignerSetTxKey(chainId, sstx.SignerSetNonce)
}

func (btx *BatchTxConfirmation) GetStoreIndex(chainId ChainID) []byte {
	return MakeBatchTxKey(chainId, btx.ExternalTokenId, btx.BatchNonce)
}

func (cctx *ContractCallTxConfirmation) GetStoreIndex(chainId ChainID) []byte {
	return MakeContractCallTxKey(chainId, cctx.InvalidationScope, cctx.InvalidationNonce)
}

//////////////
// Validate //
//////////////

func (u *SignerSetTxConfirmation) Validate() error {
	if u.SignerSetNonce == 0 {
		return fmt.Errorf("nonce must be set")
	}
	if !common.IsHexAddress(u.ExternalSigner) {
		return sdkerrors.Wrap(ErrInvalid, "ethereum signer must be address")
	}
	if u.Signature == nil {
		return fmt.Errorf("signature must be set")
	}
	return nil
}

func (u *ContractCallTxConfirmation) Validate() error {
	if u.InvalidationNonce == 0 {
		return fmt.Errorf("invalidation nonce must be set")
	}
	if u.InvalidationScope == nil {
		return fmt.Errorf("invalidation scope must be set")
	}
	if !common.IsHexAddress(u.ExternalSigner) {
		return sdkerrors.Wrap(ErrInvalid, "ethereum signer must be address")
	}
	if u.Signature == nil {
		return fmt.Errorf("signature must be set")
	}
	return nil
}

func (u *BatchTxConfirmation) Validate() error {
	if u.BatchNonce == 0 {
		return fmt.Errorf("nonce must be set")
	}
	//if !common.IsHexAddress(u.TokenContract) {
	//	return fmt.Errorf("token contract address must be valid ethereum address")
	//}
	if !common.IsHexAddress(u.ExternalSigner) {
		return sdkerrors.Wrap(ErrInvalid, "ethereum signer must be address")
	}
	if u.Signature == nil {
		return fmt.Errorf("signature must be set")
	}
	return nil
}
