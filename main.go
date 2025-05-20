package main

import (
	"database/sql"
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
	secret         string
	polkaAPI       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET_KEY")
	polkaAPI := os.Getenv("POLKA_API")
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
	cfg := apiConfig{fileserverHits: atomic.Int32{},
		db:       dbQueries,
		platform: platform,
		secret:   secret,
		polkaAPI: polkaAPI}

	//An http handler is an interface that in
	// Remember to add trailing slash to match everything that has app in the pathname
	serverMux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	serverMux.HandleFunc("GET /admin/metrics", cfg.getAPIMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.deleteAllUsers)
	serverMux.HandleFunc("POST /api/users", cfg.createUser)
	serverMux.HandleFunc("GET /api/chirps", cfg.getChirps)
	serverMux.Handle("POST /api/chirps", cfg.validateJWTMiddleware(http.HandlerFunc(cfg.createChirp)))
	serverMux.HandleFunc("POST /api/login", cfg.login)
	serverMux.HandleFunc("POST /api/refresh", cfg.RefreshTokenHandler)
	serverMux.HandleFunc("POST /api/revoke", cfg.RevokeRefreshTokenHandler)
	serverMux.Handle("PUT /api/users", cfg.validateJWTMiddleware(http.HandlerFunc(cfg.UpdateUser)))
	serverMux.HandleFunc("POST /api/polka/webhooks", cfg.PolkaWebhookHandler)
	serverMux.Handle("DELETE /api/chirps/{chirpID}", cfg.validateJWTMiddleware(http.HandlerFunc(cfg.deleteChirp)))
	//Registers the handler given the path
	//HandlerFunc makes it so a function can act as http.handler
	serverMux.Handle("GET /api/chirps/{chirpID}", validateUUIDMiddleware([]string{"chirpID"}, http.HandlerFunc(cfg.getChirp)))

	//Registers the handler func given the path
	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	server.ListenAndServe()
}
