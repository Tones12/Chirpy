package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"regexp"
	"fmt"

	"github.com/Tones12/Chirpy/internal/database"
	"github.com/Tones12/Chirpy/internal/auth"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func MapDBChirpToChirp(dbChirp database.Chirp) Chirp {
	return Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, req *http.Request) {
		type parameters struct {
			Body	string `json:"body"`
		}

		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			msg := fmt.Sprintf("error decoding JSON: %s", err)
			respondWithError(w, 500, msg)
			return
		}

		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			msg := fmt.Sprintf("error getting bearer token: %s", err)
			respondWithError(w, 400, msg)
		}
		userID, err := auth.ValidateJWT(token, cfg.secret)
		if err != nil {
			msg := fmt.Sprintf("error validating user: %s", err)
			respondWithError(w, 401, msg)
		}

		// Check chirp length
		if len(params.Body) > 140 {
			respondWithError(w, 400, "Chirp is too long, must be 140 characters or less")
			return
		}
		// Check for profanities: kerfuffle, sharbert, fornax
		badWords := []string{"kerfuffle", "sharbert", "fornax"}
		dirtyText := params.Body
		cleanText := censorText(dirtyText, badWords)
		chirpParams := database.CreateChirpParams{
			Body:   cleanText,
			UserID: userID,
		}
		//Database insert
		dbChirp, err := cfg.db.CreateChirp(req.Context(), chirpParams)
		if err != nil {
			msg := fmt.Sprintf("Error creating chirp: %s", err)
			respondWithError(w, 400, msg)
			return
		}
		//Convert to JSON chirp struct
		chirp := MapDBChirpToChirp(dbChirp)
		respondWithJSON(w, 201, chirp)
}

func censorText(input string, badWords []string) string {
	if len(badWords) == 0 {
		return input
	}
	escapedWords := make([]string, len(badWords))
	for i, word := range badWords {
		escapedWords[i] = regexp.QuoteMeta(word)
	}
	pattern := `(?i)\b(` + strings.Join(escapedWords, "|") + `)\b`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(input, "****")
}