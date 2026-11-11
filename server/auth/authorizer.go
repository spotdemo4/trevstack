package auth

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v3/gen/auth/v1"
)

// Authorizer authenticates calls with the session token in their
// `authorization: Bearer <token>` metadata, and passes the session's user on
// to interceptors and handlers through the call's context.
type Authorizer struct {
	sessions *Manager
}

var _ trevrpc.Authorizer = (*Authorizer)(nil)

// NewAuthorizer returns an Authorizer that verifies sessions with manager.
func NewAuthorizer(manager *Manager) *Authorizer {
	return &Authorizer{sessions: manager}
}

// Authorize rejects calls to methods that are not public with Unauthenticated
// unless they carry the token of an unexpired session, and returns the context
// of an authenticated call with the username of its session, which
// UsernameFromContext reads. Calls to public methods keep their context.
func (a *Authorizer) Authorize(ctx context.Context, info trevrpc.HandlerContext) (context.Context, error) {
	if isPublic(info.MethodDescriptor) {
		return ctx, nil
	}

	token, err := trevrpc.BearerToken(info.Metadata)
	if err != nil {
		return ctx, err
	}
	username, err := a.sessions.Verify(ctx, token)
	if err != nil {
		if errors.Is(err, ErrInvalidSession) {
			return ctx, trevrpc.Unauthenticated("invalid or expired session")
		}
		return ctx, trevrpc.Internal("could not verify session")
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

// isPublic reports whether method is marked with the (auth.v1.public) option,
// which lets calls without a session through. Methods the server does not
// have, which have no descriptor, are not public, so callers that are not
// signed in cannot tell which methods exist.
func isPublic(method protoreflect.MethodDescriptor) bool {
	if method == nil {
		return false
	}
	public, _ := proto.GetExtension(method.Options(), authv1.E_Public).(bool)
	return public
}
