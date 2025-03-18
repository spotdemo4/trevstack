//go:build !dev

package client

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:client
var eclient embed.FS

func init() {
	log.Println("Initializing client for production")
	client, err := fs.Sub(eclient, "client")
	if err != nil {
		log.Fatalf("failed to get client: %v", err)
	}

	fs := http.FS(client)
	embedfs = &fs
}
