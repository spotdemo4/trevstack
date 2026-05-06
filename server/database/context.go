package database

import (
	"context"
	"database/sql"
)

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
