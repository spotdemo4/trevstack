package database

import (
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewSQLiteConnection(name string) (*gorm.DB, error) {
	// Find config diretory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	// Create database directory if not exists
	settingsPath := filepath.Join(configDir, "trevstack")
	err = os.MkdirAll(settingsPath, 0766)
	if err != nil {
		return nil, err
	}

	// Open database
	dbPath := filepath.Join(settingsPath, name)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: NewLogger(),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
