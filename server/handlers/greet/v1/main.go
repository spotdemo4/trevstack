package v1

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	greetv1 "github.com/spotdemo4/trevstack/server/connect/greet/v1"
	"github.com/spotdemo4/trevstack/server/connect/greet/v1/greetv1connect"
)

type Handler struct{}

func New(opt connect.HandlerOption) (string, http.Handler) {
	return greetv1connect.NewGreetServiceHandler(
		&Handler{},
		opt,
	)
}

func (h *Handler) Greet(
	ctx context.Context,
	req *greetv1.GreetRequest,
) (*greetv1.GreetResponse, error) {

	greeting := "Hello, " + req.Name + "!"
	response := &greetv1.GreetResponse{
		Greeting: greeting,
	}

	return response, nil
}
