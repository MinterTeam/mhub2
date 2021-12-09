package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeColdStorageTransfer defines the type for a ColdStorageTransferProposal
	ProposalTypeColdStorageTransfer = "ColdStorageTransfer"
	ProposalTypeTokenInfosChange    = "TokenInfosChange"
)

// Assert ColdStorageTransferProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &ColdStorageTransferProposal{}
var _ govtypes.Content = &TokenInfosChangeProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeColdStorageTransfer)
	govtypes.RegisterProposalType(ProposalTypeTokenInfosChange)
	govtypes.RegisterProposalTypeCodec(&ColdStorageTransferProposal{}, "mhub2/ColdStorageTransferProposal")
	govtypes.RegisterProposalTypeCodec(&TokenInfosChangeProposal{}, "mhub2/TokenInfosChangeProposal")
}

func NewColdStorageTransferProposal(chainId ChainID, amount sdk.Coins) *ColdStorageTransferProposal {
	return &ColdStorageTransferProposal{chainId.String(), amount}
}

func NewTokenInfosChangeProposal(tokenInfos *TokenInfos) *TokenInfosChangeProposal {
	return &TokenInfosChangeProposal{NewInfos: tokenInfos}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *ColdStorageTransferProposal) GetTitle() string { return "ColdStorageTransferProposal" }

// GetDescription returns the description of a community pool spend proposal.
func (csp *ColdStorageTransferProposal) GetDescription() string { return "ColdStorageTransferProposal" }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *ColdStorageTransferProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *ColdStorageTransferProposal) ProposalType() string { return ProposalTypeColdStorageTransfer }

// ValidateBasic runs basic stateless validity checks
func (csp *ColdStorageTransferProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (csp ColdStorageTransferProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Cold Storage Transfer Proposal:
  Amount:      %s`, csp.Amount))
	return b.String()
}

func (tic *TokenInfosChangeProposal) GetTitle() string { return "TokenInfosChangeProposal" }

func (tic *TokenInfosChangeProposal) GetDescription() string { return "TokenInfosChangeProposal" }

func (tic *TokenInfosChangeProposal) ProposalRoute() string { return RouterKey }

func (tic *TokenInfosChangeProposal) ProposalType() string { return ProposalTypeTokenInfosChange }

func (tic *TokenInfosChangeProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(tic)
	if err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (tic TokenInfosChangeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Token Infos Change Proposal:
  New Tokens:      %s`, tic.NewInfos))
	return b.String()
}
