package database

import (
	"github.com/spotdemo4/trevstack/server/internal/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&models.User{}, &models.Item{}, &models.File{})
	if err != nil {
		return err
	}

	return nil
}
