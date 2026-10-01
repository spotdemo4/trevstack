package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// scope is the instrumentation scope name for logs bridged to OpenTelemetry.
const scope = "trev.zip/template/stack/server"

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

	// Records are also sent to the global OpenTelemetry LoggerProvider, which
	// drops them until telemetry.Setup installs one.
	return slog.New(slog.NewMultiHandler(
		newHandler(os.Stdout, loglevel),
		&levelHandler{Handler: otelslog.NewHandler(scope), level: loglevel},
	))
}

// levelHandler applies the configured minimum level to a handler that has no
// level option of its own.
type levelHandler struct {
	slog.Handler
	level slog.Level
}

func (h *levelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level && h.Handler.Enabled(ctx, level)
}

func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelHandler{Handler: h.Handler.WithAttrs(attrs), level: h.level}
}

func (h *levelHandler) WithGroup(name string) slog.Handler {
	return &levelHandler{Handler: h.Handler.WithGroup(name), level: h.level}
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
