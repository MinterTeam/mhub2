package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	mhub2 "github.com/MinterTeam/mhub2/module/x/mhub2/types"
	"github.com/MinterTeam/mhub2/module/x/oracle/types"
	"github.com/MinterTeam/mhub2/oracle/config"
	"github.com/MinterTeam/mhub2/oracle/cosmos"
	"github.com/MinterTeam/mhub2/oracle/services/ethereum/gasprice"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

const usdteCoinId = 1993

var pipInBip = sdk.NewInt(1000000000000000000)

func main() {
	logger := log.NewTMLogger(os.Stdout)

	cfg := config.Get()

	cosmos.Setup(cfg)

	orcAddress, orcPriv := cosmos.GetAccount(cfg.Cosmos.Mnemonic)

	logger.Info("Orc address", "address", orcAddress.String())

	minterClient, err := http_client.New(cfg.Minter.ApiAddr)
	if err != nil {
		panic(err)
	}

	cosmosConn, err := grpc.DialContext(context.Background(), cfg.Cosmos.GrpcAddr, grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 5,
	}))

	if err != nil {
		panic(err)
	}
	defer cosmosConn.Close()

	ethGasPrice, err := gasprice.NewService(cfg, logger)

	if err != nil {
		panic(err)
	}

	for {
		relayPrices(cfg, minterClient, ethGasPrice, cosmosConn, orcAddress, orcPriv, logger)

		time.Sleep(1 * time.Second)
	}
}

func relayPrices(
	cfg *config.Config,
	minterClient *http_client.Client,
	ethGasPrice *gasprice.Service,
	cosmosConn *grpc.ClientConn,
	orcAddress sdk.AccAddress,
	orcPriv *secp256k1.PrivKey,
	logger log.Logger,
) {
	oracleClient := types.NewQueryClient(cosmosConn)
	mhub2Client := mhub2.NewQueryClient(cosmosConn)

	response, err := oracleClient.CurrentEpoch(context.Background(), &types.QueryCurrentEpochRequest{})
	if err != nil {
		logger.Error("Error getting current epoch", "err", err.Error())
		time.Sleep(time.Second)
		return
	}

	// check if already voted
	for _, vote := range response.GetEpoch().GetVotes() {
		if vote.Oracle == orcAddress.String() {
			return
		}
	}

	coins, err := mhub2Client.TokenInfos(context.Background(), &mhub2.TokenInfosRequest{})
	if err != nil {
		logger.Error("Error getting tokens list", "err", err.Error())
		time.Sleep(time.Second)
		return
	}

	prices := &types.Prices{List: []*types.Price{}}

	basecoinPrice := getBasecoinPrice(logger, minterClient)
	ethPrice := getEthPrice(logger)

	for _, coin := range coins.List.TokenInfos {
		if coin.ChainId != "minter" {
			continue
		}

		if coin.ExternalTokenId == "0" {
			prices.List = append(prices.List, &types.Price{
				Name:  coin.Denom,
				Value: basecoinPrice,
			})

			continue
		}

		switch coin.Denom {
		case "usdt", "usdc", "busd", "dai", "ust", "pax", "tusd", "husd":
			prices.List = append(prices.List, &types.Price{
				Name:  coin.Denom,
				Value: sdk.NewDec(1),
			})
		case "wbtc":
			prices.List = append(prices.List, &types.Price{
				Name:  coin.Denom,
				Value: getBitcoinPrice(logger),
			})
		case "weth":
			prices.List = append(prices.List, &types.Price{
				Name:  coin.Denom,
				Value: ethPrice,
			})
		case "bnb":
			prices.List = append(prices.List, &types.Price{
				Name:  coin.Denom,
				Value: getBnbPrice(logger),
			})
		case "oneinch":
			prices.List = append(prices.List, &types.Price{
				Name:  coin.Denom,
				Value: get1inchPrice(logger),
			})
		default:
			var route []uint64
			if coin.Denom == "hubabuba" {
				const hubCoinId = 1902
				route = []uint64{hubCoinId}
			}

			minterId, _ := strconv.Atoi(coin.ExternalTokenId)
			response, err := minterClient.EstimateCoinIDSellExtended(0, uint64(minterId), pipInBip.String(), 0, "optimal", route)
			if err != nil {
				_, payload, err := http_client.ErrorBody(err)
				if err != nil {
					logger.Error("Error estimating coin sell", "coin", coin.Denom, "err", err.Error())
				} else {
					logger.Error("Error estimating coin sell", "coin", coin.Denom, "err", payload.Error.Message)
				}

				time.Sleep(time.Second)
				return
			}

			priceInBasecoin, _ := sdk.NewDecFromStr(response.WillGet)
			price := priceInBasecoin.Mul(basecoinPrice).QuoInt(pipInBip)

			prices.List = append(prices.List, &types.Price{
				Name:  coin.Denom,
				Value: price,
			})
		}
	}

	prices.List = append(prices.List, &types.Price{
		Name:  "eth",
		Value: ethPrice,
	})

	prices.List = append(prices.List, &types.Price{
		Name:  "eth/gas",
		Value: ethGasPrice.GetGasPrice().Fast,
	})

	holders := getHolders(cfg)

	jsonPrices, _ := json.Marshal(prices.List)
	logger.Info("Prices", "val", string(jsonPrices))

	jsonHolders, _ := json.Marshal(holders.List)
	logger.Info("Holders", "val", string(jsonHolders))

	priceClaim := &types.MsgPriceClaim{
		Epoch:        response.Epoch.Nonce,
		Prices:       prices,
		Orchestrator: orcAddress.String(),
	}

	holdersClaim := &types.MsgHoldersClaim{
		Epoch:        response.Epoch.Nonce,
		Holders:      holders,
		Orchestrator: orcAddress.String(),
	}

	cosmos.SendCosmosTx([]sdk.Msg{priceClaim, holdersClaim}, orcAddress, orcPriv, cosmosConn, logger)
}

func getHolders(cfg *config.Config) *types.Holders {
	holders := &types.Holders{List: []*types.Holder{}}

	if cfg.HoldersUrl == "" {
		println("Holders url is not set")
		return holders
	}

	holdersResponse, err := http.Get(cfg.HoldersUrl)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(holdersResponse.Body)
	if err != nil {
		panic(err)
	}

	holdersList := HoldersResult{}
	if err := json.Unmarshal(data, &holdersList); err != nil {
		panic(err)
	}

	for _, item := range holdersList.Data {
		v, ok := sdk.NewIntFromString(item.Balance)
		if !ok {
			panic("err: " + item.Balance + " is not a number")
		}
		holders.List = append(holders.List, &types.Holder{
			Address: strings.ToLower(item.Address),
			Value:   v,
		})
	}

	return holders
}

func getBasecoinPrice(logger log.Logger, client *http_client.Client) sdk.Dec {
	return sdk.NewDecWithPrec(1, 1) // todo
	response, err := client.EstimateCoinIDSell(usdteCoinId, 0, pipInBip.String(), 0)
	if err != nil {
		_, payload, err := http_client.ErrorBody(err)
		if err != nil {
			logger.Error("Error estimating coin sell", "coin", "basecoin", "err", err.Error())
		} else {
			logger.Error("Error estimating coin sell", "coin", "basecoin", "err", payload.Error.Message)
		}

		time.Sleep(time.Second)

		return getBasecoinPrice(logger, client)
	}

	price, _ := sdk.NewIntFromString(response.WillGet)

	return price.ToDec().QuoInt(pipInBip)
}

func getEthPrice(logger log.Logger) sdk.Dec {
	_, body, err := fasthttp.Get(nil, "https://api.coingecko.com/api/v3/simple/price?ids=ethereum&vs_currencies=usd")
	if err != nil {
		logger.Error("Error getting eth price", "err", err.Error())
		time.Sleep(time.Second)
		return getEthPrice(logger)
	}
	var result CoingeckoResult
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Error getting eth price", "err", err.Error())
		time.Sleep(time.Second)
		return getEthPrice(logger)
	}

	d, err := sdk.NewDecFromStr(fmt.Sprintf("%.10f", result["ethereum"]["usd"]))
	if err != nil {
		panic(err)
	}

	return d
}

func getBitcoinPrice(logger log.Logger) sdk.Dec {
	_, body, err := fasthttp.Get(nil, "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		logger.Error("Error getting btc price", "err", err.Error())
		time.Sleep(time.Second)
		return getBitcoinPrice(logger)
	}
	var result CoingeckoResult
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Error getting btc price", "err", err.Error())
		time.Sleep(time.Second)
		return getBitcoinPrice(logger)
	}

	d, err := sdk.NewDecFromStr(fmt.Sprintf("%.10f", result["bitcoin"]["usd"]))
	if err != nil {
		panic(err)
	}

	return d
}

func getBnbPrice(logger log.Logger) sdk.Dec {
	_, body, err := fasthttp.Get(nil, "https://api.coingecko.com/api/v3/simple/price?ids=binancecoin&vs_currencies=usd")
	if err != nil {
		logger.Error("Error getting bnb price", "err", err.Error())
		time.Sleep(time.Second)
		return getBnbPrice(logger)
	}
	var result CoingeckoResult
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Error getting bnb price", "err", err.Error())
		time.Sleep(time.Second)
		return getBnbPrice(logger)
	}

	d, err := sdk.NewDecFromStr(fmt.Sprintf("%.10f", result["binancecoin"]["usd"]))
	if err != nil {
		panic(err)
	}

	return d
}

func get1inchPrice(logger log.Logger) sdk.Dec {
	_, body, err := fasthttp.Get(nil, "https://api.coingecko.com/api/v3/simple/price?ids=1inch&vs_currencies=usd")
	if err != nil {
		logger.Error("Error getting 1inch price", "err", err.Error())
		time.Sleep(time.Second)
		return get1inchPrice(logger)
	}
	var result CoingeckoResult
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Error getting 1inch price", "err", err.Error())
		time.Sleep(time.Second)
		return get1inchPrice(logger)
	}

	d, err := sdk.NewDecFromStr(fmt.Sprintf("%.10f", result["1inch"]["usd"]))
	if err != nil {
		panic(err)
	}

	return d
}

type HoldersResult struct {
	Data []struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
	} `json:"data"`
}

type CoingeckoResult map[string]map[string]float64
