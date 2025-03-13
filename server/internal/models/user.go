package models

type User struct {
	ID uint32 `gorm:"primaryKey"`

	Username string
	Password string
}
