package main

import (
	"flag"
	"fmt"
	"io"
)

const (
	defaultLogLevel = "info"
	defaultPort     = "8080"
)

type config struct {
	logLevel string
	port     string
}

func parseConfig(args []string, getenv func(string) string, output io.Writer) (config, error) {
	cfg := config{
		logLevel: getenv("LOG_LEVEL"),
		port:     getenv("PORT"),
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
		fmt.Fprintln(output, "  --log-level LEVEL   Minimum log level.")
		fmt.Fprintln(output, "  --port PORT         Port to listen on.")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Environment:")
		fmt.Fprintf(output, "  LOG_LEVEL           Fallback log level (default: %s).\n", defaultLogLevel)
		fmt.Fprintf(output, "  PORT                Fallback port (default: %s).\n", defaultPort)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Precedence: --log-level, LOG_LEVEL, then info.")
		fmt.Fprintln(output, "            --port, PORT, then 8080.")
	}

	if err := flags.Parse(args); err != nil {
		return config{}, err
	}

	return cfg, nil
}
