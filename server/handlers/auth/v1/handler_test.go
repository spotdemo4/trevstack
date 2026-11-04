package v1_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/template/stack/server/auth"
	"trev.zip/template/stack/server/database"
	authv1 "trev.zip/template/stack/server/gen/auth/v1"
	authhandler "trev.zip/template/stack/server/handlers/auth/v1"
)

const handlerTestSecret = "01234567890123456789012345678901"

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

	manager := auth.NewManager(handlerTestSecret)
	return authhandler.New(db, manager), db, manager
}

func codeOf(err error) trevrpc.Code {
	return trevrpc.StatusFromError(err).Code
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
	handler, db, _ := newAuthTest(t)
	seedUser(t, db, "alice", "correct horse battery staple")

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
			_, err := handler.Signup(context.Background(), authv1.SignupRequest_builder{
				Username: test.username,
				Password: test.password,
			}.Build())
			if got := codeOf(err); got != trevrpc.CodeInvalidArgument {
				t.Errorf("Signup() code = %v, want InvalidArgument (error %v)", got, err)
			}
			_, err = handler.Login(context.Background(), authv1.LoginRequest_builder{
				Username: test.username,
				Password: test.password,
			}.Build())
			if got := codeOf(err); got != trevrpc.CodeInvalidArgument {
				t.Errorf("Login() code = %v, want InvalidArgument (error %v)", got, err)
			}
		})
	}
}

func TestLoginIssuesJWT(t *testing.T) {
	handler, db, manager := newAuthTest(t)
	password := "correct horse battery staple"
	seedUser(t, db, "alice", password)

	response, err := handler.Login(context.Background(), authv1.LoginRequest_builder{
		Username: new("ALICE"),
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
	if response.GetExp() != claims.ExpiresAt.Unix() {
		t.Errorf("response Exp = %d, want token expiration %d", response.GetExp(), claims.ExpiresAt.Unix())
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
