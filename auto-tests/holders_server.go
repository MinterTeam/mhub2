package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type HoldersResult struct {
	Data []struct {
		Address string `json:"address"`
		Balance string `json:"balance"`
	} `json:"data"`
}

var HoldersList HoldersResult

type PricesResult struct {
	Data []struct {
		Denom string `json:"denom"`
		Price string `json:"price"`
	} `json:"data"`
}

var PricesList = PricesResult{Data: []struct {
	Denom string `json:"denom"`
	Price string `json:"price"`
}{
	{"eth", "4000.0"},
	{"ethereum/gas", "123"},
	{"bnb", "400"},
	{"bsc/gas", "5"},
	{"hub", "0.1"},
}}

func runHoldersServer() {
	go func() {
		http.HandleFunc("/holders", func(w http.ResponseWriter, r *http.Request) {
			data, err := json.Marshal(HoldersList)
			if err != nil {
				panic(err)
			}
			w.Write(data)
		})

		http.HandleFunc("/prices", func(w http.ResponseWriter, r *http.Request) {
			data, err := json.Marshal(PricesList)
			if err != nil {
				panic(err)
			}
			w.Write(data)
		})
		log.Fatal(http.ListenAndServe(":8844", nil))
	}()
}
