package v1

import (
	"connectrpc.com/connect/v2"
	"trev.zip/template/stack/server/connect/number/v1/numberv1connect"
)

type Handler struct{}

func Register(server *connect.Server) {
	numberv1connect.RegisterNumberServiceHandler(server, &Handler{})
}
