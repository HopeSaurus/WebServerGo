package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Email string `json:"email"`
	}
	data := requestBody{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, 500, "The server couldnt decode the request")
		return
	}
	if data.Email == "" {
		respondWithError(w, 400, "Bad request")
		return
	}
	user, err := cfg.db.CreateUser(req.Context(), data.Email)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed at creating the user: %s", err))
		return
	}
	response := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, 201, response)
}

func (cfg *apiConfig) deleteAllUsers(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Nope")
		return
	}
	err := cfg.db.DeleteUsers(req.Context())
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Db error: %s", err))
		return
	}
	w.WriteHeader(204)
	w.Write([]byte{})
}
