package database

import (
	"context"
	"database/sql"
)

type key struct{}

// WithDatabase returns a copy of ctx that carries db.
func WithDatabase(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, key{}, db)
}

// FromContext returns the database stored in ctx by WithDatabase, or opens
// a new one if there is none.
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
