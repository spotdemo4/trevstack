package v1

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"
	"trev.zip/llc/stack/server/auth"
	authv1 "trev.zip/llc/stack/server/connect/auth/v1"
	"trev.zip/llc/stack/server/database"
)

//go:embed select_user.sql
var selectUserSQL string

func (h *Handler) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	db := database.FromContext(ctx)

	var id int64
	var username string
	var passwordHash string
	err := db.QueryRowContext(ctx, selectUserSQL, normalizeUsername(req.GetUsername())).Scan(
		&id,
		&username,
		&passwordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(req.GetPassword()))
		return nil, invalidCredentialsError()
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not load user"))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.GetPassword())); err != nil {
		return nil, invalidCredentialsError()
	}

	token, expires, err := h.auth.Issue(id, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not create session"))
	}
	if err := auth.SetResponseCookie(ctx, h.auth.Cookie(token, expires)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not set session cookie"))
	}

	response := &authv1.LoginResponse{}
	response.SetJwt(token)
	return response, nil
}

func invalidCredentialsError() error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
}
