package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

func New() *slog.Logger {
	var loglevel slog.Level

	level := os.Getenv("LOG_LEVEL")
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

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: loglevel,
	}))
}

type key struct{}

func WithLog(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, key{}, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return New()
	}

	logger, ok := ctx.Value(key{}).(*slog.Logger)
	if !ok {
		return New()
	}

	return logger
}
