//go:build dev

package logger

import (
	"io"
	"log/slog"

	"github.com/Marlliton/slogpretty"
)

func newHandler(out io.Writer, level slog.Level) slog.Handler {
	opts := slogpretty.DefaultOptions()
	opts.Level = level

	return slogpretty.New(out, opts)
}
