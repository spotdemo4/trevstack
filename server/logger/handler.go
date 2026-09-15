//go:build !dev

package logger

import (
	"io"
	"log/slog"
)

func newHandler(out io.Writer, level slog.Level) slog.Handler {
	return slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: level,
	})
}
