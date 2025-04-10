// TrevStack HTTP Server
package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
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

var clientFS *embed.FS
var dbFS *embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Get env
	env, err := getEnv()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Migrate database
	err = database.Migrate(env.DatabaseUrl, dbFS)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get database
	db := &bob.DB{}
	switch env.DatabaseType {
	case "postgres":
		log.Println("Using Postgres")

		db, err = database.NewPostgresConnection(env.DatabaseUrl)
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}

	case "sqlite", "sqlite3":
		log.Println("Using SQLite")

		db, err = database.NewSQLiteConnection(env.DatabaseUrl)
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
	mux.Handle("/", client.NewClientHandler(env.Key, clientFS))
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
	Port         string
	Key          string
	DatabaseType string
	DatabaseUrl  *url.URL
}

func getEnv() (*env, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Failed to load .env file, using environment variables")
	}

	// Create
	env := env{
		Port: os.Getenv("PORT"),
		Key:  os.Getenv("KEY"),
	}

	// Validate
	if env.Port == "" {
		env.Port = "8080"
	}
	if env.Key == "" {
		return nil, errors.New("env 'key' not found")
	}

	// Validate DATABASE_URL
	dbstr := os.Getenv("DATABASE_URL")
	if dbstr == "" {
		return nil, errors.New("env 'DATABASE_URL' not found")
	}

	dbsp := strings.Split(dbstr, ":")
	dburl, err := url.Parse(dbstr)
	if err != nil || len(dbsp) < 2 {
		return nil, errors.New("env 'DATABASE_URL' formatted incorrectly")
	}
	env.DatabaseType = dbsp[0]
	env.DatabaseUrl = dburl

	return &env, nil
}
