package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", handlerReadiness)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux, // a hypothetical ServeMux or custom handler
	}

	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("/asseets/logo.png", http.FileServer(http.Dir(".")))

	server.ListenAndServe()
}
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
