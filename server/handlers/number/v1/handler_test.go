package v1_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/proto"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/llc/trevrpc/trevrpc-go/protovalidatetrevrpc"
	"trev.zip/template/stack/server/v2/database"
	numberv1handler "trev.zip/template/stack/server/v2/handlers/number/v1"
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

// validate checks req against the buf.validate rules of its message, as the
// server does before it calls a handler.
func validate(t *testing.T, req proto.Message) error {
	t.Helper()
	desc := req.ProtoReflect().Descriptor()
	check, err := protovalidatetrevrpc.New(nil).ValidateFunc(desc)
	if err != nil {
		t.Fatalf("compile rules of %s: %v", desc.FullName(), err)
	}
	if check == nil {
		t.Fatalf("%s has no rules", desc.FullName())
	}
	return check(req)
}

// assertInvalid checks that the server rejects req with InvalidArgument
// before it reaches the handler.
func assertInvalid(t *testing.T, req proto.Message) {
	t.Helper()
	if got := trevrpc.StatusFromError(validate(t, req)).Code; got != trevrpc.CodeInvalidArgument {
		t.Errorf("validate(%v) code = %v, want InvalidArgument", req, got)
	}
}

// assertValid checks that the server passes req on to the handler.
func assertValid(t *testing.T, req proto.Message) {
	t.Helper()
	if err := validate(t, req); err != nil {
		t.Errorf("validate(%v) = %v, want nil", req, err)
	}
}
