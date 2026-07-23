package main

import (
	"net/http"
	"fmt"
	"sync/atomic"
	"strconv"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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

func main() {
	var apiCfg apiConfig
	
	mux := http.NewServeMux()

	server := &http.Server{
		Handler:	mux,
		Addr:		":8080",
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fileserverHitsStatement := "Hits: " + strconv.FormatInt(int64(apiCfg.fileserverHits.Load()), 10)
		w.Write([]byte(fileserverHitsStatement))
	})
	mux.HandleFunc("/reset", apiCfg.resetMetrics)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	
}
