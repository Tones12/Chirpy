package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/Tones12/Chirpy/internal/database"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type User struct {
	ID			uuid.UUID 	`json:"id"`
	CreatedAt	time.Time 	`json:"created_at"`
	UpdatedAt	time.Time 	`json:"updated_at"`
	Email		string    	`json:"email"`
	Password	string		`json:"-"`
}

func MapDBUserToUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}

type apiConfig struct {
	fileserverHits	atomic.Int32
	db				*database.Queries
	platform		string
	secret			string
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
		return
	}
	dbQueries := database.New(db)
	
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:			secret,
	}

	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshAccessToken)

	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeRefreshToken)

	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUsersUpdate)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsRetrieve)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGet)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerChirpsDelete)
	
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetricsfunc)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerUpgrade)

	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Error: %s", err)
	}
}