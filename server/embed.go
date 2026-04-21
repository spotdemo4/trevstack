//go:build !dev

package main

import "embed"

//go:embed all:web
var webfs embed.FS

func init() {
	WebFS = webfs
}
