package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		logLevel     string
		port         string
		wantLogLevel string
		wantPort     string
	}{
		{
			name:         "defaults",
			wantLogLevel: defaultLogLevel,
			wantPort:     defaultPort,
		},
		{
			name:         "environment",
			logLevel:     "warn",
			port:         "9090",
			wantLogLevel: "warn",
			wantPort:     "9090",
		},
		{
			name:         "flags",
			args:         []string{"--log-level", "debug", "--port", "7070"},
			wantLogLevel: "debug",
			wantPort:     "7070",
		},
		{
			name:         "equals flags",
			args:         []string{"--log-level=error", "--port=6060"},
			wantLogLevel: "error",
			wantPort:     "6060",
		},
		{
			name:         "flags override environment",
			args:         []string{"--log-level", "debug", "--port", "7070"},
			logLevel:     "warn",
			port:         "9090",
			wantLogLevel: "debug",
			wantPort:     "7070",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			cfg, err := parseConfig(test.args, func(key string) string {
				switch key {
				case "LOG_LEVEL":
					return test.logLevel
				case "PORT":
					return test.port
				default:
					return ""
				}
			}, &output)
			if err != nil {
				t.Fatalf("parseConfig() error = %v", err)
			}
			if cfg.logLevel != test.wantLogLevel {
				t.Errorf("logLevel = %q, want %q", cfg.logLevel, test.wantLogLevel)
			}
			if cfg.port != test.wantPort {
				t.Errorf("port = %q, want %q", cfg.port, test.wantPort)
			}
		})
	}
}

func TestParseConfigHelp(t *testing.T) {
	for _, arg := range []string{"-h", "--help"} {
		t.Run(arg, func(t *testing.T) {
			var output bytes.Buffer
			_, err := parseConfig([]string{arg}, func(string) string { return "" }, &output)
			if !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("parseConfig() error = %v, want %v", err, flag.ErrHelp)
			}

			for _, want := range []string{"Usage: server", "--log-level", "--port", "--help", "LOG_LEVEL", "PORT", defaultLogLevel, defaultPort} {
				if !strings.Contains(output.String(), want) {
					t.Errorf("help output does not contain %q:\n%s", want, output.String())
				}
			}
		})
	}
}

func TestParseConfigRejectsInvalidFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "unknown flag",
			args: []string{"--unknown"},
		},
		{
			name: "missing log level value",
			args: []string{"--log-level"},
		},
		{
			name: "missing port value",
			args: []string{"--port"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			_, err := parseConfig(test.args, func(string) string { return "" }, &output)
			if err == nil {
				t.Fatal("parseConfig() error = nil, want non-nil")
			}
			if errors.Is(err, flag.ErrHelp) {
				t.Errorf("parseConfig() error = %v, want non-help error", err)
			}
		})
	}
}
