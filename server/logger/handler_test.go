//go:build !dev

package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestJSONHandler(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(newHandler(&out, slog.LevelInfo))
	logger.Info("server started", "component", "api")

	var record map[string]any
	if err := json.Unmarshal(out.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}

	if got := record["level"]; got != "INFO" {
		t.Errorf("level = %v, want INFO", got)
	}
	if got := record["msg"]; got != "server started" {
		t.Errorf("msg = %v, want server started", got)
	}
	if got := record["component"]; got != "api" {
		t.Errorf("component = %v, want api", got)
	}
}

func TestJSONHandlerLevel(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(newHandler(&out, slog.LevelWarn))
	logger.Info("ignored")
	logger.Warn("included")

	var record map[string]any
	if err := json.Unmarshal(out.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if got := record["msg"]; got != "included" {
		t.Errorf("msg = %v, want included", got)
	}
}
