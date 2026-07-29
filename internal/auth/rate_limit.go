package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LoginAttempt struct {
	Key             string     `gorm:"primaryKey;size:64"`
	Attempts        int        `gorm:"not null"`
	WindowStartedAt time.Time  `gorm:"not null"`
	BlockedUntil    *time.Time `gorm:"index"`
	UpdatedAt       time.Time
}

type LoginLimiter struct {
	database *gorm.DB
	limit    int
	window   time.Duration
	blockFor time.Duration
}

func NewLoginLimiter(database *gorm.DB, limit int, window, blockFor time.Duration) *LoginLimiter {
	return &LoginLimiter{database: database, limit: limit, window: window, blockFor: blockFor}
}

func (limiter *LoginLimiter) Key(remoteAddress, email string) string {
	sum := sha256.Sum256([]byte(remoteAddress + "\x00" + strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(sum[:])
}

func (limiter *LoginLimiter) Check(key string, now time.Time) (bool, time.Duration, error) {
	var attempt LoginAttempt
	err := limiter.database.First(&attempt, "`key` = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	if attempt.BlockedUntil != nil && attempt.BlockedUntil.After(now) {
		return false, attempt.BlockedUntil.Sub(now), nil
	}
	return true, 0, nil
}

func (limiter *LoginLimiter) RecordFailure(key string, now time.Time) error {
	return limiter.database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoNothing: true,
		}).Create(&LoginAttempt{
			Key: key, Attempts: 0, WindowStartedAt: now,
		}).Error; err != nil {
			return err
		}

		var attempt LoginAttempt
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&attempt, "`key` = ?", key).Error
		if err != nil {
			return err
		}
		if now.Sub(attempt.WindowStartedAt) >= limiter.window {
			attempt.Attempts = 0
			attempt.WindowStartedAt = now
			attempt.BlockedUntil = nil
		}
		attempt.Attempts++
		if attempt.Attempts >= limiter.limit {
			blockedUntil := now.Add(limiter.blockFor)
			attempt.BlockedUntil = &blockedUntil
		}
		return tx.Save(&attempt).Error
	})
}

func (limiter *LoginLimiter) Clear(key string) error {
	return limiter.database.Delete(&LoginAttempt{}, "`key` = ?", key).Error
}
