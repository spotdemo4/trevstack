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
	"github.com/spotdemo4/trevstack/server/database"
	numberv1handler "github.com/spotdemo4/trevstack/server/handlers/number/v1"
	"github.com/spotdemo4/trevstack/server/handlers/web"
	"github.com/spotdemo4/trevstack/server/interceptors"
	"github.com/spotdemo4/trevstack/server/logger"
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

	li := interceptors.NewLogInterceptor(log)
	vi := validate.NewInterceptor()

	web := web.Handler(ctx, WebFS)

	api := http.NewServeMux()
	api.Handle(numberv1handler.New(connect.WithInterceptors(li, vi)))

	mux := http.NewServeMux()
	mux.Handle("/", web)
	mux.Handle("/numbers/", http.StripPrefix("/numbers", web))
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
