package database

import (
	"embed"
	"log"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/spotdemo4/dbmate-sqlite-modernc/pkg/driver/sqlite" // Modernc sqlite
)

func Migrate(url *url.URL, dbFS *embed.FS) error {
	if dbFS == nil {
		return nil
	}

	db := dbmate.New(url)
	db.Driver()
	db.FS = dbFS

	log.Println("Migrations:")
	migrations, err := db.FindMigrations()
	if err != nil {
		return err
	}
	for _, m := range migrations {
		log.Println(m.Version, m.FilePath)
	}

	log.Println("\nApplying...")
	err = db.CreateAndMigrate()
	if err != nil {
		return err
	}

	return nil
}
