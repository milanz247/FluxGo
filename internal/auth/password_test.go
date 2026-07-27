package auth_test

import (
	"strings"
	"testing"

	"fluxgo/internal/auth"
)

func TestPasswordHashAndVerification(t *testing.T) {
	hash, err := auth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if strings.Contains(hash, "correct horse battery staple") {
		t.Fatal("password hash contains the plaintext password")
	}
	if !auth.VerifyPassword("correct horse battery staple", hash) {
		t.Fatal("expected the correct password to verify")
	}
	if auth.VerifyPassword("wrong password", hash) {
		t.Fatal("did not expect an incorrect password to verify")
	}
}

func TestPasswordHashesUseUniqueSalts(t *testing.T) {
	first, err := auth.HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := auth.HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("expected unique salts to produce different hashes")
	}
}
