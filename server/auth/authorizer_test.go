package auth

import (
	"context"
	"testing"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/v2/gen/auth/v1"
	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

func authorize(manager *Manager, method string, metadata trevrpc.Metadata) error {
	_, err := authorizeContext(manager, method, metadata)
	return err
}

func authorizeContext(manager *Manager, method string, metadata trevrpc.Metadata) (context.Context, error) {
	return NewAuthorizer(manager).AuthorizeContext(context.Background(), trevrpc.HandlerContext{
		FullMethod: method,
		Metadata:   metadata,
	})
}

func bearer(token string) trevrpc.Metadata {
	return trevrpc.Metadata{"authorization": []byte("Bearer " + token)}
}

func codeOf(err error) trevrpc.Code {
	return trevrpc.StatusFromError(err).Code
}

func TestIsPublicAuthMethod(t *testing.T) {
	for _, test := range []struct {
		method string
		want   bool
	}{
		{method: authv1.AuthService_Login_FullMethodName, want: true},
		{method: authv1.AuthService_Logout_FullMethodName, want: true},
		{method: authv1.AuthService_Signup_FullMethodName, want: true},
		{method: cursorv1.CursorService_Send_FullMethodName, want: false},
		{method: "/auth.v1.AuthService/Unknown", want: false},
	} {
		if got := isPublicAuthMethod(test.method); got != test.want {
			t.Errorf("isPublicAuthMethod(%q) = %t, want %t", test.method, got, test.want)
		}
	}
}

func TestAuthorizerAllowsPublicMethodsWithoutToken(t *testing.T) {
	manager, _, _ := newTestManager(t)
	for _, method := range []string{
		authv1.AuthService_Login_FullMethodName,
		authv1.AuthService_Logout_FullMethodName,
		authv1.AuthService_Signup_FullMethodName,
	} {
		ctx, err := authorizeContext(manager, method, nil)
		if err != nil {
			t.Fatalf("AuthorizeContext(%s) error = %v, want nil", method, err)
		}
		if username, ok := UsernameFromContext(ctx); ok {
			t.Errorf("AuthorizeContext(%s) username = %q, want none", method, username)
		}
	}
}

func TestAuthorizerAcceptsValidToken(t *testing.T) {
	manager, db, _ := newTestManager(t)
	token, _ := issue(t, manager, db, "alice")
	for _, metadata := range []trevrpc.Metadata{
		bearer(token),
		{"authorization": []byte("bearer " + token)},
	} {
		ctx, err := authorizeContext(manager, cursorv1.CursorService_Send_FullMethodName, metadata)
		if err != nil {
			t.Fatalf("AuthorizeContext(%q) error = %v, want nil", metadata["authorization"], err)
		}
		// Interceptors and handlers learn the user without looking the
		// session up again.
		if username, ok := UsernameFromContext(ctx); !ok || username != "alice" {
			t.Errorf("AuthorizeContext(%q) username = (%q, %t), want alice", metadata["authorization"], username, ok)
		}
	}
}

func TestAuthorizerRequiresToken(t *testing.T) {
	manager, _, _ := newTestManager(t)
	err := authorize(manager, cursorv1.CursorService_List_FullMethodName, nil)
	if got := codeOf(err); got != trevrpc.CodeUnauthenticated {
		t.Errorf("code = %v, want Unauthenticated (error %v)", got, err)
	}
}

func TestAuthorizerRejectsMalformedMetadata(t *testing.T) {
	manager, db, _ := newTestManager(t)
	token, _ := issue(t, manager, db, "alice")
	for _, value := range []string{"", "Bearer", "Bearer ", "Basic " + token, token, "Bearer " + token + " extra"} {
		err := authorize(manager, cursorv1.CursorService_Send_FullMethodName, trevrpc.Metadata{"authorization": []byte(value)})
		if got := codeOf(err); got != trevrpc.CodeUnauthenticated {
			t.Errorf("Authorize(%q) code = %v, want Unauthenticated (error %v)", value, got, err)
		}
	}
}

func TestAuthorizerRejectsEndedSessions(t *testing.T) {
	manager, db, now := newTestManager(t)
	expired, expires := issue(t, manager, db, "alice")
	*now = expires
	revoked, _ := issue(t, manager, db, "alice")
	if err := manager.Revoke(t.Context(), revoked); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}

	for name, token := range map[string]string{
		"expired": expired,
		"revoked": revoked,
		"unknown": "not-a-session",
		// A JWT from before sessions were opaque.
		"legacy JWT": "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhbGljZSJ9.c2lnbmF0dXJl",
	} {
		t.Run(name, func(t *testing.T) {
			err := authorize(manager, cursorv1.CursorService_Send_FullMethodName, bearer(token))
			if got := codeOf(err); got != trevrpc.CodeUnauthenticated {
				t.Errorf("code = %v, want Unauthenticated (error %v)", got, err)
			}
		})
	}
}

func TestAuthorizerFailsInternallyWithoutDatabase(t *testing.T) {
	manager, db, _ := newTestManager(t)
	token, _ := issue(t, manager, db, "alice")
	_ = db.Close()

	err := authorize(manager, cursorv1.CursorService_Send_FullMethodName, bearer(token))
	if got := codeOf(err); got != trevrpc.CodeInternal {
		t.Errorf("code = %v, want Internal (error %v)", got, err)
	}
}
