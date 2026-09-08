package main

import (
	"fmt"
	"net/http"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, req *http.Request) {
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