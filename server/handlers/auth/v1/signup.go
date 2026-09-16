package v1

import (
	"context"
	_ "embed"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	authv1 "trev.zip/llc/stack/server/connect/auth/v1"
	"trev.zip/llc/stack/server/database"
)

//go:embed insert_user.sql
var insertUserSQL string

func (h *Handler) Signup(
	ctx context.Context,
	req *authv1.SignupRequest,
) (*authv1.SignupResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not hash password"))
	}

	db := database.FromContext(ctx)
	_, err = db.ExecContext(ctx, insertUserSQL, normalizeUsername(req.GetUsername()), string(hash))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("username already exists"))
		}
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not create user"))
	}

	return &authv1.SignupResponse{}, nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
