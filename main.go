package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()
	// mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	apiCfg := apiConfig{}
	mux.HandleFunc("/healthz", handlerReadiness)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/metrics", apiCfg.handlerName)
	mux.HandleFunc("/reset", apiCfg.reset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux, // a hypothetical ServeMux or custom handler
	}

	server.ListenAndServe()

}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits.Load())
}
