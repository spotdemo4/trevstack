// Package v1 implements the auth.v1 AuthService.
//
// Requests reach the handlers only after the server has checked them against
// the buf.validate rules in their protos, so the handlers rely on those rules.
package v1

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/template/stack/server/v3/auth"
	authv1 "trev.zip/template/stack/server/v3/gen/auth/v1"
)

var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("invalid-password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}()

// Handler implements authv1.AuthServiceServer.
type Handler struct {
	db   *sql.DB
	auth *auth.Manager
}

// New returns a Handler that stores users in db and issues sessions with
// manager.
func New(db *sql.DB, manager *auth.Manager) *Handler {
	return &Handler{db: db, auth: manager}
}

// Register registers the AuthService on server.
func Register(server *trevrpc.Server, db *sql.DB, manager *auth.Manager) {
	authv1.RegisterAuthServiceServer(server, New(db, manager))
}
