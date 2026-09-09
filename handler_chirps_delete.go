package main

import (
	"fmt"
	"net/http"
	
	"github.com/Tones12/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		msg := fmt.Sprintf("error getting bearer token: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		msg := fmt.Sprintf("error validating user: %s", err)
		respondWithError(w, 403, msg)
		return
	}
	
	parsedChirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		msg := fmt.Sprintf("Error parsing chirp id: %s", err)
		respondWithError(w, 400, msg)
		return
	}
	chirp, err := cfg.db.GetChirp(req.Context(), parsedChirpID)
	if err != nil {
		msg := fmt.Sprintf("Error getting chirp: %s", err)
		respondWithError(w, 404, msg)
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, 403, "Unauthorized user, cannot delete chirp")
		return
	}

	err = cfg.db.DeleteChirp(req.Context(), parsedChirpID)
	if err != nil {
		msg := fmt.Sprintf("Error deleting chirp: %s", err)
		respondWithError(w, 404, msg)
		return
	}
	
	w.WriteHeader(204)
}