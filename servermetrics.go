package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

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
