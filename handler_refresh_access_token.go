package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Tones12/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshAccessToken(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		msg := fmt.Sprintf("error getting bearer token: %s", err)
		respondWithError(w, 400, msg)
		return
	}

	dbRefreshToken, err := cfg.db.GetUserByRefreshToken(req.Context(), refreshToken)
	if err != nil {
		msg := fmt.Sprintf("error user refresh token not found: %s", err)
		respondWithError(w, 401, msg)
		return
	}

	if dbRefreshToken.RevokedAt.Valid {
		respondWithError(w, 401, "error: user refresh token has been revoked")
		return
	}
	if dbRefreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "error: user refresh token has expired")
		return
	}
	
	newAccessToken, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.secret, (time.Duration(1)*time.Hour))
	if err != nil {
		msg := fmt.Sprintf("Error getting new access token: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	type newAccToken struct {
		Token			string		`json:"token"`
	}
	newAccessTokenResponse := newAccToken{
		Token: newAccessToken,
	}
	respondWithJSON(w, 200, newAccessTokenResponse)
}