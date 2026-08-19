package main

import (
	"encoding/json"
	"net/http"
	"fmt"

	"github.com/Tones12/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email				string `json:"email"`
		Password			string `json:"password"`
		ExpiryTimeinSeconds	int `json:"expires_in_seconds"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{
		ExpiryTimeinSeconds: 3600,
	}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding JSON: %s", err)
		respondWithError(w, 400, msg)
		return
	}
	if params.Email == "" {
		respondWithError(w, 400, "Error, email missing")
		return
	}
	if params.Password == "" {
		respondWithError(w, 400, "Error, password missing")
		return
	}
	if params.ExpiryTimeinSeconds <= 0 || params.ExpiryTimeinSeconds > 3600 {
		params.ExpiryTimeinSeconds = 3600
	}
	dbUser, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		msg := fmt.Sprintf("incorrect email or password: %s", err)
		respondWithError(w, 401, msg)
		return
	}
	ok, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPasswords) //need to look up stored password based on email
	if err != nil {
		msg := fmt.Sprintf("Error checking password: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	if !ok {
		respondWithError(w, 401, "incorrect email or password")
		return
	}
	userResponse := MapDBUserToUser(dbUser)
	respondWithJSON(w, 200, userResponse)
}