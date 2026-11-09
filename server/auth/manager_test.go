package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"trev.zip/template/stack/server/v2/database"
)

// newTestManager returns a Manager on a migrated in-memory database with the
// users alice and bob, whose clock reads *now.
func newTestManager(t *testing.T) (*Manager, *sql.DB, *time.Time) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=true")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	for _, username := range []string{"alice", "bob"} {
		if _, err := db.ExecContext(t.Context(), "INSERT INTO users (username, password_hash) VALUES (?, 'hash')", username); err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	now := time.Date(2026, time.January, 2, 3, 4, 5, 600, time.UTC)
	manager := NewManager(db)
	manager.now = func() time.Time { return now }
	return manager, db, &now
}

func userID(t *testing.T, db *sql.DB, username string) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRowContext(t.Context(), "SELECT id FROM users WHERE username = ?", username).Scan(&id); err != nil {
		t.Fatalf("query user id: %v", err)
	}
	return id
}

func issue(t *testing.T, manager *Manager, db *sql.DB, username string) (string, time.Time) {
	t.Helper()
	token, expires, err := manager.Issue(t.Context(), userID(t, db, username))
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return token, expires
}

func sessionCount(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM sessions").Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	return count
}

func assertSession(t *testing.T, manager *Manager, token, want string) {
	t.Helper()
	username, err := manager.Verify(t.Context(), token)
	if err != nil || username != want {
		t.Errorf("Verify() = (%q, %v), want (%q, nil)", username, err, want)
	}
}

func assertNoSession(t *testing.T, manager *Manager, token string) {
	t.Helper()
	if _, err := manager.Verify(t.Context(), token); !errors.Is(err, ErrInvalidSession) {
		t.Errorf("Verify() error = %v, want ErrInvalidSession", err)
	}
}

func TestManagerIssueAndVerify(t *testing.T) {
	manager, db, now := newTestManager(t)

	token, expires := issue(t, manager, db, "alice")
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != tokenBytes {
		t.Errorf("token %q is not %d bytes of unpadded base64url (error %v)", token, tokenBytes, err)
	}
	if want := now.Add(tokenTTL).Truncate(time.Second); !expires.Equal(want) {
		t.Errorf("expires = %v, want %v", expires, want)
	}
	assertSession(t, manager, token, "alice")
}

func TestManagerStoresOnlyTokenHashes(t *testing.T) {
	manager, db, _ := newTestManager(t)
	token, _ := issue(t, manager, db, "alice")

	var stored []byte
	if err := db.QueryRowContext(t.Context(), "SELECT token_hash FROM sessions").Scan(&stored); err != nil {
		t.Fatalf("query session: %v", err)
	}
	if want := sha256.Sum256([]byte(token)); string(stored) != string(want[:]) {
		t.Errorf("token_hash = %x, want SHA-256 of the token %x", stored, want)
	}
	assertNoSession(t, manager, base64.RawURLEncoding.EncodeToString(stored))
}

func TestManagerIssuesDistinctTokens(t *testing.T) {
	manager, db, _ := newTestManager(t)
	first, _ := issue(t, manager, db, "alice")
	second, _ := issue(t, manager, db, "alice")
	if first == second {
		t.Fatal("Issue() returned the same token twice")
	}
	assertSession(t, manager, first, "alice")
	assertSession(t, manager, second, "alice")
}

func TestManagerVerifyRejectsUnknownTokens(t *testing.T) {
	manager, db, _ := newTestManager(t)
	issue(t, manager, db, "alice")
	for _, token := range []string{"", "not-a-token", base64.RawURLEncoding.EncodeToString(make([]byte, tokenBytes))} {
		assertNoSession(t, manager, token)
	}
}

func TestManagerSessionsExpire(t *testing.T) {
	manager, db, now := newTestManager(t)
	token, expires := issue(t, manager, db, "alice")

	*now = expires.Add(-time.Second)
	assertSession(t, manager, token, "alice")
	*now = expires
	assertNoSession(t, manager, token)
}

func TestManagerIssueDeletesExpiredSessions(t *testing.T) {
	manager, db, now := newTestManager(t)
	_, expires := issue(t, manager, db, "alice")
	issue(t, manager, db, "bob")

	*now = expires
	current, _ := issue(t, manager, db, "bob")
	if got := sessionCount(t, db); got != 1 {
		t.Errorf("sessions = %d, want only the one just issued", got)
	}
	assertSession(t, manager, current, "bob")
}

func TestManagerRevoke(t *testing.T) {
	manager, db, _ := newTestManager(t)
	revoked, _ := issue(t, manager, db, "alice")
	other, _ := issue(t, manager, db, "alice")

	if err := manager.Revoke(t.Context(), revoked); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	assertNoSession(t, manager, revoked)
	assertSession(t, manager, other, "alice")

	for _, token := range []string{revoked, "not-a-token"} {
		if err := manager.Revoke(t.Context(), token); err != nil {
			t.Errorf("Revoke(%q) of no session error = %v, want nil", token, err)
		}
	}
}

func TestManagerSessionsEndWithTheirUser(t *testing.T) {
	manager, db, _ := newTestManager(t)
	token, _ := issue(t, manager, db, "alice")

	if _, err := db.ExecContext(t.Context(), "DELETE FROM users WHERE username = 'alice'"); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	assertNoSession(t, manager, token)
	if got := sessionCount(t, db); got != 0 {
		t.Errorf("sessions = %d, want 0 after deleting their user", got)
	}
}
