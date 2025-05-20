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
		Payload string `json:"body"`
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

	userID, ok := req.Context().Value("userUUID").(uuid.UUID)
	if !ok {
		respondWithError(w, 400, "Error with the authentication")
	}

	params := database.CreateChirpParams{
		Body:   cleansedBody,
		UserID: userID,
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
	author := req.URL.Query().Get("author_id")
	var data []database.Chirp
	var err error
	if author != "" {
		authorUUID, err := uuid.Parse(author)
		if err != nil {
			respondWithError(w, 400, "Bad request")
			return
		}
		data, err = cfg.db.GetChirpsFromUser(req.Context(), authorUUID)
		if err != nil {
			respondWithError(w, 500, fmt.Sprintf("Something went wrong: %s", err))
			return
		}
	} else {
		data, err = cfg.db.GetChirps(req.Context())
		if err != nil {
			respondWithError(w, 500, fmt.Sprintf("Something went wrong: %s", err))
		}
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

func (cfg *apiConfig) getChirp(w http.ResponseWriter, req *http.Request) {
	userId := req.Context().Value("chirpID").(uuid.UUID)

	chirp, err := cfg.db.GetChirp(req.Context(), userId)
	if err != nil {
		respondWithError(w, 404, "Chirp Not Found")
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

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID := ctx.Value("userUUID").(uuid.UUID)

	chirpID := req.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, 400, "Bad request")
		return
	}
	chirp, err := cfg.db.GetChirp(ctx, chirpUUID)
	if err != nil {
		respondWithError(w, 404, "Chirp Not Found")
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, 403, "You are not authorized to delete this chirp")
		return
	}

	err = cfg.db.DeleteChirp(ctx, chirpUUID)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed at deleting the chirp: %s", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
