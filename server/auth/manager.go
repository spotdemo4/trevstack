// Package auth issues and verifies JWT session cookies.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// CookieName is the name of the session cookie.
	CookieName = "stack_session"
	issuer     = "stack/session/v2"
	tokenTTL   = 24 * time.Hour
)

// Claims are the JWT claims of a session token.
type Claims struct {
	jwt.RegisteredClaims
}

// Validate requires the subject and issued-at claims.
func (c Claims) Validate() error {
	var errs []error
	if c.Subject == "" {
		errs = append(errs, fmt.Errorf("%w: sub", jwt.ErrTokenRequiredClaimMissing))
	}
	if c.IssuedAt == nil {
		errs = append(errs, fmt.Errorf("%w: iat", jwt.ErrTokenRequiredClaimMissing))
	}
	return errors.Join(errs...)
}

// Manager issues and parses session tokens and builds their cookies.
type Manager struct {
	now    func() time.Time
	secret []byte
	secure bool
}

// NewManager returns a Manager that signs tokens with secret. When secure is
// true, cookies are only sent over HTTPS.
func NewManager(secret string, secure bool) *Manager {
	return &Manager{
		now:    time.Now,
		secret: []byte(secret),
		secure: secure,
	}
}

// Issue signs a token for username and returns it with its expiry.
func (m *Manager) Issue(username string) (string, time.Time, error) {
	now := m.now()
	expires := jwt.NewNumericDate(now.Add(tokenTTL))
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expires,
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires.Time, nil
}

// Parse verifies a signed token and returns its claims.
func (m *Manager) Parse(value string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(
		value,
		claims,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(m.now),
		jwt.WithIssuer(issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// IsExpiredOnly reports whether err from Parse means the token is valid
// except for having expired.
func IsExpiredOnly(err error) bool {
	if !errors.Is(err, jwt.ErrTokenExpired) {
		return false
	}

	for _, other := range []error{
		jwt.ErrInvalidKey,
		jwt.ErrInvalidKeyType,
		jwt.ErrTokenMalformed,
		jwt.ErrTokenUnverifiable,
		jwt.ErrTokenSignatureInvalid,
		jwt.ErrTokenRequiredClaimMissing,
		jwt.ErrTokenInvalidAudience,
		jwt.ErrTokenUsedBeforeIssued,
		jwt.ErrTokenInvalidIssuer,
		jwt.ErrTokenInvalidSubject,
		jwt.ErrTokenNotValidYet,
		jwt.ErrTokenInvalidId,
		jwt.ErrInvalidType,
	} {
		if errors.Is(err, other) {
			return false
		}
	}

	return true
}

// Cookie returns the session cookie for a token.
func (m *Manager) Cookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(tokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// DeleteCookie returns a cookie that clears the session cookie.
func (m *Manager) DeleteCookie() *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	}
}

// SetResponseCookie adds cookie to the response of the Connect call in ctx.
func SetResponseCookie(ctx context.Context, cookie *http.Cookie) error {
	info, ok := connect.CallInfoForServerContext(ctx)
	if !ok {
		return errors.New("connect call info unavailable")
	}
	info.ResponseHeader().Add("Set-Cookie", cookie.String())
	return nil
}
