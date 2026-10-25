// Package v1 implements the auth.v1 AuthService.
package v1

import (
	"connectrpc.com/connect/v2"
	"golang.org/x/crypto/bcrypt"
	"trev.zip/template/stack/server/auth"
	"trev.zip/template/stack/server/connect/auth/v1/authv1connect"
)

var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("invalid-password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}()

// Handler implements authv1connect.AuthServiceHandler.
type Handler struct {
	auth *auth.Manager
}

// Register registers the AuthService on server.
func Register(server *connect.Server, manager *auth.Manager) {
	authv1connect.RegisterAuthServiceHandler(server, &Handler{auth: manager})
}
