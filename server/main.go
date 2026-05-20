// main.go
package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"trev.zip/trev/stack/server/database"
	docshandler "trev.zip/trev/stack/server/handlers/docs"
	numberv1handler "trev.zip/trev/stack/server/handlers/number/v1"
	webhandler "trev.zip/trev/stack/server/handlers/web"
	"trev.zip/trev/stack/server/interceptors"
	"trev.zip/trev/stack/server/logger"
)

var WebFS embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.New()
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

	li := interceptors.NewLogInterceptor(log)
	vi := validate.NewInterceptor()

	api := http.NewServeMux()
	api.Handle(numberv1handler.New(connect.WithInterceptors(li, vi)))

	mux := http.NewServeMux()
	mux.Handle("/", webhandler.New(ctx, WebFS))
	mux.Handle("/docs/", docshandler.New(ctx))
	mux.Handle("/grpc/", http.StripPrefix("/grpc", api))

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true) // Use h2c so we can serve HTTP/2 without TLS.

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%s", port),
		Handler:   interceptors.WithCORS(mux),
		Protocols: p,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 10 * time.Second,
	}

	wg := sync.WaitGroup{}
	wg.Go((func() {
		log.InfoContext(ctx, "starting", "port", port)
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
