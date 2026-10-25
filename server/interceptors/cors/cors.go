// Package cors allows cross-origin Connect requests.
package cors

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

// WithCORS wraps h to allow Connect requests from any origin.
func WithCORS(h http.Handler) http.Handler {
	middleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: append(connectcors.ExposedHeaders(), "Retry-After"),
	})

	return middleware.Handler(h)
}
