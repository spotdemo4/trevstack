package docs

import (
	"io/fs"
	"net/http"
	"strings"
)

func New(docs fs.FS) http.Handler {
	if docs == nil {
		return http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/docs/")
		if path == r.URL.Path {
			http.NotFound(w, r)
			return
		}
		if path == "" {
			path = "index.html"
		}
		if !fs.ValidPath(path) {
			http.NotFound(w, r)
			return
		}

		info, err := fs.Stat(docs, path)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		http.ServeFileFS(w, r, docs, path)
	})
}
