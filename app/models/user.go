// Package models contains application domain models.
package models

import "time"

// User represents a user managed by the application.
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:150;not null"`
	Phone     string `gorm:"size:30;not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
