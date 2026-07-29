package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// NewToken returns a random public token and its database-safe SHA-256 hash.
func NewToken() (raw string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate auth token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(bytes)
	return raw, HashToken(raw), nil
}

// HashToken hashes a public token before database lookup or storage.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
