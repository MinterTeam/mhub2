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

func runHoldersServer() {
	go func() {
		http.HandleFunc("/holders", func(w http.ResponseWriter, r *http.Request) {
			data, err := json.Marshal(HoldersList)
			if err != nil {
				panic(err)
			}
			w.Write(data)
		})
		log.Fatal(http.ListenAndServe(":8844", nil))
	}()
}
