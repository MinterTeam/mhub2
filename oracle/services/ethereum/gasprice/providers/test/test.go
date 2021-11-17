package test

import (
	"github.com/MinterTeam/mhub2/oracle/services/ethereum/gasprice/providers"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p Provider) Name() string {
	return "test"
}

func (p *Provider) GetGasPrice() (*providers.GasPrice, error) {
	return &providers.GasPrice{
		Fast: sdk.NewDec(10),
	}, nil
}
