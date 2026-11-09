// Package auth issues, verifies and revokes session tokens.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"errors"
	"time"
)

const (
	tokenBytes = 32
	tokenTTL   = 24 * time.Hour
)

// ErrInvalidSession is returned for a token that names no unexpired session.
var ErrInvalidSession = errors.New("invalid or expired session")

var (
	//go:embed insert_session.sql
	insertSessionSQL string
	//go:embed delete_expired_sessions.sql
	deleteExpiredSessionsSQL string
	//go:embed select_session.sql
	selectSessionSQL string
	//go:embed delete_session.sql
	deleteSessionSQL string
)

// Manager keeps sessions in the database. A session token is random and
// opaque, and only its SHA-256 hash is stored, so the database alone cannot
// be used to sign in.
type Manager struct {
	db  *sql.DB
	now func() time.Time
}

// NewManager returns a Manager that keeps sessions in db.
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db, now: time.Now}
}

// Issue starts a session for the user with userID and returns its token and
// expiry. It first deletes expired sessions, so they do not pile up.
func (m *Manager) Issue(ctx context.Context, userID int64) (string, time.Time, error) {
	secret := make([]byte, tokenBytes)
	_, _ = rand.Read(secret) // never returns an error
	token := base64.RawURLEncoding.EncodeToString(secret)

	now := m.now()
	expires := now.Add(tokenTTL).Truncate(time.Second)
	if _, err := m.db.ExecContext(ctx, deleteExpiredSessionsSQL, now.Unix()); err != nil {
		return "", time.Time{}, err
	}
	if _, err := m.db.ExecContext(ctx, insertSessionSQL, hashToken(token), userID, expires.Unix()); err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

// Verify returns the username of the session token names, or
// ErrInvalidSession if it names none or the session has expired.
func (m *Manager) Verify(ctx context.Context, token string) (string, error) {
	var username string
	err := m.db.QueryRowContext(ctx, selectSessionSQL, hashToken(token), m.now().Unix()).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidSession
	}
	if err != nil {
		return "", err
	}
	return username, nil
}

// Revoke ends the session token names. A token that names no session is
// ignored.
func (m *Manager) Revoke(ctx context.Context, token string) error {
	_, err := m.db.ExecContext(ctx, deleteSessionSQL, hashToken(token))
	return err
}

func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}
