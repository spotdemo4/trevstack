package v1_test

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"trev.zip/template/stack/server/auth"
	authv1 "trev.zip/template/stack/server/connect/auth/v1"
	"trev.zip/template/stack/server/connect/auth/v1/authv1connect"
	"trev.zip/template/stack/server/database"
	authhandler "trev.zip/template/stack/server/handlers/auth/v1"
	ratelimitinterceptor "trev.zip/template/stack/server/interceptors/ratelimit"
)

const handlerTestSecret = "01234567890123456789012345678901"

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

func newAuthTest(
	t *testing.T,
	secure bool,
	opts ...connect.HandlerOption,
) (authv1connect.AuthServiceClient, *sql.DB, *recordingTransport, *auth.Manager) {
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

	manager := auth.NewManager(handlerTestSecret, secure)
	mux := http.NewServeMux()
	mux.Handle(authhandler.New(manager, opts...))
	srv := httptest.NewUnstartedServer(mux)
	srv.Config.BaseContext = func(_ net.Listener) context.Context {
		return database.WithDatabase(context.Background(), db)
	}
	srv.Start()
	t.Cleanup(srv.Close)

	transport := &recordingTransport{base: http.DefaultTransport}
	client := &http.Client{Transport: transport}
	return authv1connect.NewAuthServiceClient(client, srv.URL), db, transport, manager
}

func TestSignupCreatesNormalizedUser(t *testing.T) {
	client, db, _, _ := newAuthTest(t, false)

	_, err := client.Signup(context.Background(), authv1.SignupRequest_builder{
		Username: new("  Alice "),
		Password: new("correct horse battery staple"),
	}.Build())
	if err != nil {
		t.Fatalf("Signup() error = %v", err)
	}

	var username, passwordHash string
	if err := db.QueryRow("SELECT username, password_hash FROM users").Scan(&username, &passwordHash); err != nil {
		t.Fatalf("query user: %v", err)
	}
	if username != "alice" {
		t.Errorf("username = %q, want alice", username)
	}
	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("correct horse battery staple")) != nil {
		t.Error("stored password hash does not match submitted password")
	}
}

func TestSignupRejectsDuplicateUsername(t *testing.T) {
	client, _, _, _ := newAuthTest(t, false)
	request := authv1.SignupRequest_builder{
		Username: new("Alice"),
		Password: new("password"),
	}.Build()
	if _, err := client.Signup(context.Background(), request); err != nil {
		t.Fatalf("first Signup() error = %v", err)
	}
	_, err := client.Signup(context.Background(), request)
	if err == nil {
		t.Fatal("second Signup() error = nil, want AlreadyExists")
	}
	if got := connect.CodeOf(err); got != connect.CodeAlreadyExists {
		t.Errorf("code = %v, want AlreadyExists", got)
	}
}

func TestLoginIssuesJWTAndSecureCookie(t *testing.T) {
	client, db, transport, manager := newAuthTest(t, true)
	password := "correct horse battery staple"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "alice", hash); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	response, err := client.Login(context.Background(), authv1.LoginRequest_builder{
		Username: new("  ALICE "),
		Password: new(password),
	}.Build())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if response.GetJwt() == "" {
		t.Fatal("Login() returned empty JWT")
	}
	claims, err := manager.Parse(response.GetJwt())
	if err != nil {
		t.Fatalf("parse returned JWT: %v", err)
	}
	if claims.Subject != "alice" || response.GetSub() != "alice" {
		t.Errorf("subject = (%q, %q), want normalized alice", claims.Subject, response.GetSub())
	}
	if response.GetExp() != claims.ExpiresAt.Time.Unix() {
		t.Errorf("response Exp = %d, want token expiration %d", response.GetExp(), claims.ExpiresAt.Time.Unix())
	}

	if transport.last == nil {
		t.Fatal("Login() made no HTTP response")
	}
	cookies := transport.last.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Set-Cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != auth.CookieName || cookie.Value != response.GetJwt() || cookie.Path != "/" {
		t.Errorf("session cookie = %#v", cookie)
	}
	if cookie.Expires.Unix() != response.GetExp() {
		t.Errorf("cookie expiration = %d, want %d", cookie.Expires.Unix(), response.GetExp())
	}
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("session cookie security attributes = %#v", cookie)
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	client, db, transport, _ := newAuthTest(t, false)
	hash, err := bcrypt.GenerateFromPassword([]byte("correct password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", "alice", hash); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	for _, test := range []struct {
		name     string
		username string
		password string
	}{
		{name: "unknown user", username: "missing", password: "correct password"},
		{name: "wrong password", username: "alice", password: "wrong password"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.Login(context.Background(), authv1.LoginRequest_builder{
				Username: new(test.username),
				Password: new(test.password),
			}.Build())
			if err == nil {
				t.Fatal("Login() error = nil, want Unauthenticated")
			}
			if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
				t.Errorf("code = %v, want Unauthenticated", got)
			}
			if transport.last != nil && len(transport.last.Cookies()) != 0 {
				t.Errorf("invalid login set cookies: %v", transport.last.Cookies())
			}
		})
	}
}

func TestAuthRateLimitReturnsResourceExhausted(t *testing.T) {
	limiter := ratelimitinterceptor.NewRateLimitInterceptor(map[string]ratelimitinterceptor.RateLimitPolicy{
		authv1connect.AuthServiceLoginProcedure: {Requests: 1, Window: time.Hour},
	}, nil)
	client, _, transport, _ := newAuthTest(
		t,
		false,
		connect.WithInterceptors(limiter),
	)
	request := authv1.LoginRequest_builder{
		Username: new("missing"),
		Password: new("password"),
	}.Build()

	if _, err := client.Login(context.Background(), request); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("first Login() code = %v, want Unauthenticated", connect.CodeOf(err))
	}
	_, err := client.Login(context.Background(), request)
	if got := connect.CodeOf(err); got != connect.CodeResourceExhausted {
		t.Fatalf("second Login() code = %v, want ResourceExhausted", got)
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("second Login() error type = %T, want *connect.Error", err)
	}
	if got := connectErr.Meta().Get("Retry-After"); got == "" {
		t.Error("second Login() Retry-After is empty")
	}
	if transport.last == nil || transport.last.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second Login() HTTP response = %#v, want status 429", transport.last)
	}
}

func TestAuthRateLimitUsesIndependentProcedureBuckets(t *testing.T) {
	limiter := ratelimitinterceptor.NewRateLimitInterceptor(map[string]ratelimitinterceptor.RateLimitPolicy{
		authv1connect.AuthServiceLoginProcedure:  {Requests: 1, Window: time.Hour},
		authv1connect.AuthServiceSignupProcedure: {Requests: 1, Window: time.Hour},
	}, nil)
	client, _, _, _ := newAuthTest(
		t,
		false,
		connect.WithInterceptors(limiter),
	)
	password := "correct horse battery staple"

	if _, err := client.Signup(context.Background(), authv1.SignupRequest_builder{
		Username: new("alice"),
		Password: new(password),
	}.Build()); err != nil {
		t.Fatalf("Signup() error = %v", err)
	}
	if _, err := client.Login(context.Background(), authv1.LoginRequest_builder{
		Username: new("alice"),
		Password: new(password),
	}.Build()); err != nil {
		t.Fatalf("Login() after Signup() error = %v", err)
	}

	if _, err := client.Signup(context.Background(), authv1.SignupRequest_builder{
		Username: new("bob"),
		Password: new(password),
	}.Build()); connect.CodeOf(err) != connect.CodeResourceExhausted {
		t.Errorf("second Signup() code = %v, want ResourceExhausted", connect.CodeOf(err))
	}
	if _, err := client.Login(context.Background(), authv1.LoginRequest_builder{
		Username: new("alice"),
		Password: new(password),
	}.Build()); connect.CodeOf(err) != connect.CodeResourceExhausted {
		t.Errorf("second Login() code = %v, want ResourceExhausted", connect.CodeOf(err))
	}
	for request := 1; request <= 2; request++ {
		if _, err := client.Logout(context.Background(), authv1.LogoutRequest_builder{}.Build()); err != nil {
			t.Errorf("Logout() request %d error = %v", request, err)
		}
	}
}

func TestLogoutDeletesSessionCookie(t *testing.T) {
	client, _, transport, _ := newAuthTest(t, true)

	_, err := client.Logout(context.Background(), authv1.LogoutRequest_builder{}.Build())
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if transport.last == nil {
		t.Fatal("Logout() made no HTTP response")
	}
	cookies := transport.last.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Set-Cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != auth.CookieName || cookie.Path != "/" || cookie.MaxAge != -1 || cookie.Value != "" {
		t.Errorf("deletion cookie = %#v", cookie)
	}
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("deletion cookie security attributes = %#v", cookie)
	}
	if !cookie.Expires.Equal(time.Unix(1, 0)) {
		t.Errorf("deletion cookie Expires = %v, want Unix(1, 0)", cookie.Expires)
	}
}
