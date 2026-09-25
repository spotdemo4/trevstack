package v1

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"trev.zip/template/stack/server/auth"
	authv1 "trev.zip/template/stack/server/connect/auth/v1"
)

func (h *Handler) Logout(
	ctx context.Context,
	_ *authv1.LogoutRequest,
) (*authv1.LogoutResponse, error) {
	if err := auth.SetResponseCookie(ctx, h.auth.DeleteCookie()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.New("could not clear session cookie"))
	}
	return &authv1.LogoutResponse{}, nil
}
