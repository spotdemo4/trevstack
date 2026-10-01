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

type recordingHandler struct {
	records *[]slog.Record
}

func (h recordingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h recordingHandler) WithGroup(string) slog.Handler { return h }

func (h recordingHandler) Handle(_ context.Context, r slog.Record) error {
	*h.records = append(*h.records, r)
	return nil
}

func TestLevelHandler(t *testing.T) {
	var records []slog.Record
	inner := recordingHandler{records: &records}
	logger := slog.New(&levelHandler{Handler: inner, level: slog.LevelWarn})

	logger.Info("ignored")
	logger.With("component", "api").Warn("included")

	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	if got := records[0].Message; got != "included" {
		t.Errorf("message = %q, want included", got)
	}
}
