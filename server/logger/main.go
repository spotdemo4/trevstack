package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

func New(level string) *slog.Logger {
	var loglevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		loglevel = slog.LevelDebug
	case "error":
		loglevel = slog.LevelError
	case "warn":
		loglevel = slog.LevelWarn
	default:
		loglevel = slog.LevelInfo
	}

	return slog.New(newHandler(os.Stdout, loglevel))
}

type key struct{}

func WithLog(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, key{}, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return New("info")
	}

	logger, ok := ctx.Value(key{}).(*slog.Logger)
	if !ok {
		return New("info")
	}

	return logger
}
