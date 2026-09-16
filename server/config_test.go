package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

const testJWTSecret = "01234567890123456789012345678901"

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		logLevel         string
		port             string
		jwtSecret        string
		authCookieSecure string
		wantLogLevel     string
		wantPort         string
		wantCookieSecure bool
	}{
		{
			name:         "defaults",
			jwtSecret:    testJWTSecret,
			wantLogLevel: defaultLogLevel,
			wantPort:     defaultPort,
		},
		{
			name:         "environment",
			logLevel:     "warn",
			port:         "9090",
			jwtSecret:    testJWTSecret,
			wantLogLevel: "warn",
			wantPort:     "9090",
		},
		{
			name:             "secure cookie environment",
			jwtSecret:        testJWTSecret,
			authCookieSecure: "true",
			wantLogLevel:     defaultLogLevel,
			wantPort:         defaultPort,
			wantCookieSecure: true,
		},
		{
			name:         "flags",
			args:         []string{"--log-level", "debug", "--port", "7070"},
			jwtSecret:    testJWTSecret,
			wantLogLevel: "debug",
			wantPort:     "7070",
		},
		{
			name:         "equals flags",
			args:         []string{"--log-level=error", "--port=6060"},
			jwtSecret:    testJWTSecret,
			wantLogLevel: "error",
			wantPort:     "6060",
		},
		{
			name:         "flags override environment",
			args:         []string{"--log-level", "debug", "--port", "7070"},
			logLevel:     "warn",
			port:         "9090",
			jwtSecret:    testJWTSecret,
			wantLogLevel: "debug",
			wantPort:     "7070",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			cfg, err := parseConfig(test.args, func(key string) string {
				switch key {
				case "AUTH_COOKIE_SECURE":
					return test.authCookieSecure
				case "JWT_SECRET":
					return test.jwtSecret
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
			if cfg.authCookieSecure != test.wantCookieSecure {
				t.Errorf("authCookieSecure = %t, want %t", cfg.authCookieSecure, test.wantCookieSecure)
			}
			if cfg.jwtSecret != testJWTSecret {
				t.Errorf("jwtSecret = %q, want test secret", cfg.jwtSecret)
			}
		})
	}
}

func TestParseConfigRejectsMissingOrShortJWTSecret(t *testing.T) {
	for _, test := range []struct {
		name   string
		secret string
		want   string
	}{
		{name: "missing", want: "JWT_SECRET is required"},
		{name: "short", secret: "too-short", want: "JWT_SECRET must be at least 32 bytes"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			_, err := parseConfig(nil, func(key string) string {
				if key == "JWT_SECRET" {
					return test.secret
				}
				return ""
			}, &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("parseConfig() error = %v, want %q", err, test.want)
			}
			if !strings.Contains(output.String(), test.want) {
				t.Errorf("output does not contain %q: %s", test.want, output.String())
			}
		})
	}
}

func TestParseConfigRejectsInvalidCookieSecure(t *testing.T) {
	var output bytes.Buffer
	_, err := parseConfig(nil, func(key string) string {
		if key == "AUTH_COOKIE_SECURE" {
			return "sometimes"
		}
		if key == "JWT_SECRET" {
			return testJWTSecret
		}
		return ""
	}, &output)
	if err == nil || !strings.Contains(err.Error(), "invalid AUTH_COOKIE_SECURE value") {
		t.Fatalf("parseConfig() error = %v, want invalid cookie secure error", err)
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

			for _, want := range []string{"Usage: server", "--log-level", "--port", "--help", "AUTH_COOKIE_SECURE", "JWT_SECRET", "LOG_LEVEL", "PORT", supportedLogLevels, defaultLogLevel, defaultPort} {
				if !strings.Contains(output.String(), want) {
					t.Errorf("help output does not contain %q:\n%s", want, output.String())
				}
			}
		})
	}
}

func TestParseConfigRejectsInvalidFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		logLevel   string
		wantOutput string
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
			name:       "unsupported log level flag",
			args:       []string{"--log-level", "trace"},
			wantOutput: supportedLogLevels,
		},
		{
			name:       "unsupported log level environment",
			logLevel:   "trace",
			wantOutput: supportedLogLevels,
		},
		{
			name: "missing port value",
			args: []string{"--port"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			_, err := parseConfig(test.args, func(key string) string {
				switch key {
				case "JWT_SECRET":
					return testJWTSecret
				case "LOG_LEVEL":
					return test.logLevel
				default:
					return ""
				}
			}, &output)
			if err == nil {
				t.Fatal("parseConfig() error = nil, want non-nil")
			}
			if errors.Is(err, flag.ErrHelp) {
				t.Errorf("parseConfig() error = %v, want non-help error", err)
			}
			if test.wantOutput != "" && !strings.Contains(output.String(), test.wantOutput) {
				t.Errorf("output does not contain %q:\n%s", test.wantOutput, output.String())
			}
		})
	}
}

func TestParseConfigHelpIgnoresInvalidEnvironment(t *testing.T) {
	for _, arg := range []string{"-h", "--help"} {
		t.Run(arg, func(t *testing.T) {
			var output bytes.Buffer
			_, err := parseConfig([]string{arg}, func(key string) string {
				switch key {
				case "AUTH_COOKIE_SECURE":
					return "not-a-bool"
				case "JWT_SECRET":
					return ""
				default:
					return ""
				}
			}, &output)
			if !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("parseConfig() error = %v, want %v", err, flag.ErrHelp)
			}
		})
	}
}
