package v1

import (
	"context"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v3/gen/auth/v1"
)

// Logout ends the session whose token the call carries. A call without a
// well-formed token, or with one that names no session, has nothing to end
// and succeeds.
func (h *Handler) Logout(
	ctx context.Context,
	_ *authv1.LogoutRequest,
) (*authv1.LogoutResponse, error) {
	if token, err := trevrpc.BearerToken(trevrpc.RequestMetadataFromContext(ctx)); err == nil {
		if err := h.auth.Revoke(ctx, token); err != nil {
			return nil, trevrpc.Internal("could not end session")
		}
	}
	return &authv1.LogoutResponse{}, nil
}
