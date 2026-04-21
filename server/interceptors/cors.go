package interceptors

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

func WithCORS(h http.Handler) http.Handler {
	middleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})

	return middleware.Handler(h)
}
