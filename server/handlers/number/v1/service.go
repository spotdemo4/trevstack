// Package v1 implements the number.v1 NumberService.
package v1

import (
	"connectrpc.com/connect/v2"
	"trev.zip/template/stack/server/connect/number/v1/numberv1connect"
)

// Handler implements numberv1connect.NumberServiceHandler.
type Handler struct{}

// Register registers the NumberService on server.
func Register(server *connect.Server) {
	numberv1connect.RegisterNumberServiceHandler(server, &Handler{})
}
