package main

import (
	"bytes"
	"errors"
	"flag"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name                  string
		args                  []string
		logLevel              string
		port                  string
		allowedOrigins        string
		tlsCert               string
		tlsKey                string
		trustedProxyCIDRs     string
		wantLogLevel          string
		wantPort              string
		wantAllowedOrigins    []string
		wantTLSCert           string
		wantTLSKey            string
		wantTrustedProxyCIDRs []netip.Prefix
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
			name:               "allowed origins environment",
			allowedOrigins:     "http://localhost:3000, HTTPS://App.Example.com",
			wantLogLevel:       defaultLogLevel,
			wantPort:           defaultPort,
			wantAllowedOrigins: []string{"http://localhost:3000", "https://app.example.com"},
		},
		{
			name:         "tls environment",
			tlsCert:      "cert.pem",
			tlsKey:       "key.pem",
			wantLogLevel: defaultLogLevel,
			wantPort:     defaultPort,
			wantTLSCert:  "cert.pem",
			wantTLSKey:   "key.pem",
		},
		{
			name:         "tls flags override environment",
			args:         []string{"--tls-cert", "flag-cert.pem", "--tls-key=flag-key.pem"},
			tlsCert:      "cert.pem",
			tlsKey:       "key.pem",
			wantLogLevel: defaultLogLevel,
			wantPort:     defaultPort,
			wantTLSCert:  "flag-cert.pem",
			wantTLSKey:   "flag-key.pem",
		},
		{
			name:              "trusted proxy environment",
			trustedProxyCIDRs: "10.0.0.1/8, 2001:db8::1/32",
			wantLogLevel:      defaultLogLevel,
			wantPort:          defaultPort,
			wantTrustedProxyCIDRs: []netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("2001:db8::/32"),
			},
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
				case "ALLOWED_ORIGINS":
					return test.allowedOrigins
				case "LOG_LEVEL":
					return test.logLevel
				case "PORT":
					return test.port
				case "TLS_CERT":
					return test.tlsCert
				case "TLS_KEY":
					return test.tlsKey
				case "TRUSTED_PROXY_CIDRS":
					return test.trustedProxyCIDRs
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
			if !slices.Equal(cfg.allowedOrigins, test.wantAllowedOrigins) {
				t.Errorf("allowedOrigins = %v, want %v", cfg.allowedOrigins, test.wantAllowedOrigins)
			}
			if cfg.tlsCert != test.wantTLSCert || cfg.tlsKey != test.wantTLSKey {
				t.Errorf("TLS files = (%q, %q), want (%q, %q)", cfg.tlsCert, cfg.tlsKey, test.wantTLSCert, test.wantTLSKey)
			}
			if !slices.Equal(cfg.trustedProxyCIDRs, test.wantTrustedProxyCIDRs) {
				t.Errorf("trustedProxyCIDRs = %v, want %v", cfg.trustedProxyCIDRs, test.wantTrustedProxyCIDRs)
			}
		})
	}
}

func TestParseConfigRejectsInvalidAllowedOrigins(t *testing.T) {
	for _, value := range []string{"localhost:3000", "http://localhost:3000/", "ftp://example.com", "https://", "http://user@example.com", ""} {
		t.Run(value, func(t *testing.T) {
			var output bytes.Buffer
			_, err := parseConfig(nil, func(key string) string {
				switch key {
				case "ALLOWED_ORIGINS":
					return "http://localhost:3000," + value
				default:
					return ""
				}
			}, &output)
			if err == nil || !strings.Contains(err.Error(), "invalid ALLOWED_ORIGINS value") {
				t.Fatalf("parseConfig() error = %v, want invalid allowed origins error", err)
			}
		})
	}
}

func TestParseConfigRejectsInvalidTrustedProxyCIDRs(t *testing.T) {
	var output bytes.Buffer
	_, err := parseConfig(nil, func(key string) string {
		if key == "TRUSTED_PROXY_CIDRS" {
			return "10.0.0.0/8,not-a-cidr"
		}
		return ""
	}, &output)
	if err == nil || !strings.Contains(err.Error(), "invalid TRUSTED_PROXY_CIDRS value") {
		t.Fatalf("parseConfig() error = %v, want invalid trusted proxy CIDRs error", err)
	}
}

func TestParseConfigRejectsPartialTLSFiles(t *testing.T) {
	for _, args := range [][]string{{"--tls-cert", "cert.pem"}, {"--tls-key", "key.pem"}} {
		t.Run(args[0], func(t *testing.T) {
			var output bytes.Buffer
			_, err := parseConfig(args, func(string) string { return "" }, &output)
			want := "a TLS certificate and key must be set together"
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("parseConfig() error = %v, want %q", err, want)
			}
			if !strings.Contains(output.String(), want) {
				t.Errorf("output does not contain %q: %s", want, output.String())
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

			for _, want := range []string{"Usage: server", "--log-level", "--port", "--tls-cert", "--tls-key", "--help", "ALLOWED_ORIGINS", "LOG_LEVEL", "PORT", "TLS_CERT", "TLS_KEY", "TRUSTED_PROXY_CIDRS", "OTEL_", supportedLogLevels, defaultLogLevel, defaultPort} {
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
				case "ALLOWED_ORIGINS":
					return "not-an-origin"
				case "TLS_CERT":
					return "cert-without-key.pem"
				case "TRUSTED_PROXY_CIDRS":
					return "not-a-cidr"
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

func TestEnsureConfigDirRestrictsPermissions(t *testing.T) {
	// os.UserConfigDir uses XDG_CONFIG_HOME on Linux but HOME on macOS
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	configHome, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir() error = %v", err)
	}
	// An existing directory with looser permissions is tightened.
	want := filepath.Join(configHome, "trevstack")
	if err := os.MkdirAll(want, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	got, err := ensureConfigDir()
	if err != nil {
		t.Fatalf("ensureConfigDir() error = %v", err)
	}
	if got != want {
		t.Errorf("ensureConfigDir() = %q, want %q", got, want)
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("stat %s: %v", got, err)
	}
	if mode := info.Mode().Perm(); mode != 0700 {
		t.Errorf("mode for %s = %04o, want 0700", got, mode)
	}
}
