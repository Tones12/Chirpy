package main

import (
	"net/http"
	"fmt"
	"sync/atomic"
	"encoding/json"
	"log"
	"strings"
	"regexp"
	"os"
	"database/sql"
	"github.com/joho/godotenv"
	"time"
	"github.com/google/uuid"
	"github.com/Tones12/Chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
}

type errReturnVals struct {
	Error string `json:"error"`
}

type validReturnVals struct {
	CleanedBody string `json:"cleaned_body"`
}

type parameters struct {
	Body string `json:"body"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}


func MapDBUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}
	
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func(cfg *apiConfig) resetMetrics(w http.ResponseWriter, req *http.Request) {
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
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{
    	fileserverHits: atomic.Int32{},
    	db:             dbQueries,
		platform:		platform,
	}

	mux := http.NewServeMux()

	server := &http.Server{
		Handler:	mux,
		Addr:		":8080",
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fileserverHits := apiCfg.fileserverHits.Load()
		text := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", fileserverHits)
		w.Write([]byte(text))
	})
	
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, req *http.Request){
		if apiCfg.platform != "dev" {
			respondWithError(w, 403, "Forbidden")
		}
		apiCfg.resetMetrics(w, req)
		err = db.DeleteUsers(r.Context())
	})
	
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, req *http.Request){
		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			msg := fmt.Sprintf("Error decoding JSON: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		
		if len(params.Body) > 140 {
			respondWithError(w, 400, "Chirp is too long")
			return
		}

		// Check for profanities: kerfuffle, sharbert, fornax
		badWords := []string{"kerfuffle", "sharbert", "fornax"}
		dirtyText := params.Body

		cleanText := censorText(dirtyText, badWords)

		validResp := validReturnVals{
			CleanedBody: cleanText,
		}
		
		respondWithJSON(w, 200, validResp)
	})
	
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, req *http.Request){
		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			msg := fmt.Sprintf("Error decoding JSON: %s", err)
			respondWithError(w, 500, msg)
			return
		}
		user, err := CreateUser(r.Context(), params.Body)
		if err != nil {
			msg := fmt.Sprintf("Error creating User: %s", err)
			respondWithError(w, 500, msg)
			return
		}

		userResponse := MapDBUsertoUser(user)
		
		respondWithJSON(w, 200, userResponse)
	})

	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Error: %s", err)
	}
}