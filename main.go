package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/Orujhimaru/chirpy-go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	mux := http.NewServeMux()
	// mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
	}

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("POST /admin/reset", apiCfg.reset)

	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getSingleChirp)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux, // a hypothetical ServeMux or custom handler
	}

	server.ListenAndServe()

}

func cleanUp(body string) string {
	result := ""
	normalArr := strings.Split(body, " ")
	splitArr := strings.Split(strings.ToLower(body), " ")
	for i, str := range splitArr {
		if str != "kerfuffle" && str != "sharbert" && str != "fornax" {
			if i != 0 {
				result += fmt.Sprintf(" %s", normalArr[i])
			} else {
				result += normalArr[i]
			}

		} else {
			result += " ****"
		}
	}
	return result
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>
</html>
	`, cfg.fileserverHits.Load())))
}

//	func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
//		w.WriteHeader(http.StatusOK)
//		fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits.Load())
//	}
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

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	err := cfg.db.Reset(r.Context())
	if err != nil {
		fmt.Printf("Reset error: %v\n", err)
		return
	}
	// rows, _ := result.RowsAffected()
	// fmt.Printf("Reset deleted %d rows\n", rows)
	// ... rest of your handler
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
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
