// main.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/XSAM/otelsql"
	"trev.zip/template/stack/server/auth"
	"trev.zip/template/stack/server/connect/auth/v1/authv1connect"
	"trev.zip/template/stack/server/database"
	authv1handler "trev.zip/template/stack/server/handlers/auth/v1"
	docshandler "trev.zip/template/stack/server/handlers/docs"
	numberv1handler "trev.zip/template/stack/server/handlers/number/v1"
	webhandler "trev.zip/template/stack/server/handlers/web"
	authinterceptor "trev.zip/template/stack/server/interceptors/auth"
	corsinterceptor "trev.zip/template/stack/server/interceptors/cors"
	loginterceptor "trev.zip/template/stack/server/interceptors/log"
	ratelimitinterceptor "trev.zip/template/stack/server/interceptors/ratelimit"
	"trev.zip/template/stack/server/logger"
	"trev.zip/template/stack/server/telemetry"
)

const (
	telemetryShutdownTimeout = 5 * time.Second

	loginRateLimitRequests  = 5
	loginRateLimitWindow    = time.Minute
	signupRateLimitRequests = 3
	signupRateLimitWindow   = time.Hour
)

var (
	DocsFS fs.FS
	WebFS  fs.FS

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
	ctx = database.WithDatabase(ctx, db)

	_, err = otelsql.RegisterDBStatsMetrics(db)
	if err != nil {
		return fmt.Errorf("could not register database metrics: %w", err)
	}

	err = database.Migrate(ctx, db)
	if err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
	}

	// Remote trace context is not trusted by default since the API is public.
	// Spans from clients are linked to the server's root span instead.
	oi, err := otelconnect.NewServerInterceptor()
	if err != nil {
		return fmt.Errorf("could not create telemetry interceptor: %w", err)
	}

	sessionManager := auth.NewManager(cfg.jwtSecret, cfg.authCookieSecure)
	ai := authinterceptor.NewAuthInterceptor(sessionManager)
	li := loginterceptor.NewLogInterceptor(log)
	rli := ratelimitinterceptor.NewRateLimitInterceptor(
		map[string]ratelimitinterceptor.RateLimitPolicy{
			authv1connect.AuthServiceLoginProcedure: {
				Requests: loginRateLimitRequests,
				Window:   loginRateLimitWindow,
			},
			authv1connect.AuthServiceSignupProcedure: {
				Requests: signupRateLimitRequests,
				Window:   signupRateLimitWindow,
			},
		},
		cfg.trustedProxyCIDRs,
	)
	vi := validate.NewServerInterceptor()

	rpc := connect.NewServer(oi, rli.WrapServer, li.WrapServer, ai.WrapServer, vi)
	authv1handler.Register(rpc, sessionManager)
	numberv1handler.Register(rpc)

	api := http.NewServeMux()
	connecthttp.Mount(api, rpc)

	mux := http.NewServeMux()
	mux.Handle("/", webhandler.New(WebFS))
	mux.Handle("/docs/", docshandler.New(DocsFS))
	mux.Handle("/grpc/", http.StripPrefix("/grpc", api))

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true) // Use h2c so we can serve HTTP/2 without TLS.

	server := &http.Server{
		Addr:      fmt.Sprintf(":%s", cfg.port),
		Handler:   corsinterceptor.WithCORS(mux),
		Protocols: p,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Bind synchronously so a bad or taken port fails startup instead of
	// leaving the process running without a listener.
	listener, err := listen(server.Addr, cfg.proxyProtocol, cfg.trustedProxyCIDRs)
	if err != nil {
		return fmt.Errorf("could not listen: %w", err)
	}

	var serveErr error
	wg := sync.WaitGroup{}
	wg.Go((func() {
		log.InfoContext(ctx, "starting", "port", cfg.port, "version", version)
		err := server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			serveErr = fmt.Errorf("could not serve: %w", err)
			stop()
		}
	}))

	<-ctx.Done()
	log.InfoContext(ctx, "shutting down")
	server.Shutdown(context.Background())

	wg.Wait()
	return serveErr
}
