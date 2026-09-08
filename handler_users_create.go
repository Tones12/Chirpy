package main

import (
	"encoding/json"
	"net/http"
	"fmt"

	"github.com/Tones12/Chirpy/internal/auth"
	"github.com/Tones12/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
		type parameters struct {
			Email		string `json:"email"`
			Password	string `json:"password"`
		}
		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			msg := fmt.Sprintf("Error decoding JSON: %s", err)
			respondWithError(w, 500, msg)
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
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			msg := fmt.Sprintf("Error hashing password: %s", err)
			respondWithError(w, 500, msg)
			return
		}

		userInput := database.CreateUserParams{Email: params.Email, HashedPasswords: hashedPassword}

		user, err := cfg.db.CreateUser(req.Context(), userInput)
		if err != nil {
			msg := fmt.Sprintf("Error creating User: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		userResponse := MapDBUserToUser(user) // no password field when mapping to user so not sent out

		respondWithJSON(w, 201, userResponse)
	}