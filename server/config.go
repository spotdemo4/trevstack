package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	defaultLogLevel    = "info"
	defaultPort        = "8080"
	minJWTSecretLength = 32
	supportedLogLevels = "debug, info, warn, error"
)

type config struct {
	authCookieSecure bool
	jwtSecret        string
	logLevel         string
	port             string
}

func parseConfig(args []string, getenv func(string) string, output io.Writer) (config, error) {
	cfg := config{
		jwtSecret: getenv("JWT_SECRET"),
		logLevel:  getenv("LOG_LEVEL"),
		port:      getenv("PORT"),
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
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage: server [--log-level LEVEL] [--port PORT]")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Options:")
		fmt.Fprintln(output, "  -h, --help          Show this help menu.")
		fmt.Fprintf(output, "  --log-level LEVEL   Minimum log level (%s).\n", supportedLogLevels)
		fmt.Fprintln(output, "  --port PORT         Port to listen on.")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Environment:")
		fmt.Fprintln(output, "  AUTH_COOKIE_SECURE  Mark the auth cookie Secure (default: false).")
		fmt.Fprintln(output, "  JWT_SECRET          JWT signing secret (required, at least 32 bytes).")
		fmt.Fprintf(output, "  LOG_LEVEL           Fallback log level (default: %s).\n", defaultLogLevel)
		fmt.Fprintf(output, "  PORT                Fallback port (default: %s).\n", defaultPort)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Precedence: --log-level, LOG_LEVEL, then info.")
		fmt.Fprintln(output, "            --port, PORT, then 8080.")
	}

	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if value := getenv("AUTH_COOKIE_SECURE"); value != "" {
		secure, err := strconv.ParseBool(value)
		if err != nil {
			return config{}, fmt.Errorf("invalid AUTH_COOKIE_SECURE value: %w", err)
		}
		cfg.authCookieSecure = secure
	}
	if !isSupportedLogLevel(cfg.logLevel) {
		err := fmt.Errorf("unsupported log level %q (supported: %s)", cfg.logLevel, supportedLogLevels)
		fmt.Fprintln(output, err)
		return config{}, err
	}
	if cfg.jwtSecret == "" {
		err := fmt.Errorf("JWT_SECRET is required")
		fmt.Fprintln(output, err)
		return config{}, err
	}
	if len(cfg.jwtSecret) < minJWTSecretLength {
		err := fmt.Errorf("JWT_SECRET must be at least %d bytes", minJWTSecretLength)
		fmt.Fprintln(output, err)
		return config{}, err
	}

	return cfg, nil
}

func isSupportedLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
