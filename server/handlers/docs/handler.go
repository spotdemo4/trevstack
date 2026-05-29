package docs

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

	scalargo "github.com/bdpiprava/scalar-go"
	"trev.zip/llc/stack/server/logger"
)

//go:embed openapi.yaml
var apiSpec []byte

func New(ctx context.Context) http.Handler {
	log := logger.FromContext(ctx)

	html, err := scalargo.NewV2(
		scalargo.WithSpecBytes(apiSpec),
	)
	if err != nil {
		log.ErrorContext(ctx, "failed to generate API documentation", "error", err)
		return http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, html)
	})
}
