package main

import (
	"encoding/json"
	"log"
	"net/http"
)

 type errReturnVals struct {
	Error string
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respBody := errReturnVals{
		Error: msg,
	}
	respondWithJSON(w, code, respBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}