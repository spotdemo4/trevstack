package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseConfigPort(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  string
		want string
	}{
		{
			name: "default",
			want: defaultPort,
		},
		{
			name: "empty environment",
			env:  "",
			want: defaultPort,
		},
		{
			name: "environment",
			env:  "9090",
			want: "9090",
		},
		{
			name: "flag",
			args: []string{"--port", "7070"},
			want: "7070",
		},
		{
			name: "equals flag",
			args: []string{"--port=6060"},
			want: "6060",
		},
		{
			name: "flag overrides environment",
			args: []string{"--port", "7070"},
			env:  "9090",
			want: "7070",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			cfg, err := parseConfig(test.args, func(key string) string {
				if key == "PORT" {
					return test.env
				}
				return ""
			}, &output)
			if err != nil {
				t.Fatalf("parseConfig() error = %v", err)
			}
			if cfg.port != test.want {
				t.Errorf("port = %q, want %q", cfg.port, test.want)
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

			for _, want := range []string{"Usage: server", "--port", "--help", "PORT", defaultPort} {
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
