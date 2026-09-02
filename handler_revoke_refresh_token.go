package main

import (
	"fmt"
	"net/http"

	"github.com/Tones12/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		msg := fmt.Sprintf("error getting bearer token: %s", err)
		respondWithError(w, 400, msg)
		return
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), refreshToken)
	if err != nil {
		msg := fmt.Sprintf("Error revoking refresh token: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	
	w.WriteHeader(204)
}