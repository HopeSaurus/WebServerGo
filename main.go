package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = atomic.Int32{}
	w.WriteHeader(200)
}

func (cfg *apiConfig) getAPIMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	fmt.Fprintf(w, `<html>
  								<body>
    							<h1>Welcome, Chirpy Admin</h1>
    							<p>Chirpy has been visited %d times!</p>
  				</body>
				</html>`, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Payload string `json:"body"`
	}
	type responseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(req.Body)
	data := requestBody{}
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if stringLength := len(data.Payload); stringLength > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	respondWithJSON(w, 200, responseBody{CleanedBody: removeProfanity(data.Payload)})
}

func main() {
	serverMux := http.NewServeMux()
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	cfg := apiConfig{fileserverHits: atomic.Int32{}}

	// Remember to add trailing slash to match everything that has app in the pathname
	serverMux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	serverMux.HandleFunc("GET /admin/metrics", cfg.getAPIMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.resetMetrics)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirp)

	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	server.ListenAndServe()
}
