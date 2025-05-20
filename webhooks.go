package main

import (
	"encoding/json"
	"net/http"

	"github.com/Hopesaurus/WebServerGo/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) PolkaWebhookHandler(w http.ResponseWriter, req *http.Request) {
	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w, 403, "No API key given")
		return
	}
	if apiKey != cfg.polkaAPI {
		respondWithError(w, 403, "Invalid API key")
		return
	}
	// Get the request body
	type Data struct {
		UserId string `json:"user_id"`
	}
	type requestBody struct {
		Event string `json:"event"`
		Data  Data   `json:"data"`
	}
	decoder := json.NewDecoder(req.Body)
	data := requestBody{}
	err = decoder.Decode(&data)
	if err != nil {
		respondWithError(w, 500, "Error decoding the request")
		return
	}
	if data.Event == "" {
		respondWithError(w, 400, "Bad request")
		return
	}
	if data.Event == "user.upgraded" {
		// Get the user id from the request
		userId := data.Data.UserId
		if userId == "" {
			respondWithError(w, 400, "Bad request")
			return
		}
		// Do something with the user id
		userUUID, err := uuid.Parse(userId)
		if err != nil {
			respondWithError(w, 400, "Bad request")
			return
		}
		_, err = cfg.db.UpgradeUser(req.Context(), userUUID)
		if err != nil {
			respondWithError(w, 404, "User not found")
			return
		}
		respondWithJSON(w, http.StatusNoContent, "")
	}
}
