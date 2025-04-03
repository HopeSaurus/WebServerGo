package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Hopesaurus/WebServerGo/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Payload string    `json:"body"`
		UserID  uuid.UUID `json:"user_id"`
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
	cleansedBody := removeProfanity(data.Payload)

	params := database.CreateChirpParams{
		Body:   cleansedBody,
		UserID: data.UserID,
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), params)

	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Couldnt create chirp %s", err))
	}

	responseChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	respondWithJSON(w, 200, responseChirp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, req *http.Request) {
	data, err := cfg.db.GetChirps(req.Context())
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Something went wrong: %s", err))
	}
	chirpSlice := []Chirp{}

	for _, item := range data {
		chirp := Chirp{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			Body:      item.Body,
			UserId:    item.UserID,
		}
		chirpSlice = append(chirpSlice, chirp)
	}
	respondWithJSON(w, 200, chirpSlice)
}
