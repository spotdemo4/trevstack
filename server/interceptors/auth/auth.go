package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	domainauth "trev.zip/llc/stack/server/auth"
	"trev.zip/llc/stack/server/connect/auth/v1/authv1connect"
)

type AuthInterceptor struct {
	auth *domainauth.Manager
}

func NewAuthInterceptor(manager *domainauth.Manager) *AuthInterceptor {
	return &AuthInterceptor{auth: manager}
}

func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		setCookie := func(cookie *http.Cookie) error {
			return domainauth.SetResponseCookie(ctx, cookie)
		}
		if err := i.authenticate(req.Spec().Procedure, req.Header(), setCookie); err != nil {
			return nil, err
		}
		return next(ctx, req)
	})
}

func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		setCookie := func(cookie *http.Cookie) error {
			conn.ResponseHeader().Add("Set-Cookie", cookie.String())
			return nil
		}
		if err := i.authenticate(conn.Spec().Procedure, conn.RequestHeader(), setCookie); err != nil {
			return err
		}
		return next(ctx, conn)
	})
}

func (i *AuthInterceptor) authenticate(
	procedure string,
	header http.Header,
	setCookie func(*http.Cookie) error,
) error {
	if isPublicAuthProcedure(procedure) {
		return nil
	}

	cookie, err := sessionCookie(header)
	if errors.Is(err, http.ErrNoCookie) {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	if err != nil {
		return connect.NewError(connect.CodePermissionDenied, errors.New("invalid session"))
	}

	if _, err := i.auth.Parse(cookie.Value); err != nil {
		if domainauth.IsExpiredOnly(err) {
			if cookieErr := setCookie(i.auth.DeleteCookie()); cookieErr != nil {
				return connect.NewError(connect.CodeInternal, errors.New("could not clear expired session"))
			}
			return connect.NewError(connect.CodeUnauthenticated, errors.New("session expired"))
		}
		return connect.NewError(connect.CodePermissionDenied, errors.New("invalid session"))
	}

	return nil
}

func sessionCookie(header http.Header) (*http.Cookie, error) {
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
