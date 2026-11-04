package v1_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"trev.zip/template/stack/server/database"
	numberv1handler "trev.zip/template/stack/server/handlers/number/v1"
)

// newTest opens an in-memory SQLite DB, runs migrations, and returns a
// handler that uses it. The returned *sql.DB is the same instance the handler
// uses, so callers can seed rows or assert state directly.
func newTest(t *testing.T) (*numberv1handler.Handler, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=true")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Each :memory: connection is its own database, so pin the pool to one
	// connection so every query in the test sees the same DB.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return numberv1handler.New(db), db
}

// seed inserts a single row with an explicit timestamp.
func seed(t *testing.T, db *sql.DB, name string, number uint32, ts time.Time) {
	t.Helper()
	if _, err := db.ExecContext(t.Context(), `INSERT INTO numbers (name, number, timestamp) VALUES (?, ?, ?)`, name, number, ts); err != nil {
		t.Fatalf("seed: %v", err)
	}
}
