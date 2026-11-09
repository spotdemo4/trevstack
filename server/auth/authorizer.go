package auth

import (
	"context"
	"errors"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v2/gen/auth/v1"
)

// Authorizer authenticates calls with the session token in their
// `authorization: Bearer <token>` metadata.
type Authorizer struct {
	sessions *Manager
}

// NewAuthorizer returns an Authorizer that verifies sessions with manager.
func NewAuthorizer(manager *Manager) *Authorizer {
	return &Authorizer{sessions: manager}
}

// Authorize rejects calls to protected methods with Unauthenticated unless
// they carry the token of an unexpired session.
func (a *Authorizer) Authorize(ctx context.Context, info trevrpc.HandlerContext) error {
	if isPublicAuthMethod(info.FullMethod) {
		return nil
	}

	token, err := trevrpc.BearerToken(info.Metadata)
	if err != nil {
		return err
	}
	if _, err := a.sessions.Verify(ctx, token); err != nil {
		if errors.Is(err, ErrInvalidSession) {
			return trevrpc.Unauthenticated("invalid or expired session")
		}
		return trevrpc.Internal("could not verify session")
	}
	return nil
}

func isPublicAuthMethod(method string) bool {
	switch method {
	case authv1.AuthService_Login_FullMethodName,
		authv1.AuthService_Logout_FullMethodName,
		authv1.AuthService_Signup_FullMethodName:
		return true
	default:
		return false
	}
}
