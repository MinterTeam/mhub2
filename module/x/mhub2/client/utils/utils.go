package utils

import (
	"io/ioutil"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type (
	// ColdStorageTransfersJSON defines a slice of ColdStorageTransferJSON objects which can be
	// converted to a slice of ColdStorageTransfer objects.
	ColdStorageTransfersJSON []ColdStorageTransferJSON

	// ColdStorageTransferJSON defines a parameter change used in JSON input. This
	// allows values to be specified in raw JSON instead of being string encoded.
	ColdStorageTransferJSON struct {
		ChainId string    `json:"chain_id" yaml:"chain_id"`
		Amount  sdk.Coins `json:"amount" yaml:"amount"`
	}

	// ColdStorageTransferProposalJSON defines a ParameterChangeProposal with a deposit used
	// to parse parameter change proposals from a JSON file.
	ColdStorageTransferProposalJSON struct {
		ChainId string    `json:"chain_id" yaml:"chain_id"`
		Amount  sdk.Coins `json:"amount" yaml:"amount"`
		Deposit string    `json:"deposit" yaml:"deposit"`
	}

	// ColdStorageTransferProposalReq defines a parameter change proposal request body.
	ColdStorageTransferProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		ChainId  string         `json:"chain_id" yaml:"chain_id"`
		Amount   sdk.Coins      `json:"amount" yaml:"amount"`
		Proposer sdk.AccAddress `json:"proposer" yaml:"proposer"`
		Deposit  sdk.Coins      `json:"deposit" yaml:"deposit"`
	}
)

func NewColdStorageTransferJSON(chainId types.ChainID, amount sdk.Coins) ColdStorageTransferJSON {
	return ColdStorageTransferJSON{chainId.String(), amount}
}

// ToColdStorageTransfer converts a ColdStorageTransferJSON object to ColdStorageTransfer.
func (pcj ColdStorageTransferJSON) ToColdStorageTransfer() types.ColdStorageTransferProposal {
	return *types.NewColdStorageTransferProposal(types.ChainID(pcj.ChainId), pcj.Amount)
}

// ToColdStorageTransfers converts a slice of ColdStorageTransferJSON objects to a slice of
// ColdStorageTransfer.
func (pcj ColdStorageTransfersJSON) ToColdStorageTransfers() []types.ColdStorageTransferProposal {
	res := make([]types.ColdStorageTransferProposal, len(pcj))
	for i, pc := range pcj {
		res[i] = pc.ToColdStorageTransfer()
	}
	return res
}

// ParseColdStorageTransferProposalJSON reads and parses a ColdStorageTransferProposalJSON from
// file.
func ParseColdStorageTransferProposalJSON(cdc *codec.LegacyAmino, proposalFile string) (ColdStorageTransferProposalJSON, error) {
	proposal := ColdStorageTransferProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
