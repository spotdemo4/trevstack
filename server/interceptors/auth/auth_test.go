package auth

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/connect/v2/connectinprocess"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	domainauth "trev.zip/template/stack/server/auth"
	numberv1 "trev.zip/template/stack/server/connect/number/v1"
	"trev.zip/template/stack/server/connect/number/v1/numberv1connect"
	"trev.zip/template/stack/server/database"
	numberhandler "trev.zip/template/stack/server/handlers/number/v1"
)

const interceptorTestSecret = "01234567890123456789012345678901"

type addHandler struct {
	numberv1connect.UnimplementedNumberServiceHandler
	add func(context.Context, *numberv1.AddRequest) (*numberv1.AddResponse, error)
}

func (h addHandler) Add(ctx context.Context, req *numberv1.AddRequest) (*numberv1.AddResponse, error) {
	return h.add(ctx, req)
}

func newAddTest(
	manager *domainauth.Manager,
	add func(context.Context, *numberv1.AddRequest) (*numberv1.AddResponse, error),
) numberv1connect.NumberServiceClient {
	server := connect.NewServer(NewAuthInterceptor(manager).WrapServer)
	numberv1connect.RegisterNumberServiceHandler(server, addHandler{add: add})
	return numberv1connect.NewNumberServiceClient(connect.NewClient(connectinprocess.New(server)))
}

// callStreaming dispatches a server-streaming procedure through interceptor to
// next, with header as the request header.
func callStreaming(
	interceptor *AuthInterceptor,
	procedure string,
	header http.Header,
	next connect.ServerFunc,
) (*connect.CallInfo, error) {
	server := connect.NewServer(interceptor.WrapServer)
	server.Register(connect.Method{
		Spec:    connect.Spec{Procedure: procedure, StreamType: connect.StreamTypeServer},
		Handler: next,
	})
	info := &connect.CallInfo{}
	for key, values := range header {
		info.RequestHeader().SetValues(key, values)
	}
	return info, server.Call(context.Background(), procedure, info, nil)
}

func TestIsPublicAuthProcedure(t *testing.T) {
	for _, procedure := range []string{
		"/auth.v1.AuthService/Signup",
		"/auth.v1.AuthService/Login",
		"/auth.v1.AuthService/Logout",
	} {
		if !isPublicAuthProcedure(procedure) {
			t.Errorf("isPublicAuthProcedure(%q) = false, want true", procedure)
		}
	}
	for _, procedure := range []string{
		"/number.v1.NumberService/Add",
		"/auth.v1.AuthService/Other",
		"",
	} {
		if isPublicAuthProcedure(procedure) {
			t.Errorf("isPublicAuthProcedure(%q) = true, want false", procedure)
		}
	}
}

func TestAuthInterceptorSyntheticUnary(t *testing.T) {
	manager := domainauth.NewManager(interceptorTestSecret, false)
	called := false
	client := newAddTest(manager, func(context.Context, *numberv1.AddRequest) (*numberv1.AddResponse, error) {
		called = true
		return &numberv1.AddResponse{}, nil
	})

	_, err := client.Add(context.Background(), (&numberv1.AddRequest_builder{
		Name: new("security-test"), Number: new(uint32(1)),
	}).Build())
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Fatalf("Add() without cookie code = %v, want Unauthenticated", got)
	}
	if called {
		t.Fatal("handler called without cookie")
	}

	token, _, err := manager.Issue("trev")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("Cookie", (&http.Cookie{Name: domainauth.CookieName, Value: token}).String())
	if _, err := client.Add(ctx, (&numberv1.AddRequest_builder{
		Name: new("security-test"), Number: new(uint32(1)),
	}).Build()); err != nil {
		t.Fatalf("Add() with valid cookie error = %v", err)
	}
	if !called {
		t.Fatal("handler not called with valid cookie")
	}
}

func TestAuthInterceptorUnaryHandlerReceivesClaims(t *testing.T) {
	manager := domainauth.NewManager(interceptorTestSecret, false)
	token, expires, err := manager.Issue("trev")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	var gotClaims *domainauth.Claims
	client := newAddTest(manager, func(ctx context.Context, _ *numberv1.AddRequest) (*numberv1.AddResponse, error) {
		gotClaims, _ = domainauth.ClaimsFromContext(ctx)
		return &numberv1.AddResponse{}, nil
	})

	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("Cookie", (&http.Cookie{Name: domainauth.CookieName, Value: token}).String())
	if _, err := client.Add(ctx, (&numberv1.AddRequest_builder{
		Name: new("security-test"), Number: new(uint32(1)),
	}).Build()); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if gotClaims == nil {
		t.Fatal("handler received no claims")
	}
	if gotClaims.Subject != "trev" {
		t.Errorf("claims subject = %q, want trev", gotClaims.Subject)
	}
	if gotClaims.ExpiresAt == nil || gotClaims.ExpiresAt.Time.Unix() != expires.Unix() {
		t.Errorf("claims expiration = %v, want %v", gotClaims.ExpiresAt, expires)
	}
}

func TestAuthInterceptorRejectedCredentialsDoNotReachUnaryHandler(t *testing.T) {
	manager := domainauth.NewManager(interceptorTestSecret, false)
	called := false
	client := newAddTest(manager, func(context.Context, *numberv1.AddRequest) (*numberv1.AddResponse, error) {
		called = true
		return &numberv1.AddResponse{}, nil
	})

	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("Cookie", domainauth.CookieName)
	_, err := client.Add(ctx, (&numberv1.AddRequest_builder{
		Name: new("security-test"), Number: new(uint32(1)),
	}).Build())
	if err == nil {
		t.Fatal("Add() error = nil, want PermissionDenied")
	}
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Fatalf("Add() code = %v, want PermissionDenied", got)
	}
	if called {
		t.Fatal("handler called for rejected credentials")
	}
}

func TestAuthInterceptorStreamingHandler(t *testing.T) {
	manager := domainauth.NewManager(interceptorTestSecret, false)
	interceptor := NewAuthInterceptor(manager)

	tests := []struct {
		name        string
		procedure   string
		header      http.Header
		wantCode    connect.Code
		wantErr     bool
		wantCalled  bool
		wantClaims  bool
		wantExpires time.Time
	}{
		{
			name:       "public procedure without cookie",
			procedure:  "/auth.v1.AuthService/Logout",
			wantCalled: true,
		},
		{
			name:      "protected procedure without cookie",
			procedure: "/number.v1.NumberService/Stream",
			wantCode:  connect.CodeUnauthenticated,
			wantErr:   true,
		},
	}

	valid, validExpires, err := manager.Issue("trev")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	tests = append(tests, struct {
		name        string
		procedure   string
		header      http.Header
		wantCode    connect.Code
		wantErr     bool
		wantCalled  bool
		wantClaims  bool
		wantExpires time.Time
	}{
		name:        "protected procedure with valid cookie",
		procedure:   "/number.v1.NumberService/Stream",
		header:      http.Header{"Cookie": []string{(&http.Cookie{Name: domainauth.CookieName, Value: valid}).String()}},
		wantCalled:  true,
		wantClaims:  true,
		wantExpires: validExpires,
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			var handlerContext context.Context
			next := func(ctx context.Context, _ connect.Spec, _ connect.ServerStream) error {
				called = true
				handlerContext = ctx
				return nil
			}
			_, err := callStreaming(interceptor, test.procedure, test.header, next)
			if !test.wantErr {
				if err != nil {
					t.Errorf("error = %v, want nil", err)
				}
			} else if got := connect.CodeOf(err); got != test.wantCode {
				t.Errorf("code = %v, want %v (error %v)", got, test.wantCode, err)
			}
			if called != test.wantCalled {
				t.Errorf("next called = %t, want %t", called, test.wantCalled)
			}
			claims, ok := domainauth.ClaimsFromContext(handlerContext)
			if test.wantClaims {
				if !ok {
					t.Fatal("ClaimsFromContext() = missing claims, want validated claims")
				}
				if claims.Subject != "trev" {
					t.Errorf("claims subject = %q, want trev", claims.Subject)
				}
				if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Unix() != test.wantExpires.Unix() {
					t.Errorf("claims expiration = %v, want %v", claims.ExpiresAt, test.wantExpires)
				}
			} else if ok || claims != nil {
				t.Errorf("public handler claims = (%#v, %t), want (nil, false)", claims, ok)
			}
		})
	}
}

func TestAuthInterceptorStreamingHandlerClearsExpiredCookie(t *testing.T) {
	manager := domainauth.NewManager(interceptorTestSecret, false)
	interceptor := NewAuthInterceptor(manager)
	now := time.Now().UTC()
	token := signInterceptorClaims(t, domainauth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			Issuer:    "stack/session/v2",
		},
	}, interceptorTestSecret)
	header := http.Header{
		"Cookie": []string{(&http.Cookie{Name: domainauth.CookieName, Value: token}).String()},
	}

	info, err := callStreaming(interceptor, "/number.v1.NumberService/Stream", header, func(context.Context, connect.Spec, connect.ServerStream) error {
		t.Fatal("next handler called for expired token")
		return nil
	})
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Fatalf("code = %v, want Unauthenticated (error %v)", got, err)
	}

	response := &http.Response{Header: http.Header{"Set-Cookie": info.ResponseHeader().Values("Set-Cookie")}}
	cookies := response.Cookies()
	if len(cookies) != 1 || cookies[0].Name != domainauth.CookieName || cookies[0].MaxAge != -1 {
		t.Errorf("deletion cookies = %v", cookies)
	}
}

func TestAuthInterceptorProtectedUnaryRequiresCookie(t *testing.T) {
	client, _ := newProtectedTest(t)
	_, err := client.Add(context.Background(), (&numberv1.AddRequest_builder{
		Name: new("security-test"), Number: new(uint32(1)),
	}).Build())
	if err == nil {
		t.Fatal("Add() error = nil, want Unauthenticated")
	}
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Errorf("code = %v, want Unauthenticated", got)
	}
}

func TestAuthInterceptorProtectedUnaryRejectsMalformedSessionCookie(t *testing.T) {
	client, _ := newProtectedTest(t)
	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("Cookie", domainauth.CookieName)

	_, err := client.Add(ctx, (&numberv1.AddRequest_builder{
		Name: new("security-test"), Number: new(uint32(1)),
	}).Build())
	if err == nil {
		t.Fatal("Add() error = nil, want PermissionDenied")
	}
	if got := connect.CodeOf(err); got != connect.CodePermissionDenied {
		t.Errorf("code = %v, want PermissionDenied", got)
	}
}

func TestAuthInterceptorProtectedUnaryExpiredTokens(t *testing.T) {
	client, transport := newProtectedTest(t)
	now := time.Now().UTC()
	claims := domainauth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			Issuer:    "stack/session/v2",
		},
	}
	validExpired := signInterceptorClaims(t, claims, interceptorTestSecret)
	wrongSignature := signInterceptorClaims(t, claims, "wrong-secret")

	tests := []struct {
		name       string
		token      string
		wantCode   connect.Code
		wantDelete bool
	}{
		{name: "valid signature but expired", token: validExpired, wantCode: connect.CodeUnauthenticated, wantDelete: true},
		{name: "expired and wrong signature", token: wrongSignature, wantCode: connect.CodePermissionDenied, wantDelete: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport.last = nil
			ctx, info := connect.NewClientContext(context.Background())
			info.RequestHeader().Set("Cookie", (&http.Cookie{Name: domainauth.CookieName, Value: test.token}).String())
			request := (&numberv1.AddRequest_builder{
				Name: new("security-test"), Number: new(uint32(1)),
			}).Build()
			_, err := client.Add(ctx, request)
			if err == nil {
				t.Fatal("Add() error = nil, want authentication error")
			}
			if got := connect.CodeOf(err); got != test.wantCode {
				t.Errorf("code = %v, want %v (error %v)", got, test.wantCode, err)
			}
			if transport.last == nil {
				t.Fatal("Add() made no HTTP response")
			}
			cookies := transport.last.Cookies()
			if deleted := len(cookies) > 0; deleted != test.wantDelete {
				t.Errorf("deletion cookie present = %t, want %t (cookies %v)", deleted, test.wantDelete, cookies)
			}
		})
	}
}

type recordingTransport struct {
	base http.RoundTripper
	last *http.Response
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err == nil {
		t.last = resp
	}
	return resp, err
}

func newProtectedTest(t *testing.T) (numberv1connect.NumberServiceClient, *recordingTransport) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=true")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	manager := domainauth.NewManager(interceptorTestSecret, false)
	server := connect.NewServer(NewAuthInterceptor(manager).WrapServer)
	numberhandler.Register(server)
	mux := http.NewServeMux()
	connecthttp.Mount(mux, server)
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.BaseContext = func(_ net.Listener) context.Context {
		return database.WithDatabase(context.Background(), db)
	}
	srv.Start()
	t.Cleanup(srv.Close)
	transport := &recordingTransport{base: http.DefaultTransport}
	return numberv1connect.NewNumberServiceClient(connect.NewClient(connecthttp.NewTransport(&http.Client{Transport: transport}, srv.URL))), transport
}

func signInterceptorClaims(t *testing.T, claims domainauth.Claims, secret string) string {
	t.Helper()
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign claims: %v", err)
	}
	return value
}
