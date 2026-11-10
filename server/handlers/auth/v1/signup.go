package v1

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v3/gen/auth/v1"
)

//go:embed insert_user.sql
var insertUserSQL string

// Signup creates a user.
func (h *Handler) Signup(
	ctx context.Context,
	req *authv1.SignupRequest,
) (*authv1.SignupResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, trevrpc.Internal("could not hash password")
	}

	_, err = h.db.ExecContext(ctx, insertUserSQL, normalizeUsername(req.GetUsername()), string(hash))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return nil, trevrpc.AlreadyExists("username already exists")
		}
		return nil, trevrpc.Internal("could not create user")
	}

	return &authv1.SignupResponse{}, nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
