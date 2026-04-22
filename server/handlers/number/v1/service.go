package v1

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/spotdemo4/trevstack/server/connect/number/v1/numberv1connect"
)

type Handler struct{}

func New(opt ...connect.HandlerOption) (string, http.Handler) {
	return numberv1connect.NewNumberServiceHandler(
		&Handler{},
		opt...,
	)
}
