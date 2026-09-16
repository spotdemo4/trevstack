package logger

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func TestNewLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		enabled     slog.Level
		disabled    slog.Level
		hasDisabled bool
	}{
		{name: "debug", value: "debug", enabled: slog.LevelDebug},
		{name: "debug case insensitive", value: "DEBUG", enabled: slog.LevelDebug},
		{name: "info", value: "info", enabled: slog.LevelInfo, disabled: slog.LevelDebug, hasDisabled: true},
		{name: "warn", value: "warn", enabled: slog.LevelWarn, disabled: slog.LevelInfo, hasDisabled: true},
		{name: "error", value: "error", enabled: slog.LevelError, disabled: slog.LevelWarn, hasDisabled: true},
		{name: "empty defaults to info", enabled: slog.LevelInfo, disabled: slog.LevelDebug, hasDisabled: true},
		{name: "unknown defaults to info", value: "trace", enabled: slog.LevelInfo, disabled: slog.LevelDebug, hasDisabled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.value)
			ctx := context.Background()

			if !logger.Enabled(ctx, tt.enabled) {
				t.Fatalf("expected level %s to be enabled", tt.enabled)
			}
			if tt.hasDisabled && logger.Enabled(ctx, tt.disabled) {
				t.Fatalf("expected level %s to be disabled", tt.disabled)
			}
		})
	}
}

func TestContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := WithLog(context.Background(), logger)

	if got := FromContext(ctx); got != logger {
		t.Fatal("expected logger from context")
	}
	if got := FromContext(context.Background()); got == nil {
		t.Fatal("expected fallback logger for context without logger")
	}
	//lint:ignore SA1012 FromContext explicitly supports nil contexts.
	if got := FromContext(nil); got == nil {
		t.Fatal("expected fallback logger for nil context")
	}
}
