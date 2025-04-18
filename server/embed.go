//go:build !dev

package embed

import (
	"embed"
)

//go:embed all:client
var ClientFS embed.FS

//go:embed db/migrations/*.sql
var DBFS embed.FS
