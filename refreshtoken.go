package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Hopesaurus/WebServerGo/internal/auth"
)

func (cfg *apiConfig) RefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	// Get the refresh token from the request
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 403, "Please authenticate")
		return
	}
	// Validate the refresh token
	err = auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		respondWithError(w, 401, fmt.Sprintf("%s", err))
		return
	}
	tokenRow, err := cfg.db.GetToken(req.Context(), refreshToken)

	if err != nil {
		respondWithError(w, 401, fmt.Sprintf("Refresh token not found"))
		return
	}

	if tokenRow.RevokedAt.Valid {
		respondWithError(w, 401, fmt.Sprintf("Refresh token revoked"))
		return
	}

	if tokenRow.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, fmt.Sprintf("Refresh token expired"))
		return
	}

	accessToken, err := auth.MakeJWT(tokenRow.UserID, cfg.secret, 60*time.Minute)

	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error creating the jwt: %s", err))
	}

	type response struct {
		Token string `json:"token"`
	}
	responseData := response{
		Token: accessToken,
	}
	respondWithJSON(w, http.StatusOK, responseData)
}

func (cfg *apiConfig) RevokeRefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	headers := req.Header
	refreshToken, err := auth.GetBearerToken(headers)
	if err != nil {
		respondWithError(w, 403, "No token given")
		return
	}
	_, err = cfg.db.RevokeToken(req.Context(), refreshToken)

	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error revoking the token: %s", err))
		return
	}
	respondWithJSON(w, 204, "")
}
