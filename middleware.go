package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Hopesaurus/WebServerGo/internal/auth"
	"github.com/google/uuid"
)

func validateUUIDMiddleware(params []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		//For every string placeholder in the request
		for _, param := range params {
			//Check the value of the placeholder in the url
			paramValue := req.PathValue(param)
			if paramValue == "" {
				//If the value is empty end the context
				respondWithError(w, 400, "Bad request")
				return
			}

			//try to parse the param string as uuid
			uuidValue, err := uuid.Parse(paramValue)
			if err != nil {
				//if it fails fuck the context again
				respondWithError(w, 400, "Not a valid uuid")
				return
			}

			//store the valid uuid in the request context
			ctx = context.WithValue(ctx, param, uuidValue)

		}
		//call the handler with the modified context
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (cfg *apiConfig) validateJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			respondWithError(w, 403, "Please authenticate")
			return
		}
		id, err := auth.ValidateJWT(token, cfg.secret)
		if err != nil {
			respondWithError(w, 401, fmt.Sprintf("%s", err))
			return
		}
		ctx := req.Context()
		ctx = context.WithValue(ctx, "userUUID", id)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
