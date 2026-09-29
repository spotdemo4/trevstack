package web

import (
	"io/fs"
	"net/http"
	"strings"
)

// assetsDir holds the build's content-hashed files, which never change.
const assetsDir = "assets/"

func New(web fs.FS) http.Handler {
	if web == nil {
		return http.NotFoundHandler()
	}

	fileServer := http.FileServer(http.FS(web))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		if strings.HasPrefix(path, assetsDir) {
			// A missing asset is a chunk from an older build; answering with
			// index.html would cache HTML under the chunk's name.
			if !exists(web, path) {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			fileServer.ServeHTTP(w, r)
			return
		}

		// Everything else, including the service worker and index.html, must
		// be revalidated so a new deploy is noticed.
		w.Header().Set("Cache-Control", "no-cache")

		// If the requested path doesn't exist, serve index.html instead
		// This allows the frontend router to handle the request
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
