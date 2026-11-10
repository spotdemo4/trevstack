package v1_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/llc/trevrpc/trevrpc-go/protovalidatetrevrpc"
	"trev.zip/template/stack/server/v3/auth"
	"trev.zip/template/stack/server/v3/database"
	authv1 "trev.zip/template/stack/server/v3/gen/auth/v1"
	authhandler "trev.zip/template/stack/server/v3/handlers/auth/v1"
)

func newAuthTest(t *testing.T) (*authhandler.Handler, *sql.DB, *auth.Manager) {
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

	manager := auth.NewManager(db)
	return authhandler.New(db, manager), db, manager
}

func codeOf(err error) trevrpc.Code {
	return trevrpc.StatusFromError(err).Code
}

// validate checks req against the buf.validate rules of its message, as the
// server does before it calls a handler.
func validate(t *testing.T, req proto.Message) error {
	t.Helper()
	desc := req.ProtoReflect().Descriptor()
	check, err := protovalidatetrevrpc.New(nil).ValidateFunc(desc)
	if err != nil {
		t.Fatalf("compile rules of %s: %v", desc.FullName(), err)
	}
	if check == nil {
		t.Fatalf("%s has no rules", desc.FullName())
	}
	return check(req)
}

func seedUser(t *testing.T, db *sql.DB, username, password string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if _, err := db.ExecContext(t.Context(), "INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hash); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func TestSignupCreatesNormalizedUser(t *testing.T) {
	handler, db, _ := newAuthTest(t)

	_, err := handler.Signup(context.Background(), authv1.SignupRequest_builder{
		Username: new("Alice"),
		Password: new("correct horse battery staple"),
	}.Build())
	if err != nil {
		t.Fatalf("Signup() error = %v", err)
	}

	var username, passwordHash string
	if err := db.QueryRowContext(t.Context(), "SELECT username, password_hash FROM users").Scan(&username, &passwordHash); err != nil {
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
	handler, _, _ := newAuthTest(t)
	request := authv1.SignupRequest_builder{
		Username: new("Alice"),
		Password: new("password"),
	}.Build()
	if _, err := handler.Signup(context.Background(), request); err != nil {
		t.Fatalf("first Signup() error = %v", err)
	}
	_, err := handler.Signup(context.Background(), authv1.SignupRequest_builder{
		Username: new("ALICE"),
		Password: new("password"),
	}.Build())
	if got := codeOf(err); got != trevrpc.CodeAlreadyExists {
		t.Errorf("second Signup() code = %v, want AlreadyExists (error %v)", got, err)
	}
}

func TestCredentialsValidation(t *testing.T) {
	for _, test := range []struct {
		name     string
		username *string
		password *string
	}{
		{name: "missing username", password: new("password")},
		{name: "short username", username: new("al"), password: new("password")},
		{name: "long username", username: new(strings.Repeat("a", 65)), password: new("password")},
		{name: "username with spaces", username: new(" alice "), password: new("password")},
		{name: "username starting with a hyphen", username: new("-alice"), password: new("password")},
		{name: "missing password", username: new("alice")},
		{name: "short password", username: new("alice"), password: new("short")},
		{name: "password over 72 bytes", username: new("alice"), password: new(strings.Repeat("é", 37))},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validate(t, authv1.SignupRequest_builder{
				Username: test.username,
				Password: test.password,
			}.Build())
			if got := codeOf(err); got != trevrpc.CodeInvalidArgument {
				t.Errorf("SignupRequest code = %v, want InvalidArgument (error %v)", got, err)
			}
			err = validate(t, authv1.LoginRequest_builder{
				Username: test.username,
				Password: test.password,
			}.Build())
			if got := codeOf(err); got != trevrpc.CodeInvalidArgument {
				t.Errorf("LoginRequest code = %v, want InvalidArgument (error %v)", got, err)
			}
		})
	}
}

func TestCredentialsValidationAcceptsBounds(t *testing.T) {
	for _, test := range []struct {
		name     string
		username string
		password string
	}{
		{name: "shortest", username: "al1", password: "12345678"},
		{name: "longest", username: "a" + strings.Repeat("_-9", 21), password: strings.Repeat("é", 36)},
		// Passwords are measured in characters for the minimum.
		{name: "multibyte password", username: "alice", password: strings.Repeat("é", 8)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := validate(t, authv1.SignupRequest_builder{
				Username: new(test.username),
				Password: new(test.password),
			}.Build()); err != nil {
				t.Errorf("SignupRequest error = %v, want nil", err)
			}
			if err := validate(t, authv1.LoginRequest_builder{
				Username: new(test.username),
				Password: new(test.password),
			}.Build()); err != nil {
				t.Errorf("LoginRequest error = %v, want nil", err)
			}
		})
	}
}

func TestLoginIssuesSession(t *testing.T) {
	handler, db, manager := newAuthTest(t)
	password := "correct horse battery staple"
	seedUser(t, db, "alice", password)

	before := time.Now()
	response, err := handler.Login(context.Background(), authv1.LoginRequest_builder{
		Username: new("ALICE"),
		Password: new(password),
	}.Build())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	username, err := manager.Verify(t.Context(), response.GetToken())
	if err != nil {
		t.Fatalf("Verify(returned token) error = %v", err)
	}
	if username != "alice" || response.GetSub() != "alice" {
		t.Errorf("username = (%q, %q), want normalized alice", username, response.GetSub())
	}
	expires := time.Unix(response.GetExp(), 0)
	if earliest := before.Add(24 * time.Hour).Add(-time.Second); expires.Before(earliest) || expires.After(time.Now().Add(24*time.Hour)) {
		t.Errorf("Exp = %v, want 24 hours after sign-in", expires)
	}
}

// logoutContext returns the context of a Logout call with metadata.
func logoutContext(t *testing.T, metadata trevrpc.Metadata) context.Context {
	return trevrpc.NewHandlerContext(t.Context(), trevrpc.HandlerContext{
		FullMethod: authv1.AuthService_Logout_FullMethodName,
		Metadata:   metadata,
	})
}

func login(t *testing.T, handler *authhandler.Handler, username, password string) string {
	t.Helper()
	response, err := handler.Login(t.Context(), authv1.LoginRequest_builder{
		Username: new(username),
		Password: new(password),
	}.Build())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	return response.GetToken()
}

func TestLogoutEndsTheCallsSession(t *testing.T) {
	handler, db, manager := newAuthTest(t)
	seedUser(t, db, "alice", "correct horse battery staple")
	ended := login(t, handler, "alice", "correct horse battery staple")
	other := login(t, handler, "alice", "correct horse battery staple")

	ctx := logoutContext(t, trevrpc.Metadata{"authorization": []byte("Bearer " + ended)})
	if _, err := handler.Logout(ctx, &authv1.LogoutRequest{}); err != nil {
		t.Fatalf("Logout() error = %v, want nil", err)
	}
	if _, err := manager.Verify(t.Context(), ended); !errors.Is(err, auth.ErrInvalidSession) {
		t.Errorf("Verify(ended token) error = %v, want ErrInvalidSession", err)
	}
	// Signing out on one device leaves the others signed in.
	if _, err := manager.Verify(t.Context(), other); err != nil {
		t.Errorf("Verify(other token) error = %v, want nil", err)
	}
}

func TestLogoutWithoutSessionSucceeds(t *testing.T) {
	handler, _, _ := newAuthTest(t)

	for name, metadata := range map[string]trevrpc.Metadata{
		"no token":        nil,
		"malformed token": {"authorization": []byte("Basic abc")},
		"unknown token":   {"authorization": []byte("Bearer not-a-session")},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := handler.Logout(logoutContext(t, metadata), &authv1.LogoutRequest{}); err != nil {
				t.Errorf("Logout() error = %v, want nil", err)
			}
		})
	}
}

func TestLoginRejectsInvalidCredentials(t *testing.T) {
	handler, db, _ := newAuthTest(t)
	seedUser(t, db, "alice", "correct password")

	for _, test := range []struct {
		name     string
		username string
		password string
	}{
		{name: "unknown user", username: "missing", password: "correct password"},
		{name: "wrong password", username: "alice", password: "wrong password"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := handler.Login(context.Background(), authv1.LoginRequest_builder{
				Username: new(test.username),
				Password: new(test.password),
			}.Build())
			if got := codeOf(err); got != trevrpc.CodeUnauthenticated {
				t.Errorf("code = %v, want Unauthenticated (error %v)", got, err)
			}
		})
	}
}
