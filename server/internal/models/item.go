package models

import (
	"time"

	itemv1 "github.com/spotdemo4/trevstack/server/internal/services/item/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Item struct {
	ID uint32 `gorm:"primaryKey"`

	Name        string
	Description string
	Price       float32
	Quantity    int
	Added       time.Time

	// User
	UserID uint
	User   User
}

func (i Item) ToConnectV1() *itemv1.Item {
	return &itemv1.Item{
		Id:          &i.ID,
		Name:        i.Name,
		Description: i.Description,
		Price:       i.Price,
		Quantity:    uint32(i.Quantity),
		Added:       timestamppb.New(i.Added),
	}
}
