package v1_test

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/validate"
	_ "github.com/mattn/go-sqlite3"
	"trev.zip/template/stack/server/connect/number/v1/numberv1connect"
	"trev.zip/template/stack/server/database"
	numberv1handler "trev.zip/template/stack/server/handlers/number/v1"
)

// newTest spins up an in-memory SQLite DB, runs migrations, and starts an
// httptest server wired to the real handler + validate interceptor. The
// returned *sql.DB is the same instance the handler sees via context, so
// callers can seed rows or assert state directly.
func newTest(t *testing.T) (numberv1connect.NumberServiceClient, *sql.DB) {
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

	server := connect.NewServer(validate.NewServerInterceptor())
	numberv1handler.Register(server)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)

	srv := httptest.NewUnstartedServer(mux)
	srv.Config.BaseContext = func(net.Listener) context.Context {
		return database.WithDatabase(context.Background(), db)
	}
	srv.Start()
	t.Cleanup(srv.Close)

	return numberv1connect.NewNumberServiceClient(connect.NewClient(connecthttp.NewTransport(srv.Client(), srv.URL))), db
}

// seed inserts a single row with an explicit timestamp.
func seed(t *testing.T, db *sql.DB, name string, number uint32, ts time.Time) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO numbers (name, number, timestamp) VALUES (?, ?, ?)`, name, number, ts); err != nil {
		t.Fatalf("seed: %v", err)
	}
}
