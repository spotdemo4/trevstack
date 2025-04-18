// TrevStack HTTP Server
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	embed "github.com/spotdemo4/trevstack/server"
	"github.com/spotdemo4/trevstack/server/internal/database"
	"github.com/spotdemo4/trevstack/server/internal/handlers/client"
	"github.com/spotdemo4/trevstack/server/internal/handlers/file"
	"github.com/spotdemo4/trevstack/server/internal/handlers/item/v1"
	"github.com/spotdemo4/trevstack/server/internal/handlers/user/v1"
	"github.com/spotdemo4/trevstack/server/internal/interceptors"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	// Get env
	env, err := getEnv()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Migrate database
	err = database.Migrate(env.DatabaseURL, embed.DBFS)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get database
	sqlc, db, err := database.New(env.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err.Error())
	}

	// Serve GRPC Handlers
	api := http.NewServeMux()
	api.Handle(interceptors.WithCORS(user.NewAuthHandler(sqlc, env.Key)))
	api.Handle(interceptors.WithCORS(user.NewHandler(sqlc, env.Key)))
	api.Handle(interceptors.WithCORS(item.NewHandler(sqlc, env.Key)))

	// Serve web interface
	mux := http.NewServeMux()
	mux.Handle("/", client.NewClientHandler(env.Key, embed.ClientFS))
	mux.Handle("/file/", file.NewFileHandler(sqlc, env.Key))
	mux.Handle("/grpc/", http.StripPrefix("/grpc", api))

	// Start server
	log.Printf("Starting server on :%s", env.Port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", env.Port),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Gracefully shutdown on SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		slog.Warn(fmt.Sprintf("Received signal %s, exiting", sig))

		// Close HTTP server
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
		}
		cancel()

		// Close database connection
		db.Close()
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type env struct {
	Port        string
	Key         string
	DatabaseURL string
}

func getEnv() (*env, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Failed to load .env file, using environment variables")
	}

	// Create
	env := env{
		Port:        os.Getenv("PORT"),
		Key:         os.Getenv("KEY"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	// Validate
	if env.Port == "" {
		env.Port = "8080"
	}
	if env.Key == "" {
		return nil, errors.New("env 'key' not found")
	}
	if env.DatabaseURL == "" {
		return nil, errors.New("env 'DATABASE_URL' not found")
	}

	return &env, nil
}
