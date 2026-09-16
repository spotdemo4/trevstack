// main.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"trev.zip/llc/stack/server/auth"
	"trev.zip/llc/stack/server/connect/auth/v1/authv1connect"
	"trev.zip/llc/stack/server/database"
	authv1handler "trev.zip/llc/stack/server/handlers/auth/v1"
	docshandler "trev.zip/llc/stack/server/handlers/docs"
	numberv1handler "trev.zip/llc/stack/server/handlers/number/v1"
	webhandler "trev.zip/llc/stack/server/handlers/web"
	"trev.zip/llc/stack/server/interceptors"
	"trev.zip/llc/stack/server/logger"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.New(cfg.logLevel)
	ctx = logger.WithLog(ctx, log)

	db, err := database.New(ctx)
	if err != nil {
		log.ErrorContext(ctx, "could not initialize database", "error", err)
		return
	}
	ctx = database.WithDatabase(ctx, db)

	err = database.Migrate(ctx, db)
	if err != nil {
		log.ErrorContext(ctx, "could not migrate database", "error", err)
		return
	}

	sessionManager := auth.NewManager(cfg.jwtSecret, cfg.authCookieSecure)
	ai := interceptors.NewAuthInterceptor(sessionManager)
	li := interceptors.NewLogInterceptor(log)
	rli := interceptors.NewRateLimitInterceptor(
		map[string]interceptors.RateLimitPolicy{
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
		Handler:   interceptors.WithCORS(mux),
		Protocols: p,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 10 * time.Second,
	}

	wg := sync.WaitGroup{}
	wg.Go((func() {
		log.InfoContext(ctx, "starting", "port", cfg.port)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.ErrorContext(ctx, "could not listen and serve", "error", err)
		}
	}))

	<-ctx.Done()
	log.InfoContext(ctx, "shutting down")
	server.Shutdown(context.Background())

	wg.Wait()
}
