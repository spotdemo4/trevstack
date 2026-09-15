//go:build dev

package logger

import (
	"bytes"
	"log/slog"
	"regexp"
	"strings"
	"testing"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestPrettyHandler(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(newHandler(&out, slog.LevelInfo))
	logger.Info("server started", "component", "api")

	formatted := out.String()
	if !ansiPattern.MatchString(formatted) {
		t.Fatal("expected colorful ANSI output")
	}

	plain := ansiPattern.ReplaceAllString(formatted, "")
	for _, want := range []string{"INFO", "server started", "component", "api"} {
		if !strings.Contains(plain, want) {
			t.Errorf("output %q does not contain %q", plain, want)
		}
	}
}

func TestPrettyHandlerLevel(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(newHandler(&out, slog.LevelWarn))
	logger.Info("ignored")
	logger.Warn("included")

	plain := ansiPattern.ReplaceAllString(out.String(), "")
	if strings.Contains(plain, "ignored") {
		t.Errorf("output %q contains filtered message", plain)
	}
	if !strings.Contains(plain, "included") {
		t.Errorf("output %q does not contain warning", plain)
	}
}
