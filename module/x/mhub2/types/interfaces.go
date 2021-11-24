package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// ExternalTxConfirmation represents one validators signature for a given
// outgoing ethereum transaction
type ExternalTxConfirmation interface {
	proto.Message

	GetSigner() common.Address
	GetSignature() []byte
	GetStoreIndex(chainId ChainID) []byte
	Validate() error
}

// ExternalEvent represents a event from the mhub2 contract
// on the counterparty ethereum chain
type ExternalEvent interface {
	proto.Message

	GetEventNonce() uint64
	GetExternalHeight() uint64
	Hash() tmbytes.HexBytes
	Validate(ChainID) error
}

type OutgoingTx interface {
	// NOTE: currently the function signatures here don't match, figure out how to do this proprly
	// maybe add an interface arg here and typecheck in each implementation?

	// The only one that will be problematic is BatchTx which needs to pull all the constituent
	// transactions before calculating the checkpoint
	GetCheckpoint([]byte) []byte
	GetStoreIndex(chainId ChainID) []byte
	GetCosmosHeight() uint64
	SetSequence(seq uint64)
}
