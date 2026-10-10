package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

// TokenHash calculates the SHA-256 hex digest of a token string.
func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
