// Package models contains application domain models.
package models

import "time"

// User represents a user managed by the application.
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"size:150;not null"`
	Email        string `gorm:"size:255;not null;uniqueIndex"`
	PasswordHash string `gorm:"size:255;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
