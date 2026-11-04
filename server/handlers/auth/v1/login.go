package v1

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"

	"golang.org/x/crypto/bcrypt"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v2/gen/auth/v1"
)

//go:embed select_user.sql
var selectUserSQL string

// Login verifies a user's credentials and issues a session token.
func (h *Handler) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	if err := validateCredentials(req); err != nil {
		return nil, err
	}

	var username string
	var passwordHash string
	err := h.db.QueryRowContext(ctx, selectUserSQL, normalizeUsername(req.GetUsername())).Scan(
		&username,
		&passwordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(req.GetPassword()))
		return nil, invalidCredentialsError()
	}
	if err != nil {
		return nil, trevrpc.Internal("could not load user")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.GetPassword())); err != nil {
		return nil, invalidCredentialsError()
	}

	token, expires, err := h.auth.Issue(username)
	if err != nil {
		return nil, trevrpc.Internal("could not create session")
	}

	response := &authv1.LoginResponse{}
	response.SetJwt(token)
	response.SetSub(username)
	response.SetExp(expires.Unix())
	return response, nil
}

func invalidCredentialsError() error {
	return trevrpc.Unauthenticated("invalid credentials")
}
