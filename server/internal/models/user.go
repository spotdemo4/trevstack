package models

import (
	"fmt"

	userv1 "github.com/spotdemo4/trevstack/server/internal/services/user/v1"
)

type User struct {
	ID uint32 `gorm:"primaryKey"`

	Username  string
	Password  string
	Challenge *string

	// Passkeys
	Passkeys []Passkey

	// Profile picture
	ProfilePictureID *uint
	ProfilePicture   *File
}

func (u User) ToConnectV1() *userv1.User {
	var ppid *string
	if u.ProfilePicture != nil {
		id := fmt.Sprintf("/file/%d", u.ProfilePicture.ID)
		ppid = &id
	}

	return &userv1.User{
		Id:             u.ID,
		Username:       u.Username,
		ProfilePicture: ppid,
	}
}
