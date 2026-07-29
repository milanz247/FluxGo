package auth_test

import (
	"testing"
	"time"

	"fluxgo/internal/auth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoginLimiterBlocksAndClearsAttempts(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:limiter?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrator().CreateTable(&auth.LoginAttempt{}); err != nil {
		t.Fatal(err)
	}

	limiter := auth.NewLoginLimiter(database, 3, time.Minute, 5*time.Minute)
	key := limiter.Key("127.0.0.1", "user@example.com")
	now := time.Now()
	for range 3 {
		if err := limiter.RecordFailure(key, now); err != nil {
			t.Fatal(err)
		}
	}

	allowed, retryAfter, err := limiter.Check(key, now)
	if err != nil {
		t.Fatal(err)
	}
	if allowed || retryAfter <= 0 {
		t.Fatal("expected the login bucket to be blocked")
	}

	if err := limiter.Clear(key); err != nil {
		t.Fatal(err)
	}
	allowed, _, err = limiter.Check(key, now)
	if err != nil || !allowed {
		t.Fatal("expected a cleared bucket to allow login")
	}
}
