package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/stephenafamo/bob"
	_ "modernc.org/sqlite" // Sqlite
)

func NewSQLiteConnection(name string) (*bob.DB, error) {
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
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create new bob db
	bobdb := bob.NewDB(db)

	return &bobdb, nil
}
