package types

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "mhub2"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the module name router key
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)

const (
	_ = byte(iota)
	// Key Delegation
	ValidatorExternalAddressKey
	OrchestratorValidatorAddressKey
	ExternalOrchestratorAddressKey

	// Core types
	ExternalSignatureKey
	ExternalEventVoteRecordKey
	OutgoingTxKey
	SendToExternalKey

	// Latest nonce indexes
	LastEventNonceByValidatorKey
	LastObservedEventNonceKey
	LatestSignerSetTxNonceKey
	LastSlashedOutgoingTxBlockKey
	LastSlashedSignerSetTxNonceKey
	LastOutgoingBatchNonceKey

	OutgoingSequence

	// LastSendToExternalIDKey indexes the lastTxPoolID
	LastSendToExternalIDKey

	// LastExternalBlockHeightKey indexes the latest Ethereum block height
	LastExternalBlockHeightKey

	TokenInfosKey

	// LastUnBondingBlockHeightKey indexes the last validator unbonding block height
	LastUnBondingBlockHeightKey

	LastObservedSignerSetKey

	TxStatusKey

	TxFeeRecordKey
)

////////////////////
// Key Delegation //
////////////////////

// MakeOrchestratorValidatorAddressKey returns the following key format
// prefix
// [0xe8][cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func MakeOrchestratorValidatorAddressKey(chainId ChainID, orc sdk.AccAddress) []byte {
	return bytes.Join([][]byte{{OrchestratorValidatorAddressKey}, chainId.Bytes(), orc.Bytes()}, []byte{})
}

// MakeValidatorExternalAddressKey returns the following key format
// prefix              cosmos-validator
// [0x0][cosmosvaloper1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func MakeValidatorExternalAddressKey(chainId ChainID, validator sdk.ValAddress) []byte {
	return bytes.Join([][]byte{{ValidatorExternalAddressKey}, chainId.Bytes(), validator.Bytes()}, []byte{})
}

// MakeExternalOrchestratorAddressKey returns the following key format
// prefix              cosmos-validator
// [0x0][cosmosvaloper1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func MakeExternalOrchestratorAddressKey(chainId ChainID, eth common.Address) []byte {
	return bytes.Join([][]byte{{ExternalOrchestratorAddressKey}, chainId.Bytes(), eth.Bytes()}, []byte{})
}

/////////////////////////
// External Signatures //
/////////////////////////

// MakeExternalSignatureKey returns the following key format
// prefix   nonce                    validator-address // todo: add chain id
// [0x0][0 0 0 0 0 0 0 1][cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func MakeExternalSignatureKey(chainId ChainID, storeIndex []byte, validator sdk.ValAddress) []byte {
	return bytes.Join([][]byte{{ExternalSignatureKey}, chainId.Bytes(), storeIndex, validator.Bytes()}, []byte{})
}

/////////////////////////////////
// External Event Vote Records //
/////////////////////////////////

// MakeExternalEventVoteRecordKey returns the following key format
// prefix     nonce                             claim-details-hash // todo: add chain id
// [0x5][0 0 0 0 0 0 0 1][fd1af8cec6c67fcf156f1b61fdf91ebc04d05484d007436e75342fc05bbff35a]
func MakeExternalEventVoteRecordKey(chainId ChainID, eventNonce uint64, claimHash []byte) []byte {
	return bytes.Join([][]byte{{ExternalEventVoteRecordKey}, chainId.Bytes(), sdk.Uint64ToBigEndian(eventNonce), claimHash}, []byte{})
}

//////////////////
// Outgoing Txs //
//////////////////

// MakeOutgoingTxKey returns the store index passed with a prefix
func MakeOutgoingTxKey(chainId ChainID, storeIndex []byte) []byte {
	return bytes.Join([][]byte{{OutgoingTxKey}, chainId.Bytes(), storeIndex}, []byte{})
}

//////////////////////
// Send To Etheruem //
//////////////////////

// MakeSendToExternalKey returns the following key format
// prefix token_id fee_amount id
// [0x9][000][1000000000][0 0 0 0 0 0 0 1]
func MakeSendToExternalKey(chainId ChainID, id uint64, fee ExternalToken) []byte {
	amount := make([]byte, 32)
	return bytes.Join([][]byte{{SendToExternalKey}, chainId.Bytes(), []byte(fee.ExternalTokenId), fee.Amount.BigInt().FillBytes(amount), sdk.Uint64ToBigEndian(id)}, []byte{})
}

// MakeLastEventNonceByValidatorKey indexes lateset event nonce by validator
// MakeLastEventNonceByValidatorKey returns the following key format
// prefix              cosmos-validator
// [0x0][cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn]
func MakeLastEventNonceByValidatorKey(chainId ChainID, validator sdk.ValAddress) []byte {
	return bytes.Join([][]byte{{LastEventNonceByValidatorKey}, chainId.Bytes(), validator.Bytes()}, []byte{})
}

func MakeSignerSetTxKey(chainId ChainID, nonce uint64) []byte {
	return bytes.Join([][]byte{{SignerSetTxPrefixByte}, chainId.Bytes(), sdk.Uint64ToBigEndian(nonce)}, []byte{})
}

func MakeBatchTxKey(chainId ChainID, externalTokenId string, nonce uint64) []byte {
	return bytes.Join([][]byte{{BatchTxPrefixByte}, chainId.Bytes(), []byte(externalTokenId), sdk.Uint64ToBigEndian(nonce)}, []byte{})
}

func MakeContractCallTxKey(chainId ChainID, invalscope []byte, invalnonce uint64) []byte {
	return bytes.Join([][]byte{{ContractCallTxPrefixByte}, chainId.Bytes(), invalscope, sdk.Uint64ToBigEndian(invalnonce)}, []byte{})
}

func GetTxStatusKey(inTxHash string) []byte {
	return bytes.Join([][]byte{{TxStatusKey}, []byte(inTxHash)}, []byte{})
}

func GetTxFeeRecordKey(inTxHash string) []byte {
	return bytes.Join([][]byte{{TxFeeRecordKey}, []byte(inTxHash)}, []byte{})
}
