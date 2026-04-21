// main.go
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
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
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	li := interceptors.NewLogInterceptor(logger)
	vi := validate.NewInterceptor()
	api := http.NewServeMux()

	api.Handle(greetv1handler.New(connect.WithInterceptors(li, vi)))

	mux := http.NewServeMux()
	mux.Handle("/", webHandler())
	mux.Handle("/grpc/", http.StripPrefix("/grpc", api))

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true) // Use h2c so we can serve HTTP/2 without TLS.

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", 8080),
		Handler:   interceptors.WithCORS(mux),
		Protocols: p,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 10 * time.Second,
	}

	wg := sync.WaitGroup{}
	wg.Go((func() {
		log.Println("Starting server on port 8080")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}))

	<-ctx.Done()
	log.Println("Shutting down server")
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
