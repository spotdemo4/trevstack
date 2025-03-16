package handlers

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/spotdemo4/trevstack/server/internal/interceptors"
)

func NewClientHandler(client embed.FS, key string) http.Handler {
	clientFs, err := fs.Sub(client, "client")
	if err != nil {
		log.Fatalf("failed to get sub filesystem: %v", err)
	}

	return interceptors.WithAuthRedirect(http.FileServer(http.FS(clientFs)), key)
}
