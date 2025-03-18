// TrevStack HTTP Server
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	connectcors "connectrpc.com/cors"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"

	"github.com/spotdemo4/trevstack/server/internal/database"
	"github.com/spotdemo4/trevstack/server/internal/handlers"
	"github.com/spotdemo4/trevstack/server/internal/handlers/client"
	"github.com/spotdemo4/trevstack/server/internal/handlers/item/v1"
	"github.com/spotdemo4/trevstack/server/internal/handlers/user/v1"
)

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env file, using environment variables")
	}

	// Get environment variables for server
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
	if env.Port == "" {
		env.Port = "8080"
	}
	if env.Key == "" {
		log.Fatal("KEY is required")
	}

	// Get environment variables for database
	db := &gorm.DB{}
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

	// Init database
	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Serve GRPC Handlers
	api := http.NewServeMux()
	api.Handle(withCORS(user.NewAuthHandler(db, env.Key)))
	api.Handle(withCORS(user.NewHandler(db, env.Key)))
	api.Handle(withCORS(item.NewHandler(db, env.Key)))

	// Serve web interface
	mux := http.NewServeMux()
	mux.Handle("/", client.NewClientHandler(env.Key))
	mux.Handle("/file/", handlers.NewFileHandler(db, env.Key))
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
		log.Printf("Received signal %s", sig)
		log.Println("Exiting")

		// Close HTTP server
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
		}
		cancel()

		// Close database connection
		sqlDB, err := db.DB() // Get underlying SQL database
		if err == nil {
			sqlDB.Close()
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// withCORS adds CORS support to a Connect HTTP handler.
func withCORS(pattern string, h http.Handler) (string, http.Handler) {
	middleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	return pattern, middleware.Handler(h)
}
