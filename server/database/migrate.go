package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spotdemo4/trevstack/server/logger"
)

//go:embed migrations
var migrations embed.FS
var sqlre = regexp.MustCompile(`[\n\t\r]+\s*`)

func Migrate(ctx context.Context, db *sql.DB) error {
	log := logger.FromContext(ctx)

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		sqlBytes, err := migrations.ReadFile(filepath.Join("migrations", entry.Name()))
		if err != nil {
			return err
		}

		log.DebugContext(ctx, "applying migration", "file", entry.Name(), "sql", sqlre.ReplaceAllString(string(sqlBytes), " "))
		_, err = db.ExecContext(ctx, string(sqlBytes))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}
