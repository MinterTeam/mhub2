package ethgasstation

import (
	"encoding/json"

	"github.com/MinterTeam/mhub2/oracle/services/ethereum/gasprice/providers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/valyala/fasthttp"
)

type result struct {
	Fast int64 `json:"fast"`
}

type Provider struct {
}

func New() *Provider {
	return &Provider{}
}

func (p Provider) Name() string {
	return "ethgasstation"
}

func (p *Provider) GetGasPrice() (*providers.GasPrice, error) {
	_, body, err := fasthttp.Get(nil, "https://ethgasstation.info/api/ethgasAPI.json")

	if err != nil {
		return nil, err
	}

	var result result

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &providers.GasPrice{
		Fast: sdk.NewDec(result.Fast).QuoInt64(10),
	}, nil
}
