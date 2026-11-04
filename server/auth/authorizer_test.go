package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	authv1 "trev.zip/template/stack/server/gen/auth/v1"
	numberv1 "trev.zip/template/stack/server/gen/number/v1"
)

const authorizerTestSecret = managerTestSecret

func authorize(t *testing.T, method string, metadata trevrpc.Metadata) error {
	t.Helper()
	authorizer := NewAuthorizer(NewManager(authorizerTestSecret))
	return authorizer.Authorize(context.Background(), trevrpc.HandlerContext{
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
		{method: authv1.AuthService_Signup_FullMethodName, want: true},
		{method: numberv1.NumberService_Add_FullMethodName, want: false},
		{method: "/auth.v1.AuthService/Unknown", want: false},
	} {
		if got := isPublicAuthMethod(test.method); got != test.want {
			t.Errorf("isPublicAuthMethod(%q) = %t, want %t", test.method, got, test.want)
		}
	}
}

func TestAuthorizerAllowsPublicMethodsWithoutToken(t *testing.T) {
	for _, method := range []string{authv1.AuthService_Login_FullMethodName, authv1.AuthService_Signup_FullMethodName} {
		if err := authorize(t, method, nil); err != nil {
			t.Errorf("Authorize(%s) error = %v, want nil", method, err)
		}
	}
}

func TestAuthorizerAcceptsValidToken(t *testing.T) {
	token, _, err := NewManager(authorizerTestSecret).Issue("alice")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	for _, metadata := range []trevrpc.Metadata{
		bearer(token),
		{"authorization": []byte("bearer " + token)},
	} {
		if err := authorize(t, numberv1.NumberService_Add_FullMethodName, metadata); err != nil {
			t.Errorf("Authorize(%q) error = %v, want nil", metadata["authorization"], err)
		}
	}
}

func TestAuthorizerRequiresToken(t *testing.T) {
	err := authorize(t, numberv1.NumberService_List_FullMethodName, nil)
	if got := codeOf(err); got != trevrpc.CodeUnauthenticated {
		t.Errorf("code = %v, want Unauthenticated (error %v)", got, err)
	}
}

func TestAuthorizerRejectsMalformedMetadata(t *testing.T) {
	token, _, err := NewManager(authorizerTestSecret).Issue("alice")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	for _, value := range []string{"", "Bearer", "Bearer ", "Basic " + token, token, "Bearer " + token + " extra"} {
		err := authorize(t, numberv1.NumberService_Add_FullMethodName, trevrpc.Metadata{"authorization": []byte(value)})
		if got := codeOf(err); got != trevrpc.CodeUnauthenticated {
			t.Errorf("Authorize(%q) code = %v, want Unauthenticated (error %v)", value, got, err)
		}
	}
}

func TestAuthorizerExpiredTokens(t *testing.T) {
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			Issuer:    "stack/session/v2",
		},
	}

	for _, test := range []struct {
		name     string
		token    string
		wantCode trevrpc.Code
	}{
		{name: "valid signature but expired", token: signClaims(t, claims, authorizerTestSecret), wantCode: trevrpc.CodeUnauthenticated},
		{name: "expired and wrong signature", token: signClaims(t, claims, "wrong-secret"), wantCode: trevrpc.CodePermissionDenied},
		{name: "not a token", token: "not-a-token", wantCode: trevrpc.CodePermissionDenied},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := authorize(t, numberv1.NumberService_Add_FullMethodName, bearer(test.token))
			if got := codeOf(err); got != test.wantCode {
				t.Errorf("code = %v, want %v (error %v)", got, test.wantCode, err)
			}
		})
	}
}
