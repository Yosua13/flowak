package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// RandomToken generates a cryptographically secure random base64url string.
func RandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
