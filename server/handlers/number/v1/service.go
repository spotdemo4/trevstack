// Package v1 implements the number.v1 NumberService.
//
// Requests reach the handlers only after the server has checked them against
// the buf.validate rules in their protos, so the handlers rely on those rules.
package v1

import (
	"database/sql"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	numberv1 "trev.zip/template/stack/server/v2/gen/number/v1"
)

// Handler implements numberv1.NumberServiceServer.
type Handler struct {
	db *sql.DB
}

// New returns a Handler that stores numbers in db.
func New(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// Register registers the NumberService on server.
func Register(server *trevrpc.Server, db *sql.DB) {
	numberv1.RegisterNumberServiceServer(server, New(db))
}
