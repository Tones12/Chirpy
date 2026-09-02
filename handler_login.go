package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Tones12/Chirpy/internal/auth"
	"github.com/Tones12/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email				string `json:"email"`
		Password			string `json:"password"`
	}
	type userTokens struct {
		ID				uuid.UUID 	`json:"id"`
		CreatedAt		time.Time 	`json:"created_at"`
		UpdatedAt		time.Time 	`json:"updated_at"`
		Email			string    	`json:"email"`
		Token			string		`json:"token"`
		RefreshToken	string		`json:"refresh_token"`
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
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
	
	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, (time.Duration(1)*time.Hour))
	if err != nil {
		msg := fmt.Sprintf("Error getting access token: %s", err)
		respondWithError(w, 500, msg)
		return
	}

	refreshToken := auth.MakeRefreshToken()
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: dbUser.ID,
	}
	_, err = cfg.db.CreateRefreshToken(req.Context(), refreshTokenParams)
	if err != nil {
		msg := fmt.Sprintf("Error getting refresh token: %s", err)
		respondWithError(w, 500, msg)
		return
	}
	userResponse := userTokens{
		ID:				dbUser.ID,
		CreatedAt:		dbUser.CreatedAt,
		UpdatedAt:		dbUser.UpdatedAt,	
		Email:			dbUser.Email,
		Token:			token,
		RefreshToken:	refreshToken,
	}
	respondWithJSON(w, 200, userResponse)
}