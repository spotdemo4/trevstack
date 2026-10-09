package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect/v2"
	domainauth "trev.zip/template/stack/server/auth"
	"trev.zip/template/stack/server/connect/auth/v1/authv1connect"
)

type AuthInterceptor struct {
	auth *domainauth.Manager
}

func NewAuthInterceptor(manager *domainauth.Manager) *AuthInterceptor {
	return &AuthInterceptor{auth: manager}
}

func (i *AuthInterceptor) WrapServer(next connect.ServerFunc) connect.ServerFunc {
	return func(
		ctx context.Context,
		spec connect.Spec,
		stream connect.ServerStream,
	) error {
		claims, err := i.authenticate(ctx, spec.Procedure)
		if err != nil {
			return err
		}
		if claims != nil {
			ctx = domainauth.WithClaims(ctx, claims)
		}
		return next(ctx, spec, stream)
	}
}

func (i *AuthInterceptor) authenticate(
	ctx context.Context,
	procedure string,
) (*domainauth.Claims, error) {
	if isPublicAuthProcedure(procedure) {
		return nil, nil
	}

	info, ok := connect.CallInfoForServerContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, "connect call info unavailable")
	}

	cookie, err := sessionCookie(info.RequestHeader())
	if errors.Is(err, http.ErrNoCookie) {
		return nil, connect.NewError(connect.CodeUnauthenticated, "authentication required")
	}
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, "invalid session")
	}

	claims, err := i.auth.Parse(cookie.Value)
	if err != nil {
		if domainauth.IsExpiredOnly(err) {
			if cookieErr := domainauth.SetResponseCookie(ctx, i.auth.DeleteCookie()); cookieErr != nil {
				return nil, connect.NewError(connect.CodeInternal, "could not clear expired session")
			}
			return nil, connect.NewError(connect.CodeUnauthenticated, "session expired")
		}
		return nil, connect.NewError(connect.CodePermissionDenied, "invalid session")
	}

	return claims, nil
}

func sessionCookie(header *connect.Header) (*http.Cookie, error) {
	var session *http.Cookie
	for _, line := range header.Values("Cookie") {
		for part := range strings.SplitSeq(line, ";") {
			part = strings.TrimSpace(part)
			name, _, hasValue := strings.Cut(part, "=")
			if strings.TrimSpace(name) != domainauth.CookieName {
				continue
			}
			if !hasValue || name != domainauth.CookieName || session != nil {
				return nil, errors.New("invalid session cookie")
			}

			cookies, err := http.ParseCookie(part)
			if err != nil || len(cookies) != 1 || cookies[0].Name != domainauth.CookieName {
				return nil, errors.New("invalid session cookie")
			}
			session = cookies[0]
		}
	}
	if session == nil {
		return nil, http.ErrNoCookie
	}
	return session, nil
}

func isPublicAuthProcedure(procedure string) bool {
	switch procedure {
	case authv1connect.AuthServiceLoginProcedure,
		authv1connect.AuthServiceLogoutProcedure,
		authv1connect.AuthServiceSignupProcedure:
		return true
	default:
		return false
	}
}
