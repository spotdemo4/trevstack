package web

import (
	"io/fs"
	"net/http"
	"strings"
)

func New(web fs.FS) http.Handler {
	if web == nil {
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
	return err == nil
}
