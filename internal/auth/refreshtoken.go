package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)
	keyString := hex.EncodeToString(key)
	return keyString, nil
}

func ValidateRefreshToken(tokenString string) error {
	if len(tokenString) != 64 {
		return fmt.Errorf("Invalid token length")
	}

	return nil
}
