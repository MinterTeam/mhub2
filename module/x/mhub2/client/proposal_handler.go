package client

import (
	"github.com/MinterTeam/mhub2/module/x/mhub2/client/cli"
	"github.com/MinterTeam/mhub2/module/x/mhub2/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// ProposalColdStorageHandler is the param change proposal handler.
var ProposalColdStorageHandler = govclient.NewProposalHandler(cli.NewSubmitColdStorageTransferProposalTxCmd, rest.ColdStorageTransferProposalRESTHandler)
var ProposalTokensChangeHandler = govclient.NewProposalHandler(cli.NewSubmitTokenInfosChangeProposalTxCmd, rest.TokenInfosChangeProposalRESTHandler)
