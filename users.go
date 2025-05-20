package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Hopesaurus/WebServerGo/internal/auth"
	"github.com/Hopesaurus/WebServerGo/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitifempty"`
	RefreshToken string    `json:"refresh_token,omitifempty"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	data := requestBody{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, 500, "The server couldnt decode the request")
		return
	}
	if data.Email == "" || data.Password == "" {
		respondWithError(w, 400, "Bad request")
		return
	}

	hashedPass, err := auth.HashPassword(data.Password)

	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed hashing password: %s", err))
	}

	params := database.CreateUserParams{
		Email:          data.Email,
		HashedPassword: hashedPass,
	}

	user, err := cfg.db.CreateUser(req.Context(), params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("The email is already registered"))
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

func (cfg *apiConfig) login(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	data := requestBody{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, 500, "The server couldnt decode the request")
		return
	}
	if data.Email == "" || data.Password == "" {
		respondWithError(w, 400, "Bad request")
		return
	}
	user, err := cfg.db.GetUser(req.Context(), data.Email)

	if err != nil {
		respondWithError(w, 503, "User not found")
		return
	}

	err = auth.CheckPasswordHash(user.HashedPassword, data.Password)

	if err != nil {
		respondWithError(w, 503, fmt.Sprintf("%s", err))
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(3600*int(time.Second)))
	if err != nil {
		respondWithError(w, 503, fmt.Sprintf("%s", err))
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()

	params := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(60 * 24 * int(time.Hour))),
	}

	dbResult, err := cfg.db.CreateRefreshToken(req.Context(), params)
	if err != nil {
		respondWithError(w, 503, fmt.Sprintf("%s", err))
		return
	}

	response := User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: dbResult.Token,
	}
	respondWithJSON(w, 200, response)
}

func (cfg *apiConfig) UpdateUser(w http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	ctx := req.Context()
	userID := ctx.Value("userUUID").(uuid.UUID)
	data := requestBody{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, 500, "The server couldnt decode the request")
		return
	}
	if data.Email == "" || data.Password == "" {
		respondWithError(w, 400, "Bad request")
		return
	}
	hashedPass, err := auth.HashPassword(data.Password)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed hashing password: %s", err))
		return
	}
	userRow, err := cfg.db.UpdateUser(ctx, database.UpdateUserParams{
		ID:             userID,
		Email:          data.Email,
		HashedPassword: hashedPass,
	})
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Failed at creating the user: %s", err))
		return
	}
	response := User{
		ID:        userRow.ID,
		CreatedAt: userRow.CreatedAt,
		UpdatedAt: userRow.UpdatedAt,
		Email:     userRow.Email,
	}

	respondWithJSON(w, 200, response)
}
