package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var HoldersList = map[string]string{}

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
