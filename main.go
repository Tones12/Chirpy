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
	"github.com/tones12/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
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
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Error: %s", err)
	}
	dbQueries := database.New(db)

	var apiCfg apiConfig
	
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
	
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	
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

	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Error: %s", err)
	}
}
