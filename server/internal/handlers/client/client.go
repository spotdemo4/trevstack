package client

import (
	"embed"
	"io"
	"io/fs"
	"net/http"

	"github.com/spotdemo4/trevstack/server/internal/interceptors"
)

func NewClientHandler(key string, clientFS embed.FS) http.Handler {
	_, err := clientFS.ReadDir(".")
	if err == io.EOF {
		return http.NotFoundHandler()
	}

	client, err := fs.Sub(clientFS, "client")
	if err != nil {
		return http.NotFoundHandler()
	}

	fs := http.FS(client)
	return interceptors.WithAuthRedirect(http.FileServer(fs), key)
}
