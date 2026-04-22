// main.go
package main

import (
	"context"
	"embed"
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
	greetv1handler "github.com/spotdemo4/trevstack/server/handlers/greet/v1"
	"github.com/spotdemo4/trevstack/server/interceptors"
	"github.com/spotdemo4/trevstack/server/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := logger.New()
	ctx = logger.WithLog(ctx, log)

	li := interceptors.NewLogInterceptor(log)
	vi := validate.NewInterceptor()
	api := http.NewServeMux()

	api.Handle(greetv1handler.New(connect.WithInterceptors(li, vi)))

	mux := http.NewServeMux()
	mux.Handle("/", webHandler())
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

var WebFS embed.FS

func webHandler() http.Handler {
	entries, err := WebFS.ReadDir(".")
	if err != nil || len(entries) == 0 {
		return http.NotFoundHandler()
	}

	web, err := fs.Sub(WebFS, "web")
	if err != nil {
		return http.NotFoundHandler()
	}

	return http.FileServer(http.FS(web))
}
