package interceptors

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

func WithAuthRedirect(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("start", "method", r.Method, "path", r.URL.Path)
		pathItems := strings.Split(r.URL.Path, "/")

		if len(pathItems) < 2 {
			next.ServeHTTP(w, r)
			return
		}

		switch pathItems[1] {

		case "auth":
			fallthrough
		case "_app":
			fallthrough
		case "favicon.png":
			next.ServeHTTP(w, r)
			return

		default:
			// Check if the request contains a valid cookie token
			cookies := getCookies(r.Header.Get("Cookie"))
			for _, cookie := range cookies {
				if cookie.Name == "token" {
					_, err := validateToken(cookie.Value, key)
					if err == nil {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			// Otherwise redirect
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}
	})
}

type authInterceptor struct {
	key string
}

func NewAuthInterceptor(key string) *authInterceptor {
	return &authInterceptor{
		key: key,
	}
}

func (i *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	// Same as previous UnaryInterceptorFunc.
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		// Check if the request is from a client
		if req.Spec().IsClient {
			return next(ctx, req)
		}

		// Check if the request contains a valid cookie token
		cookies := getCookies(req.Header().Get("Cookie"))
		for _, cookie := range cookies {
			if cookie.Name == "token" {
				subject, err := validateToken(cookie.Value, i.key)
				if err == nil {
					ctx, err = i.newContext(ctx, subject)
					if err == nil {
						return next(ctx, req)
					}
				}
			}
		}

		// Check if the request contains a valid authorization bearer token
		authorization := req.Header().Get("Authorization")
		if authorization != "" && len(authorization) > 7 {
			subject, err := validateToken(authorization[7:], i.key)
			if err == nil {
				ctx, err = i.newContext(ctx, subject)
				if err == nil {
					return next(ctx, req)
				}
			}
		}

		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("could not authenticate"),
		)
	})
}

func (*authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

func (i *authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		// Check if the request contains a valid cookie token
		cookies := getCookies(conn.RequestHeader().Get("Cookie"))
		for _, cookie := range cookies {
			if cookie.Name == "token" {
				subject, err := validateToken(cookie.Value, i.key)
				if err == nil {
					ctx, err = i.newContext(ctx, subject)
					if err == nil {
						return next(ctx, conn)
					}
				}
			}
		}

		// Check if the request contains a valid authorization bearer token
		authorization := conn.RequestHeader().Get("Authorization")
		if authorization != "" && len(authorization) > 7 {
			subject, err := validateToken(authorization[7:], i.key)
			if err == nil {
				ctx, err = i.newContext(ctx, subject)
				if err == nil {
					return next(ctx, conn)
				}
			}
		}

		return connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("could not authenticate"),
		)
	})
}

func getCookies(rawCookies string) []*http.Cookie {
	header := http.Header{}
	header.Add("Cookie", rawCookies)
	request := http.Request{Header: header}

	return request.Cookies()
}

func validateToken(tokenString string, key string) (subject string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(key), nil
	})
	if err != nil {
		return "", err
	}

	switch {
	case token.Valid:
		subject, err := token.Claims.GetSubject()
		if err != nil {
			return "", err
		}

		return subject, nil

	case errors.Is(err, jwt.ErrTokenMalformed):
		log.Println("Token is malformed")
		return "", err

	case errors.Is(err, jwt.ErrSignatureInvalid):
		log.Println("Token signature is invalid")
		return "", err

	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		log.Println("Token is expired or not valid yet")
		return "", err

	default:
		log.Println("Token is invalid")
		return "", err
	}
}

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// userKey is the key for user.User values in Contexts. It is
// unexported; clients use user.NewContext and user.FromContext
// instead of using this key directly.
var userKey key

// NewContext returns a new Context that carries value u.
func (i *authInterceptor) newContext(ctx context.Context, subject string) (context.Context, error) {
	id, err := strconv.Atoi(subject)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, userKey, id), nil
}

// FromContext returns the User value stored in ctx, if any.
func UserFromContext(ctx context.Context) (int, bool) {
	u, ok := ctx.Value(userKey).(int)
	return u, ok
}
