package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var options = []string{
	// Use a private cache to avoid issues with multiple connections.
	// Shared cache is an obsolete feature that SQLite discourages using.
	// WAL mode provides better concurrent access without shared cache complexity.
	// See: https://www.sqlite.org/sharedcache.html
	"cache=private",

	// Enable foreign key constraints by default.
	// This is not enabled by default in SQLite for backwards compatibility reasons.
	// See: https://www.sqlite.org/foreignkeys.html
	"_foreign_keys=true",

	// Busy timeout of 10 seconds to avoid immediate "database is locked" errors.
	// The default is 0 (fail immediately), which is too aggressive for most applications.
	// 10 seconds allows complex queries and transactions to complete while preventing
	// indefinite hangs. Can be overridden based on application needs.
	// See: https://www.sqlite.org/pragma.html#pragma_busy_timeout
	"_busy_timeout=10000",

	// WAL mode is almost always better than the default DELETE mode.
	// There are some use-cases where DELETE makes sense (e.g. mostly read-only databases),
	// but for most use-cases WAL is better.
	// See: https://www.sqlite.org/wal.html
	"_journal_mode=WAL",

	// Normal synchronous mode is a good balance between performance and safety.
	// In WAL mode, NORMAL is safe from corruption and equivalent to FULL for
	// application crashes. Only a power loss can cause recently committed transactions
	// to roll back. For most applications using WAL mode, NORMAL is the best choice.
	// See: https://www.sqlite.org/pragma.html#pragma_synchronous
	"_synchronous=NORMAL", // use uppercase for consistency with SQLite docs

	// Set a reasonable cache size to avoid using too much memory.
	// Negative value means size in KiB, positive means number of pages.
	// When the cache is full, SQLite will evict pages using an LRU algorithm.
	// See: https://www.sqlite.org/pragma.html#pragma_cache_size
	"_cache_size=-32768", // -32768 means 32 MiB of cache.
}

//go:embed numbers.sql
var numbersSQL string

func New(ctx context.Context) (*sql.DB, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(userConfigDir, "trevstack")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(configDir, "trevstack.db")
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?%s", dbPath, strings.Join(options, "&")))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, numbersSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type key struct{}

func WithDatabase(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, key{}, db)
}

func FromContext(ctx context.Context) *sql.DB {
	if ctx == nil {
		ndb, err := New(ctx)
		if err != nil {
			return nil
		}

		return ndb
	}

	db, ok := ctx.Value(key{}).(*sql.DB)
	if !ok {
		ndb, err := New(ctx)
		if err != nil {
			return nil
		}

		return ndb
	}

	return db
}
