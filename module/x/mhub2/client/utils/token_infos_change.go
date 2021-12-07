package utils

import (
	"io/ioutil"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

type (
	// TokenInfosChangesJSON defines a slice of TokenInfosChangeJSON objects which can be
	// converted to a slice of TokenInfosChange objects.
	TokenInfosChangesJSON []TokenInfosChangeJSON

	// TokenInfosChangeJSON defines a parameter change used in JSON input. This
	// allows values to be specified in raw JSON instead of being string encoded.
	TokenInfosChangeJSON struct {
		NewTokenInfos *types.TokenInfos `json:"new_token_infos" yaml:"new_token_infos"`
	}

	// TokenInfosChangeProposalJSON defines a ParameterChangeProposal with a deposit used
	// to parse parameter change proposals from a JSON file.
	TokenInfosChangeProposalJSON struct {
		NewTokenInfos *types.TokenInfos `json:"new_token_infos" yaml:"new_token_infos"`
		Deposit       string            `json:"deposit" yaml:"deposit"`
	}

	// TokenInfosChangeProposalReq defines a parameter change proposal request body.
	TokenInfosChangeProposalReq struct {
		BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

		NewTokenInfos *types.TokenInfos `json:"new_token_infos" yaml:"new_token_infos"`
		Proposer      sdk.AccAddress    `json:"proposer" yaml:"proposer"`
		Deposit       sdk.Coins         `json:"deposit" yaml:"deposit"`
	}
)

func NewTokenInfosChangeJSON(tokenInfos *types.TokenInfos) TokenInfosChangeJSON {
	return TokenInfosChangeJSON{tokenInfos}
}

// ToTokenInfosChange converts a TokenInfosChangeJSON object to TokenInfosChange.
func (pcj TokenInfosChangeJSON) ToTokenInfosChange() types.TokenInfosChangeProposal {
	return *types.NewTokenInfosChangeProposal(pcj.NewTokenInfos)
}

// ToTokenInfosChanges converts a slice of TokenInfosChangeJSON objects to a slice of
// TokenInfosChange.
func (pcj TokenInfosChangesJSON) ToTokenInfosChanges() []types.TokenInfosChangeProposal {
	res := make([]types.TokenInfosChangeProposal, len(pcj))
	for i, pc := range pcj {
		res[i] = pc.ToTokenInfosChange()
	}
	return res
}

// ParseTokenInfosChangeProposalJSON reads and parses a TokenInfosChangeProposalJSON from
// file.
func ParseTokenInfosChangeProposalJSON(cdc *codec.LegacyAmino, proposalFile string) (TokenInfosChangeProposalJSON, error) {
	proposal := TokenInfosChangeProposalJSON{}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err := cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
