//go:build !dev

package main

import (
	"embed"
	"io/fs"
)

//go:embed all:web
var webfs embed.FS

//go:embed all:docs
var docsfs embed.FS

func init() {
	DocsFS = mustSub(docsfs, "docs")
	WebFS = mustSub(webfs, "web")
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
