package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	expectedID := "83064af3-bb81-4514-a6d4-afba340825cd"
	uuid, err := uuid.Parse(expectedID)
	if err != nil {
		t.Errorf("Invalid uuid")
	}

	token, err := MakeJWT(uuid, "test", time.Hour)
	if err != nil {
		t.Errorf("Error creating the jwt: %s", err)
	}
	resultID, err := ValidateJWT(token, "test")
	if err != nil {
		t.Errorf("Invalid jwt %s", err)
	}
	if resultID != uuid {
		t.Errorf(`Error %s is different than %s`, resultID, uuid)
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	tokenString := "ThisIsForTesting"

	headers.Add("Authorization", "Bearer "+tokenString)
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Error(err)
	}
	if token != tokenString {
		t.Errorf("Wanted token to be %s instead it is %s", tokenString, token)
	}
}
