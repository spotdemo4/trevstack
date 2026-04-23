package web

import (
	"context"
	"embed"
	"io/fs"
	"net/http"

	"github.com/spotdemo4/trevstack/server/logger"
)

func Handler(ctx context.Context, webfs embed.FS) http.Handler {
	log := logger.FromContext(ctx)

	entries, err := webfs.ReadDir(".")
	if err != nil || len(entries) == 0 {
		log.WarnContext(ctx, "web build not found")
		return http.NotFoundHandler()
	}

	web, err := fs.Sub(webfs, "web")
	if err != nil {
		log.ErrorContext(ctx, "could not access web build", "error", err)
		return http.NotFoundHandler()
	}

	return http.FileServer(http.FS(web))
}
