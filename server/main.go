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
	"github.com/stephenafamo/bob"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/spotdemo4/trevstack/server/internal/database"
	"github.com/spotdemo4/trevstack/server/internal/handlers/client"
	"github.com/spotdemo4/trevstack/server/internal/handlers/file"
	"github.com/spotdemo4/trevstack/server/internal/handlers/item/v1"
	"github.com/spotdemo4/trevstack/server/internal/handlers/user/v1"
	"github.com/spotdemo4/trevstack/server/internal/interceptors"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Get env
	env, err := getEnv()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get database
	db := &bob.DB{}
	switch env.DBType {
	case "postgres":
		log.Println("Using Postgres")

		if env.DBUser == "" {
			log.Fatal("DB_USER is required")
		}
		if env.DBPass == "" {
			log.Fatal("DB_PASS is required")
		}
		if env.DBHost == "" {
			log.Fatal("DB_HOST is required")
		}
		if env.DBPort == "" {
			log.Fatal("DB_PORT is required")
		}
		if env.DBName == "" {
			log.Fatal("DB_NAME is required")
		}

		db, err = database.NewPostgresConnection(env.DBUser, env.DBPass, env.DBHost, env.DBPort, env.DBName)
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}

	case "sqlite":
		log.Println("Using SQLite")

		if env.DBName == "" {
			log.Fatal("DB_NAME is required")
		}

		db, err = database.NewSQLiteConnection(env.DBName)
		if err != nil {
			log.Fatalf("failed to connect to sqlite: %v", err)
		}

	default:
		log.Fatal("DB_TYPE must be either postgres or sqlite")
	}

	// Serve GRPC Handlers
	api := http.NewServeMux()
	api.Handle(interceptors.WithCORS(user.NewAuthHandler(db, env.Key)))
	api.Handle(interceptors.WithCORS(user.NewHandler(db, env.Key)))
	api.Handle(interceptors.WithCORS(item.NewHandler(db, env.Key)))

	// Serve web interface
	mux := http.NewServeMux()
	mux.Handle("/", client.NewClientHandler(env.Key))
	mux.Handle("/file/", file.NewFileHandler(db, env.Key))
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
	DBType string
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string
	Port   string
	Key    string
}

func getEnv() (*env, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Failed to load .env file, using environment variables")
	}

	// Create
	env := env{
		DBType: os.Getenv("DB_TYPE"),
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
		Port:   os.Getenv("PORT"),
		Key:    os.Getenv("KEY"),
	}

	// Validate
	if env.Port == "" {
		env.Port = "8080"
	}
	if env.Key == "" {
		return nil, errors.New("env 'key' not found")
	}

	return &env, nil
}
