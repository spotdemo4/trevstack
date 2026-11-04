package auth

import (
	"context"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v2/gen/auth/v1"
)

// Authorizer authenticates calls with the session token in their
// `authorization: Bearer <token>` metadata.
type Authorizer struct {
	auth *Manager
}

// NewAuthorizer returns an Authorizer that verifies sessions with manager.
func NewAuthorizer(manager *Manager) *Authorizer {
	return &Authorizer{auth: manager}
}

// Authorize rejects calls to protected methods without a valid session token:
// a missing, malformed or expired token with Unauthenticated, and a token that
// fails verification with PermissionDenied.
func (a *Authorizer) Authorize(_ context.Context, info trevrpc.HandlerContext) error {
	if isPublicAuthMethod(info.FullMethod) {
		return nil
	}

	token, err := trevrpc.BearerToken(info.Metadata)
	if err != nil {
		return err
	}
	if _, err := a.auth.Parse(token); err != nil {
		if IsExpiredOnly(err) {
			return trevrpc.Unauthenticated("session expired")
		}
		return trevrpc.PermissionDenied("invalid session")
	}
	return nil
}

func isPublicAuthMethod(method string) bool {
	switch method {
	case authv1.AuthService_Login_FullMethodName,
		authv1.AuthService_Signup_FullMethodName:
		return true
	default:
		return false
	}
}
