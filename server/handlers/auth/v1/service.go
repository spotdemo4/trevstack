package v1

import (
	"net/http"

	"connectrpc.com/connect"
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

type Handler struct {
	auth *auth.Manager
}

func New(manager *auth.Manager, opt ...connect.HandlerOption) (string, http.Handler) {
	return authv1connect.NewAuthServiceHandler(
		&Handler{auth: manager},
		opt...,
	)
}
