package main

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"log"
	"net/http"
)

var EthFeeResult = sdk.NewInt(1e18)

type result struct {
	Result string `json:"result"`
}

func runEthFeeCalculatorServer() {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			chainId := r.URL.Query().Get("chain_id")
			contract := r.URL.Query().Get("contract")
			value, _ := sdk.NewIntFromString(r.URL.Query().Get("value"))

			_, _, _ = chainId, contract, value.String()

			data, err := json.Marshal(result{Result: EthFeeResult.String()})
			if err != nil {
				panic(err)
			}
			w.Write(data)
		})
		log.Fatal(http.ListenAndServe(":8840", nil))
	}()
}
