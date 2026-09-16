package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CookieName = "stack_session"
	issuer     = "stack"
	tokenTTL   = 24 * time.Hour
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (c Claims) Validate() error {
	var errs []error
	if c.Subject == "" {
		errs = append(errs, fmt.Errorf("%w: sub", jwt.ErrTokenRequiredClaimMissing))
	}
	if c.IssuedAt == nil {
		errs = append(errs, fmt.Errorf("%w: iat", jwt.ErrTokenRequiredClaimMissing))
	}
	if c.Username == "" {
		errs = append(errs, fmt.Errorf("%w: username", jwt.ErrTokenRequiredClaimMissing))
	}
	return errors.Join(errs...)
}

type Manager struct {
	now    func() time.Time
	secret []byte
	secure bool
}

func NewManager(secret string, secure bool) *Manager {
	return &Manager{
		now:    time.Now,
		secret: []byte(secret),
		secure: secure,
	}
}

func (m *Manager) Issue(userID int64, username string) (string, time.Time, error) {
	now := m.now()
	expires := now.Add(tokenTTL)
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expires, nil
}

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

func SetResponseCookie(ctx context.Context, cookie *http.Cookie) error {
	info, ok := connect.CallInfoForHandlerContext(ctx)
	if !ok {
		return errors.New("connect call info unavailable")
	}
	info.ResponseHeader().Add("Set-Cookie", cookie.String())
	return nil
}
