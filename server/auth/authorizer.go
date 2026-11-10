package auth

import (
	"context"
	"errors"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v2/gen/auth/v1"
)

// Authorizer authenticates calls with the session token in their
// `authorization: Bearer <token>` metadata, and passes the session's user on
// to interceptors and handlers through the call's context.
type Authorizer struct {
	sessions *Manager
}

var _ trevrpc.ContextAuthorizer = (*Authorizer)(nil)

// NewAuthorizer returns an Authorizer that verifies sessions with manager.
func NewAuthorizer(manager *Manager) *Authorizer {
	return &Authorizer{sessions: manager}
}

// Authorize rejects calls to protected methods with Unauthenticated unless
// they carry the token of an unexpired session.
func (a *Authorizer) Authorize(ctx context.Context, info trevrpc.HandlerContext) error {
	_, err := a.AuthorizeContext(ctx, info)
	return err
}

// AuthorizeContext authorizes calls like Authorize, and returns the context of
// an authenticated call with the username of its session, which
// UsernameFromContext reads. Calls to public methods keep their context.
func (a *Authorizer) AuthorizeContext(ctx context.Context, info trevrpc.HandlerContext) (context.Context, error) {
	if isPublicAuthMethod(info.FullMethod) {
		return ctx, nil
	}

	token, err := trevrpc.BearerToken(info.Metadata)
	if err != nil {
		return nil, err
	}
	username, err := a.sessions.Verify(ctx, token)
	if err != nil {
		if errors.Is(err, ErrInvalidSession) {
			return nil, trevrpc.Unauthenticated("invalid or expired session")
		}
		return nil, trevrpc.Internal("could not verify session")
	}
	return ContextWithUsername(ctx, username), nil
}

type usernameKey struct{}

// ContextWithUsername returns a copy of ctx that carries username as the
// Authorizer's context of an authenticated call does, for tests.
func ContextWithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey{}, username)
}

// UsernameFromContext returns the username of the session that authenticated
// the call ctx belongs to, and false for calls to public methods, which need
// no session.
func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey{}).(string)
	return username, ok
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
