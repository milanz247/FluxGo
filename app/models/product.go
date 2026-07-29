package models

import "time"

// Product is a stock item with a name, price, and quantity on hand.
type Product struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"size:150;not null;index"`
	Price     float64 `gorm:"not null"`
	Qty       int     `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
