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

	"connectrpc.com/connect"
	"connectrpc.com/validate"
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
)

const (
	loginRateLimitRequests  = 5
	loginRateLimitWindow    = time.Minute
	signupRateLimitRequests = 3
	signupRateLimitWindow   = time.Hour
)

var (
	DocsFS fs.FS
	WebFS  fs.FS
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

	err = run(cfg, log)
	if err != nil {
		log.Error("exiting", "error", err)
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

	err = database.Migrate(ctx, db)
	if err != nil {
		return fmt.Errorf("could not migrate database: %w", err)
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
	vi := validate.NewInterceptor()

	api := http.NewServeMux()
	api.Handle(authv1handler.New(sessionManager, connect.WithInterceptors(rli, li, ai, vi)))
	api.Handle(numberv1handler.New(connect.WithInterceptors(li, ai, vi)))

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
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("could not listen: %w", err)
	}

	var serveErr error
	wg := sync.WaitGroup{}
	wg.Go((func() {
		log.InfoContext(ctx, "starting", "port", cfg.port)
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
