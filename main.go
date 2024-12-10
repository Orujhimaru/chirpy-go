package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()
	// mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	apiCfg := apiConfig{}

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("POST /admin/reset", apiCfg.reset)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.validate)
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

func (cfg *apiConfig) validate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type returnVals struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		Text string `json:"body"`
	}
	type cleanUpStr struct {
		Text string `json:"cleaned_body"`
	}

	type returnValid struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		Text bool `json:"valid"`
	}
	type returnError struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		Text string `json:"error"`
	}
	error := returnError{Text: "Something went wrong"}
	lenError := returnError{Text: "Chirp is too long"}
	// validResponse := returnValid{Text: true}

	decoder := json.NewDecoder(r.Body)
	params := returnVals{}
	err := decoder.Decode(&params)
	if err != nil {
		data, err := json.Marshal(error)
		if err != nil {
			// w.Write(data)
		}
		w.Write(data)
		// w.Write()
	}
	if len(params.Text) > 140 {
		w.WriteHeader(400)
		data, err := json.Marshal(lenError)
		if err != nil {

		}
		w.Write(data)

	} else {
		w.WriteHeader(200)
		data, err := json.Marshal(cleanUpStr{Text: cleanUp(params.Text)})
		if err != nil {

		}
		w.Write(data)
	}

	// w.Write(json.Marshall(params))
}

type apiConfig struct {
	fileserverHits atomic.Int32
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Hits: %v", cfg.fileserverHits.Load())
}
