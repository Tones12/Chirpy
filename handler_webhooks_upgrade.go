package main

import (
	"encoding/json"
	"net/http"
	"fmt"

	"github.com/Tones12/Chirpy/internal/auth"
	"github.com/Tones12/Chirpy/internal/database"	
)


func (cfg *apiConfig) handlerUpgrade(w http.ResponseWriter, req *http.Request) {
		type parameters struct {
			Event		string		`json:"event"`
			Data		EventData	`json:"data"`
		}
}

type EventData struct {
	UserID string `json:"user_id"`
}
		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			msg := fmt.Sprintf("Error decoding JSON: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			msg := fmt.Sprintf("Error hashing password: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			msg := fmt.Sprintf("Error getting bearer token: %s", err)
			respondWithError(w, 401, msg)
			return
		}
		userID, err := auth.ValidateJWT(token, cfg.secret)
		if err != nil {
			msg := fmt.Sprintf("Error validating token: %s", err)
			respondWithError(w, 401, msg)
			return
		}
		userParams := database.UpdateUserParams{
				Email: params.Email,
				HashedPasswords: hashedPassword,
				ID: userID,
		}
		var user database.UpdateUserRow

		user, err = cfg.db.UpdateUser(req.Context(), userParams)
		if err != nil {
			msg := fmt.Sprintf("Error updating User: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		userResponse := User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		}

		respondWithJSON(w, 200, userResponse)
	}