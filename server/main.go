// main.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/XSAM/otelsql"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/llc/trevrpc/trevrpc-go/oteltrevrpc"
	"trev.zip/template/stack/server/v2/auth"
	"trev.zip/template/stack/server/v2/database"
	authv1 "trev.zip/template/stack/server/v2/gen/auth/v1"
	authv1handler "trev.zip/template/stack/server/v2/handlers/auth/v1"
	numberv1handler "trev.zip/template/stack/server/v2/handlers/number/v1"
	webhandler "trev.zip/template/stack/server/v2/handlers/web"
	loginterceptor "trev.zip/template/stack/server/v2/interceptors/log"
	ratelimitinterceptor "trev.zip/template/stack/server/v2/interceptors/ratelimit"
	"trev.zip/template/stack/server/v2/logger"
	"trev.zip/template/stack/server/v2/telemetry"
)

const (
	telemetryShutdownTimeout = 5 * time.Second

	// certificateHashesPath publishes the hashes of the development
	// certificate, which browsers trust it by.
	certificateHashesPath = "/certificate-hashes"

	loginRateLimitRequests  = 5
	loginRateLimitWindow    = time.Minute
	signupRateLimitRequests = 3
	signupRateLimitWindow   = time.Hour
)

var (
	WebFS fs.FS

	// devOrigins are admitted along with ALLOWED_ORIGINS. Builds without the prod
	// tag admit the web app's development server.
	devOrigins []string

	// version is set at build time with -ldflags "-X main.version=...".
	version = "dev"
)

func main() {
	cfg, err := parseConfig(os.Args[1:], os.Getenv, os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		os.Exit(2)
	}

	log := logger.New(cfg.logLevel)

	shutdownTelemetry, err := telemetry.Setup(context.Background(), version)
	if err != nil {
		err = fmt.Errorf("could not initialize telemetry: %w", err)
	} else {
		err = run(cfg, log)
	}
	if err != nil {
		log.Error("exiting", "error", err)
	}

	// Flush after the final log so it is exported too. This can't be deferred
	// because os.Exit skips deferred calls.
	ctx, cancel := context.WithTimeout(context.Background(), telemetryShutdownTimeout)
	shutdownErr := shutdownTelemetry(ctx)
	cancel()
	if shutdownErr != nil {
		log.Error("could not shut down telemetry", "error", shutdownErr)
	}

	if err != nil {
		os.Exit(1)
	}
}

func run(cfg config, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx = logger.WithLog(ctx, log)

	db, err := database.New(ctx)
	if err != nil {
		return fmt.Errorf("could not initialize database: %w", err)
	}

	_, err = otelsql.RegisterDBStatsMetrics(db)
	if err != nil {
		return fmt.Errorf("could not register database metrics: %w", err)
	}

	err = database.Migrate(ctx, db)
	if err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	tlsConfig, development, err := loadTLSConfig(cfg.tlsCert, cfg.tlsKey)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", webhandler.New(WebFS))
	if development != nil {
		log.WarnContext(ctx, "using a generated development certificate; set TLS_CERT and TLS_KEY in production")
		mux.Handle("GET "+certificateHashesPath, development)
	} else {
		// Without this, the web handler would answer with index.html.
		mux.Handle(certificateHashesPath, http.NotFoundHandler())
	}

	rpcMetrics, err := oteltrevrpc.NewMetrics()
	if err != nil {
		return fmt.Errorf("could not create telemetry metrics: %w", err)
	}

	sessionManager := auth.NewManager(cfg.jwtSecret)
	rateLimiter := ratelimitinterceptor.New(map[string]ratelimitinterceptor.Policy{
		authv1.AuthService_Login_FullMethodName: {
			Requests: loginRateLimitRequests,
			Window:   loginRateLimitWindow,
		},
		authv1.AuthService_Signup_FullMethodName: {
			Requests: signupRateLimitRequests,
			Window:   signupRateLimitWindow,
		},
	})

	server := trevrpc.NewServer()
	options := server.Options()
	// Browsers connect over WebTransport, or QMux over WebSocket when UDP is
	// blocked, at /trevrpc, from pages served by mux, at allowedOrigins, or
	// at devOrigins. Every other request is served by mux.
	options.EnableWebTransport = true
	options.WebTransportAdmission = trevrpc.SameOriginAdmission(slices.Concat(cfg.allowedOrigins, devOrigins)...)
	options.HTTPFallback = mux
	server.SetOptions(options)
	server.SetAuthorizer(auth.NewAuthorizer(sessionManager))
	// Remote trace context is not trusted since the API is public. Spans from
	// clients are linked to the server's span instead.
	server.SetInterceptors(oteltrevrpc.NewInterceptor(), rateLimiter, loginterceptor.New(log))
	server.SetMetrics(rpcMetrics)
	server.SetDiagnostics(trevrpc.LogDiagnostics(log))
	authv1handler.Register(server, db, sessionManager)
	numberv1handler.Register(server, db)

	// Native QUIC, HTTP/3 and WebTransport share a UDP socket, and their TCP
	// fallbacks and HTTP/1.1 and HTTP/2 share a TCP socket, on the same port.
	listener, err := trevrpc.Listen(":"+cfg.port, server, trevrpc.ListenOptions{
		TLSConfig:      tlsConfig,
		TrustedProxies: cfg.trustedProxyCIDRs,
	})
	if err != nil {
		return fmt.Errorf("could not listen: %w", err)
	}

	stopLog := context.AfterFunc(ctx, func() {
		log.InfoContext(context.WithoutCancel(ctx), "shutting down")
	})
	defer stopLog()

	log.InfoContext(ctx, "starting", "port", cfg.port, "version", version)
	if err := listener.Serve(ctx); err != nil {
		return fmt.Errorf("could not serve: %w", err)
	}
	return nil
}
