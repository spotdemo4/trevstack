package web

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"trev.zip/trev/stack/server/logger"
)

func New(ctx context.Context, webfs embed.FS) http.Handler {
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

	fileServer := http.FileServer(http.FS(web))

	// If the requested path doesn't exist, serve index.html instead
	// This allows the frontend router to handle the request
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" && !exists(web, path) {
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}

func exists(fsys fs.FS, path string) bool {
	_, err := fs.Stat(fsys, path)
	return !errors.Is(err, os.ErrNotExist)
}
