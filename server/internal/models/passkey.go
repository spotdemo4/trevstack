package models

import "time"

type Passkey struct {
	ID string `gorm:"primaryKey"`

	PublicKey string
	Algorithm int
	CreatedAt time.Time
	LastUsed  time.Time

	// User
	UserID uint
	User   User
}
