package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Tones12/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

type errReturnVals struct {
	Error string `json:"error"`
}

type parameters struct {
	Body   string `json:"body"`
	Email  string `json:"email"`
	UserID string `json:"user_id"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func MapDBUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
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

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter) {
	cfg.fileserverHits.Swap(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respBody := errReturnVals{
		Error: msg,
	}
	respondWithJSON(w, code, respBody)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
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

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
		return
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
	}

	mux := http.NewServeMux()

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fileserverHits := apiCfg.fileserverHits.Load()
		text := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", fileserverHits)
		w.Write([]byte(text))
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, req *http.Request) {
		if apiCfg.platform != "dev" {
			respondWithError(w, 403, "Forbidden")
			return
		}
		apiCfg.resetMetrics(w)
		err = apiCfg.db.DeleteUsers(req.Context())
	})

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			msg := fmt.Sprintf("Error decoding JSON: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		user, err := apiCfg.db.CreateUser(req.Context(), params.Email)
		if err != nil {
			msg := fmt.Sprintf("Error creating User: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		userResponse := MapDBUserToUser(user)

		respondWithJSON(w, 201, userResponse)
	})

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			msg := fmt.Sprintf("Error decoding JSON: %s", err)
			respondWithError(w, 500, msg)
			return
		}

		// Check if user id is valid
		parsedUserID, err := uuid.Parse(params.UserID)
		if err != nil {
			msg := fmt.Sprintf("Error parsing user id: %s", err)
			respondWithError(w, 400, msg)
			return
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
			UserID: parsedUserID,
		}

		//Database insert
		dbChirp, err := apiCfg.db.CreateChirp(req.Context(), chirpParams)
		if err != nil {
			msg := fmt.Sprintf("Error creating chirp: %s", err)
			respondWithError(w, 400, msg)
			return
		}

		//Convert to JSON chirp struct
		chirp := MapDBChirpToChirp(dbChirp)

		respondWithJSON(w, 201, chirp)
	})

	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, req *http.Request) {
		dbChirpsList, err := apiCfg.db.ListChirps(req.Context())
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
	})

	mux.HandleFunc("GET /api/chirps/{chirpID}", func(w http.ResponseWriter, req *http.Request) {
		parsedChirpID, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			msg := fmt.Sprintf("Error parsing chirp id: %s", err)
			respondWithError(w, 400, msg)
			return
		}
		dbChirp, err := apiCfg.db.GetChirp(req.Context(), parsedChirpID)
		if err != nil {
			msg := fmt.Sprintf("Error finding chirp: %s", err)
			respondWithError(w, 404, msg)
			return
		}
		chirp := MapDBChirpToChirp(dbChirp)
		respondWithJSON(w, 200, chirp)
	})

	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Error: %s", err)
	}
}
