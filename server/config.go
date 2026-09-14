package main

import (
	"flag"
	"fmt"
	"io"
)

const defaultPort = "8080"

type config struct {
	port string
}

func parseConfig(args []string, getenv func(string) string, output io.Writer) (config, error) {
	cfg := config{port: getenv("PORT")}
	if cfg.port == "" {
		cfg.port = defaultPort
	}

	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&cfg.port, "port", cfg.port, "port to listen on")
	flags.Usage = func() {
		fmt.Fprintln(output, "Usage: server [--port PORT]")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Options:")
		fmt.Fprintln(output, "  -h, --help    Show this help menu.")
		fmt.Fprintln(output, "  --port PORT   Port to listen on.")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Environment:")
		fmt.Fprintf(output, "  PORT          Fallback port (default: %s).\n", defaultPort)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Precedence: --port, PORT, then 8080.")
	}

	if err := flags.Parse(args); err != nil {
		return config{}, err
	}

	return cfg, nil
}
