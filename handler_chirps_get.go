package main

import (
	"fmt"
	"net/http"
	"github.com/google/uuid"
)


func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, req *http.Request) {
	parsedChirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		msg := fmt.Sprintf("Error parsing chirp id: %s", err)
		respondWithError(w, 400, msg)
		return
	}
	dbChirp, err := cfg.db.GetChirp(req.Context(), parsedChirpID)
	if err != nil {
		msg := fmt.Sprintf("Error finding chirp: %s", err)
		respondWithError(w, 404, msg)
		return
	}
	chirp := MapDBChirpToChirp(dbChirp)
	respondWithJSON(w, 200, chirp)

}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, req *http.Request) {
	dbChirpsList, err := cfg.db.ListChirps(req.Context())
	if err != nil {
		msg := fmt.Sprintf("Error getting chirps: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	chirpsList := []Chirp{}
	for _, dbChirp := range dbChirpsList {
		chirp := MapDBChirpToChirp(dbChirp)
		chirpsList = append(chirpsList, chirp)
	}
	respondWithJSON(w, 200, chirpsList)

}