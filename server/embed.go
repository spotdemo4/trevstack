//go:build !dev

package main

import (
	"embed"
	"io/fs"
)

//go:embed all:web
var webfs embed.FS

func init() {
	sub, err := fs.Sub(webfs, "web")
	if err != nil {
		panic(err)
	}
	WebFS = sub
}
