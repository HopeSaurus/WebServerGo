package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Hopesaurus/WebServerGo/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
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
	if data.Payload == "" {
		respondWithError(w, 400, "Bad request")
		return
	}
	if stringLength := len(data.Payload); stringLength > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	respondWithJSON(w, 200, responseBody{CleanedBody: removeProfanity(data.Payload)})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Cannot establish connection to the database: %s", err)
	}
	dbQueries := database.New(db)

	serverMux := http.NewServeMux()
	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}
	cfg := apiConfig{fileserverHits: atomic.Int32{}, db: dbQueries, platform: platform}

	// Remember to add trailing slash to match everything that has app in the pathname
	serverMux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	serverMux.HandleFunc("GET /admin/metrics", cfg.getAPIMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.deleteAllUsers)
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	serverMux.HandleFunc("POST /api/users", cfg.createUser)

	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	server.ListenAndServe()
}
