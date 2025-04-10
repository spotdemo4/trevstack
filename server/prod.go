//go:build !dev

package main

import (
	"embed"
	"log"
)

//go:embed all:client
var cFS embed.FS

//go:embed db/migrations/*.sql
var dFS embed.FS

func init() {
	log.Println("initializing for production")
	clientFS = &cFS
	dbFS = &dFS
}
