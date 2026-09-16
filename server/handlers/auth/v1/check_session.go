package v1

import (
	"context"

	authv1 "trev.zip/llc/stack/server/connect/auth/v1"
)

func (h *Handler) CheckSession(
	context.Context,
	*authv1.CheckSessionRequest,
) (*authv1.CheckSessionResponse, error) {
	return &authv1.CheckSessionResponse{}, nil
}
