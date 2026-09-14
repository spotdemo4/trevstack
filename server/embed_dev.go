//go:build dev

package main

import "os"

func init() {
	DocsFS = os.DirFS("../docs/dist")
}
