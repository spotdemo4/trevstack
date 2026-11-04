package main

import (
	"flag"
	"fmt"
	"io"
	"net/netip"
	"net/url"
	"strings"
)

const (
	defaultLogLevel    = "info"
	defaultPort        = "8080"
	minJWTSecretLength = 32
	supportedLogLevels = "debug, info, warn, error"
)

const usage = `Usage: server [--log-level LEVEL] [--port PORT] [--tls-cert FILE --tls-key FILE]

Options:
  -h, --help          Show this help menu.
  --log-level LEVEL   Minimum log level (%s).
  --port PORT         UDP and TCP port to listen on.
  --tls-cert FILE     PEM certificate chain; requires --tls-key.
  --tls-key FILE      PEM private key; requires --tls-cert.

Environment:
  ALLOWED_ORIGINS     Comma-separated origins of other sites whose pages may
                      call the API, such as https://app.example.com. Builds
                      with the dev tag also admit http://localhost:3000.
  JWT_SECRET          JWT signing secret (required, at least 32 bytes).
  LOG_LEVEL           Fallback log level (default: %s).
  PORT                Fallback port (default: %s).
  TLS_CERT            Fallback certificate file.
  TLS_KEY             Fallback private key file.
  TRUSTED_PROXY_CIDRS Comma-separated CIDRs of TCP proxies that send a PROXY
                      protocol header with every connection.
  OTEL_*              OpenTelemetry SDK settings; signals export only when an
                      OTLP endpoint or OTEL_{TRACES,METRICS,LOGS}_EXPORTER is set.

Precedence: --log-level, LOG_LEVEL, then info.
            --port, PORT, then 8080.
            --tls-cert and --tls-key, TLS_CERT and TLS_KEY, then a generated
            self-signed certificate for localhost.
`

type config struct {
	allowedOrigins    []string
	jwtSecret         string
	logLevel          string
	port              string
	tlsCert           string
	tlsKey            string
	trustedProxyCIDRs []netip.Prefix
}

func parseConfig(args []string, getenv func(string) string, output io.Writer) (config, error) {
	cfg := config{
		jwtSecret: getenv("JWT_SECRET"),
		logLevel:  getenv("LOG_LEVEL"),
		port:      getenv("PORT"),
		tlsCert:   getenv("TLS_CERT"),
		tlsKey:    getenv("TLS_KEY"),
	}
	if cfg.logLevel == "" {
		cfg.logLevel = defaultLogLevel
	}
	if cfg.port == "" {
		cfg.port = defaultPort
	}
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&cfg.logLevel, "log-level", cfg.logLevel, "minimum log level")
	flags.StringVar(&cfg.port, "port", cfg.port, "port to listen on")
	flags.StringVar(&cfg.tlsCert, "tls-cert", cfg.tlsCert, "certificate file")
	flags.StringVar(&cfg.tlsKey, "tls-key", cfg.tlsKey, "private key file")
	flags.Usage = func() {
		_, _ = fmt.Fprintf(output, usage, supportedLogLevels, defaultLogLevel, defaultPort)
	}

	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if value := getenv("ALLOWED_ORIGINS"); value != "" {
		allowedOrigins, err := parseAllowedOrigins(value)
		if err != nil {
			return config{}, err
		}
		cfg.allowedOrigins = allowedOrigins
	}
	if value := getenv("TRUSTED_PROXY_CIDRS"); value != "" {
		trustedProxyCIDRs, err := parseTrustedProxyCIDRs(value)
		if err != nil {
			return config{}, err
		}
		cfg.trustedProxyCIDRs = trustedProxyCIDRs
	}
	if (cfg.tlsCert == "") != (cfg.tlsKey == "") {
		err := fmt.Errorf("a TLS certificate and key must be set together")
		_, _ = fmt.Fprintln(output, err)
		return config{}, err
	}
	if !isSupportedLogLevel(cfg.logLevel) {
		err := fmt.Errorf("unsupported log level %q (supported: %s)", cfg.logLevel, supportedLogLevels)
		_, _ = fmt.Fprintln(output, err)
		return config{}, err
	}
	if cfg.jwtSecret == "" {
		err := fmt.Errorf("JWT_SECRET is required")
		_, _ = fmt.Fprintln(output, err)
		return config{}, err
	}
	if len(cfg.jwtSecret) < minJWTSecretLength {
		err := fmt.Errorf("JWT_SECRET must be at least %d bytes", minJWTSecretLength)
		_, _ = fmt.Fprintln(output, err)
		return config{}, err
	}

	return cfg, nil
}

func parseTrustedProxyCIDRs(value string) ([]netip.Prefix, error) {
	parts := strings.Split(value, ",")
	prefixes := make([]netip.Prefix, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		prefix, err := netip.ParsePrefix(part)
		if err != nil {
			return nil, fmt.Errorf("invalid TRUSTED_PROXY_CIDRS value %q: %w", part, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

// parseAllowedOrigins parses comma-separated origins, such as
// http://localhost:3000, in the form browsers send them.
func parseAllowedOrigins(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		origin, err := url.Parse(part)
		if err != nil || (origin.Scheme != "http" && origin.Scheme != "https") || origin.Host == "" ||
			origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
			return nil, fmt.Errorf("invalid ALLOWED_ORIGINS value %q: want scheme://host[:port]", part)
		}
		origins = append(origins, strings.ToLower(origin.Scheme)+"://"+strings.ToLower(origin.Host))
	}
	return origins, nil
}

func isSupportedLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
