package client

import (
	"github.com/MinterTeam/mhub2/module/x/mhub2/client/cli"
	"github.com/MinterTeam/mhub2/module/x/mhub2/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

// ProposalHandler is the param change proposal handler.
var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitColdStorageTransferProposalTxCmd, rest.ProposalRESTHandler)
