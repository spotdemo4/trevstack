package interceptors

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"trev.zip/llc/stack/server/auth"
	authv1 "trev.zip/llc/stack/server/connect/auth/v1"
	"trev.zip/llc/stack/server/connect/auth/v1/authv1connect"
	numberv1 "trev.zip/llc/stack/server/connect/number/v1"
	"trev.zip/llc/stack/server/connect/number/v1/numberv1connect"
	"trev.zip/llc/stack/server/database"
	authhandler "trev.zip/llc/stack/server/handlers/auth/v1"
	numberhandler "trev.zip/llc/stack/server/handlers/number/v1"
)

const interceptorTestSecret = "01234567890123456789012345678901"

type fakeStreamingConn struct {
	spec           connect.Spec
	header         http.Header
	responseHeader http.Header
}

func (c *fakeStreamingConn) Spec() connect.Spec           { return c.spec }
func (c *fakeStreamingConn) Peer() connect.Peer           { return connect.Peer{} }
func (c *fakeStreamingConn) Receive(any) error            { return nil }
func (c *fakeStreamingConn) RequestHeader() http.Header   { return c.header }
func (c *fakeStreamingConn) Send(any) error               { return nil }
func (c *fakeStreamingConn) ResponseHeader() http.Header  { return c.responseHeader }
func (c *fakeStreamingConn) ResponseTrailer() http.Header { return make(http.Header) }

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
		"/auth.v1.AuthService/CheckSession",
		"/auth.v1.AuthService/Other",
		"",
	} {
		if isPublicAuthProcedure(procedure) {
			t.Errorf("isPublicAuthProcedure(%q) = true, want false", procedure)
		}
	}
}

func TestAuthInterceptorCheckSession(t *testing.T) {
	manager := auth.NewManager(interceptorTestSecret, false)
	mux := http.NewServeMux()
	mux.Handle(authhandler.New(manager, connect.WithInterceptors(NewAuthInterceptor(manager))))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := authv1connect.NewAuthServiceClient(http.DefaultClient, srv.URL)

	_, err := client.CheckSession(context.Background(), &authv1.CheckSessionRequest{})
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Fatalf("CheckSession() without cookie code = %v, want Unauthenticated", got)
	}

	token, _, err := manager.Issue(42, "trev")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set("Cookie", (&http.Cookie{Name: auth.CookieName, Value: token}).String())
	if _, err := client.CheckSession(ctx, &authv1.CheckSessionRequest{}); err != nil {
		t.Fatalf("CheckSession() with valid cookie error = %v", err)
	}
}

func TestAuthInterceptorStreamingHandler(t *testing.T) {
	manager := auth.NewManager(interceptorTestSecret, false)
	interceptor := NewAuthInterceptor(manager)

	tests := []struct {
		name       string
		procedure  string
		header     http.Header
		wantCode   connect.Code
		wantErr    bool
		wantCalled bool
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

	valid, _, err := manager.Issue(42, "trev")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	tests = append(tests, struct {
		name       string
		procedure  string
		header     http.Header
		wantCode   connect.Code
		wantErr    bool
		wantCalled bool
	}{
		name:       "protected procedure with valid cookie",
		procedure:  "/number.v1.NumberService/Stream",
		header:     http.Header{"Cookie": []string{(&http.Cookie{Name: auth.CookieName, Value: valid}).String()}},
		wantCalled: true,
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			next := func(context.Context, connect.StreamingHandlerConn) error {
				called = true
				return nil
			}
			conn := &fakeStreamingConn{
				spec:           connect.Spec{Procedure: test.procedure},
				header:         test.header,
				responseHeader: make(http.Header),
			}
			err := interceptor.WrapStreamingHandler(next)(context.Background(), conn)
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
		})
	}
}

func TestAuthInterceptorStreamingHandlerClearsExpiredCookie(t *testing.T) {
	manager := auth.NewManager(interceptorTestSecret, false)
	interceptor := NewAuthInterceptor(manager)
	now := time.Now().UTC()
	token := signInterceptorClaims(t, auth.Claims{
		Username: "trev",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			Issuer:    "stack",
		},
	}, interceptorTestSecret)
	conn := &fakeStreamingConn{
		spec: connect.Spec{Procedure: "/number.v1.NumberService/Stream"},
		header: http.Header{
			"Cookie": []string{(&http.Cookie{Name: auth.CookieName, Value: token}).String()},
		},
		responseHeader: make(http.Header),
	}

	err := interceptor.WrapStreamingHandler(func(context.Context, connect.StreamingHandlerConn) error {
		t.Fatal("next handler called for expired token")
		return nil
	})(context.Background(), conn)
	if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
		t.Fatalf("code = %v, want Unauthenticated (error %v)", got, err)
	}

	response := &http.Response{Header: conn.responseHeader}
	cookies := response.Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.CookieName || cookies[0].MaxAge != -1 {
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
	info.RequestHeader().Set("Cookie", auth.CookieName)

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
	claims := auth.Claims{
		Username: "trev",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			Issuer:    "stack",
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
			info.RequestHeader().Set("Cookie", (&http.Cookie{Name: auth.CookieName, Value: test.token}).String())
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
	manager := auth.NewManager(interceptorTestSecret, false)
	mux := http.NewServeMux()
	mux.Handle(numberhandler.New(connect.WithInterceptors(NewAuthInterceptor(manager))))
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.BaseContext = func(_ net.Listener) context.Context {
		return database.WithDatabase(context.Background(), db)
	}
	srv.Start()
	t.Cleanup(srv.Close)
	transport := &recordingTransport{base: http.DefaultTransport}
	return numberv1connect.NewNumberServiceClient(&http.Client{Transport: transport}, srv.URL), transport
}

func signInterceptorClaims(t *testing.T, claims auth.Claims, secret string) string {
	t.Helper()
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign claims: %v", err)
	}
	return value
}
