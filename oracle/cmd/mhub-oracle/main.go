package main

import (
	"context"
	"encoding/json"
	"github.com/MinterTeam/mhub2/minter-connector/tx_committer"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MinterTeam/mhub2/module/x/oracle/types"
	"github.com/MinterTeam/mhub2/oracle/config"
	"github.com/MinterTeam/mhub2/oracle/cosmos"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var holdersUpdatePeriod uint64 = 144

func main() {
	logger := log.NewTMLogger(os.Stdout)
	cfg, isTestnet := config.Get()
	cosmos.Setup(cfg)

	if isTestnet {
		holdersUpdatePeriod = 1
	}

	txCommitterConn, err := grpc.Dial("127.0.0.1:7070", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer txCommitterConn.Close()
	txCommitter := tx_committer.NewTxCommitterClient(txCommitterConn)

	orcAddress, err := txCommitter.Address(context.Background(), &tx_committer.AddressRequest{})
	if err != nil {
		panic(err)
	}

	logger.Info("Orc address", "address", orcAddress.GetAddress())

	cosmosConn, err := grpc.DialContext(context.Background(), cfg.Cosmos.GrpcAddr, grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 5,
	}))

	if err != nil {
		panic(err)
	}
	defer cosmosConn.Close()

	for {
		relayPricesAndHolders(cfg, orcAddress.GetAddress(), cosmosConn, txCommitter, logger)

		time.Sleep(1 * time.Second)
	}
}

func relayPricesAndHolders(cfg *config.Config, address string, cosmosConn *grpc.ClientConn, txCommitter tx_committer.TxCommitterClient, logger log.Logger) {
	orcAddress, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}

	oracleClient := types.NewQueryClient(cosmosConn)

	response, err := oracleClient.CurrentEpoch(context.Background(), &types.QueryCurrentEpochRequest{})
	if err != nil {
		logger.Error("Error getting current epoch", "err", err.Error())
		time.Sleep(time.Second)
		return
	}

	if response.GetEpoch().GetNonce() == 0 {
		return
	}

	// check if already voted
	for _, vote := range response.GetEpoch().GetVotes() {
		if vote.Oracle == orcAddress.String() {
			return
		}
	}

	if response.GetEpoch().Nonce%holdersUpdatePeriod == 0 {
		holders := getHolders(cfg)
		jsonHolders, _ := json.Marshal(holders.List)
		logger.Info("Holders", "val", string(jsonHolders))

		holdersClaim := &types.MsgHoldersClaim{
			Epoch:        response.Epoch.Nonce,
			Holders:      holders,
			Orchestrator: orcAddress.String(),
		}

		_, err = txCommitter.CommitTx(context.TODO(), &tx_committer.CommitTxRequest{Msgs: tx_committer.MarshalMsgs([]sdk.Msg{holdersClaim})})
		if err != nil {
			logger.Error("Failed to broadcast holders", "err", err.Error())
		}
	}

	prices := getPrices(cfg)
	jsonPrices, _ := json.Marshal(prices.List)
	logger.Info("Prices", "val", string(jsonPrices))

	priceClaim := &types.MsgPriceClaim{
		Epoch:        response.Epoch.Nonce,
		Prices:       prices,
		Orchestrator: orcAddress.String(),
	}

	_, err = txCommitter.CommitTx(context.TODO(), &tx_committer.CommitTxRequest{Msgs: tx_committer.MarshalMsgs([]sdk.Msg{priceClaim})})
	if err != nil {
		logger.Error("Failed to broadcast prices", "err", err.Error())
	}
}

func getHolders(cfg *config.Config) *types.Holders {
	holders := &types.Holders{List: []*types.Holder{}}

	if cfg.HoldersUrl == "" {
		println("Holders url is not set")
		return holders
	}

	holdersResponse, err := http.Get(cfg.HoldersUrl)
	if err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return getHolders(cfg)
	}
	data, err := ioutil.ReadAll(holdersResponse.Body)
	if err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return getHolders(cfg)
	}

	holdersList := HoldersResult{}
	if err := json.Unmarshal(data, &holdersList); err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return getHolders(cfg)
	}

	for _, item := range holdersList.Data {
		v, ok := sdk.NewIntFromString(item.Balance)
		if !ok {
			println("err: " + item.Balance + " is not a number")
			time.Sleep(time.Second)
			return getHolders(cfg)
		}

		// skip balances less than 1 HUB
		if v.LT(sdk.NewInt(1e18)) {
			continue
		}

		holders.List = append(holders.List, &types.Holder{
			Address: strings.ToLower(item.Address),
			Value:   v,
		})
	}

	if len(holders.List) > 2000 {
		holders.List = holders.List[:2000]
	}

	return holders
}

func getPrices(cfg *config.Config) *types.Prices {
	prices := &types.Prices{List: []*types.Price{}}

	if cfg.PricesUrl == "" {
		println("Prices url is not set")
		return prices
	}

	pricesResponse, err := http.Get(cfg.PricesUrl)
	if err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return getPrices(cfg)
	}
	data, err := ioutil.ReadAll(pricesResponse.Body)
	if err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return getPrices(cfg)
	}

	pricesList := PricesResult{}
	if err := json.Unmarshal(data, &pricesList); err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return getPrices(cfg)
	}

	for _, item := range pricesList.Data {
		v, err := sdk.NewDecFromStr(item.Price)
		if err != nil {
			println(err.Error())
			time.Sleep(time.Second)
			return getPrices(cfg)
		}
		prices.List = append(prices.List, &types.Price{
			Name:  item.Denom,
			Value: v,
		})
	}

	return prices
}

type HoldersResult struct {
	Data []struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
	} `json:"data"`
}

type PricesResult struct {
	Data []struct {
		Denom string `json:"denom"`
		Price string `json:"price"`
	} `json:"data"`
}
