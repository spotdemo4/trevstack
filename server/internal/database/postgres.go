package database

import (
	"database/sql"

	_ "github.com/lib/pq" // Postgres
	"github.com/stephenafamo/bob"
)

func NewPostgresConnection(user, pass, host, port, name string) (*bob.DB, error) {
	dsn := "host=" + host + " user=" + user + " password=" + pass + " dbname=" + name + " port=" + port + " sslmode=disable TimeZone=UTC"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	bobdb := bob.NewDB(db)

	return &bobdb, nil
}
