package v1

import (
	"net/http"

	"connectrpc.com/connect"
	"trev.zip/trev/stack/server/connect/number/v1/numberv1connect"
)

type Handler struct{}

func New(opt ...connect.HandlerOption) (string, http.Handler) {
	return numberv1connect.NewNumberServiceHandler(
		&Handler{},
		opt...,
	)
}
