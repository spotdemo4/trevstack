package auth

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const managerTestSecret = "01234567890123456789012345678901"

func TestClaimsValidate(t *testing.T) {
	now := jwt.NewNumericDate(time.Now())
	tests := []struct {
		name   string
		claims Claims
		want   []string
	}{
		{
			name: "valid",
			claims: Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:  "42",
					IssuedAt: now,
				},
			},
		},
		{
			name: "missing required claims",
			claims: Claims{
				RegisteredClaims: jwt.RegisteredClaims{},
			},
			want: []string{"sub", "iat"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.claims.Validate()
			if len(test.want) == 0 {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Validate() error = nil, want required claim errors")
			}
			for _, want := range test.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Validate() error = %v, want %q", err, want)
				}
			}
		})
	}
}

func TestManagerIssueAndParse(t *testing.T) {
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	manager := NewManager(managerTestSecret, true)
	manager.now = func() time.Time { return now }

	value, expires, err := manager.Issue("trev")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if want := now.Add(tokenTTL).Truncate(time.Second); !expires.Equal(want) {
		t.Errorf("expires = %v, want %v", expires, want)
	}

	claims, err := manager.Parse(value)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Subject != "trev" || claims.Issuer != issuer {
		t.Errorf("claims = %#v, want subject trev, issuer %q", claims, issuer)
	}
	if !claims.ExpiresAt.Time.Equal(expires) {
		t.Errorf("ExpiresAt = %v, want %v", claims.ExpiresAt.Time, expires)
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.Split(value, ".")[1])
	if err != nil {
		t.Fatalf("decode JWT payload: %v", err)
	}
	if strings.Contains(string(payload), "username") {
		t.Errorf("JWT payload = %s, must not contain username claim", payload)
	}
}

func TestManagerIssueNumericLookingUsername(t *testing.T) {
	manager := NewManager(managerTestSecret, false)
	value, _, err := manager.Issue("42")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims, err := manager.Parse(value)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if claims.Subject != "42" {
		t.Errorf("claims.Subject = %q, want numeric-looking username 42", claims.Subject)
	}
}

func TestManagerParseRejectsInvalidTokens(t *testing.T) {
	manager := NewManager(managerTestSecret, false)
	now := time.Now().UTC()
	validClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			Issuer:    issuer,
		},
	}

	tests := []struct {
		name  string
		value string
		want  error
	}{
		{
			name:  "malformed",
			value: "not-a-jwt",
			want:  jwt.ErrTokenMalformed,
		},
		{
			name:  "wrong signature",
			value: signClaims(t, validClaims, "wrong-secret"),
			want:  jwt.ErrTokenSignatureInvalid,
		},
		{
			name: "legacy issuer",
			value: signClaims(t, Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "42",
					IssuedAt:  validClaims.IssuedAt,
					ExpiresAt: validClaims.ExpiresAt,
					Issuer:    "stack",
				},
			}, managerTestSecret),
			want: jwt.ErrTokenInvalidIssuer,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := manager.Parse(test.value)
			if err == nil || !errors.Is(err, test.want) {
				t.Errorf("Parse() error = %v, want errors.Is(..., %v)", err, test.want)
			}
		})
	}
}

func TestIsExpiredOnly(t *testing.T) {
	manager := NewManager(managerTestSecret, false)
	now := time.Now().UTC()
	expired := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			Issuer:    issuer,
		},
	}

	validExpired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expired).SignedString([]byte(managerTestSecret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	_, validExpiredErr := manager.Parse(validExpired)
	if !IsExpiredOnly(validExpiredErr) {
		t.Fatalf("IsExpiredOnly(valid expired token error) = false, error = %v", validExpiredErr)
	}

	wrongSignature := signClaims(t, expired, "wrong-secret")
	_, wrongSignatureErr := manager.Parse(wrongSignature)
	if IsExpiredOnly(wrongSignatureErr) {
		t.Fatalf("IsExpiredOnly(expired wrong-signature error) = true, error = %v", wrongSignatureErr)
	}
	if errors.Is(wrongSignatureErr, jwt.ErrTokenExpired) {
		t.Errorf("wrong-signature error = %v, want no expired component", wrongSignatureErr)
	}
	if !errors.Is(wrongSignatureErr, jwt.ErrTokenSignatureInvalid) {
		t.Errorf("wrong-signature error = %v, want signature-invalid component", wrongSignatureErr)
	}

	for name, testErr := range map[string]error{
		"nil":                nil,
		"signature only":     jwt.ErrTokenSignatureInvalid,
		"expired and issuer": errors.Join(jwt.ErrTokenExpired, jwt.ErrTokenInvalidIssuer),
	} {
		t.Run(name, func(t *testing.T) {
			if got := IsExpiredOnly(testErr); got {
				t.Errorf("IsExpiredOnly(%v) = true, want false", testErr)
			}
		})
	}
}

func TestManagerCookies(t *testing.T) {
	expires := time.Date(2026, time.January, 3, 4, 5, 6, 0, time.UTC)
	manager := NewManager(managerTestSecret, true)

	cookie := manager.Cookie("token", expires)
	if cookie.Name != CookieName || cookie.Value != "token" || cookie.Path != "/" {
		t.Errorf("session cookie identity = %#v", cookie)
	}
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != int(tokenTTL.Seconds()) {
		t.Errorf("session cookie security/expiry = %#v", cookie)
	}
	if !cookie.Expires.Equal(expires) {
		t.Errorf("session cookie Expires = %v, want %v", cookie.Expires, expires)
	}

	deleted := manager.DeleteCookie()
	if deleted.Name != CookieName || deleted.Path != "/" || deleted.Value != "" {
		t.Errorf("deletion cookie identity = %#v", deleted)
	}
	if deleted.MaxAge != -1 || !deleted.Expires.Equal(time.Unix(1, 0)) || !deleted.Secure || !deleted.HttpOnly || deleted.SameSite != http.SameSiteLaxMode {
		t.Errorf("deletion cookie = %#v", deleted)
	}
}

func signClaims(t *testing.T, claims Claims, secret string) string {
	t.Helper()
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign claims: %v", err)
	}
	return value
}
