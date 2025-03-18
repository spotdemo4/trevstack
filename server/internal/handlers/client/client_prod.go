// go:build !dev

package client

import (
	"embed"
	"net/http"
)

//go:embed client
var eclient embed.FS

func init() {
	fs := http.FS(eclient)
	embedfs = &fs
}
