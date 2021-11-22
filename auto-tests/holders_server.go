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
		Denom string `json:"address"`
		Price string `json:"balance"`
	} `json:"data"`
}

var PricesList = PricesResult{Data: []struct {
	Denom string `json:"address"`
	Price string `json:"balance"`
}{
	{"eth", "4000.0"},
	{"eth/gas", "123"},
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
