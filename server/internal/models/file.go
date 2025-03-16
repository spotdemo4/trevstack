package models

type File struct {
	ID uint32 `gorm:"primaryKey"`

	Name string
	Data []byte

	// User
	UserID uint
	User   User
}
