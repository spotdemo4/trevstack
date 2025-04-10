package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"runtime"

	_ "github.com/lib/pq" // Postgres
	"github.com/stephenafamo/bob"
)

func NewPostgresConnection(url *url.URL) (*bob.DB, error) {
	db, err := sql.Open("postgres", postgresConnectionString(url))
	if err != nil {
		return nil, err
	}

	bobdb := bob.NewDB(db)

	return &bobdb, nil
}

func postgresConnectionString(u *url.URL) string {
	hostname := u.Hostname()
	port := u.Port()
	query := u.Query()

	// support socket parameter for consistency with mysql
	if query.Get("socket") != "" {
		query.Set("host", query.Get("socket"))
		query.Del("socket")
	}

	// default hostname
	if hostname == "" && query.Get("host") == "" {
		switch runtime.GOOS {
		case "linux":
			query.Set("host", "/var/run/postgresql")
		case "darwin", "freebsd", "dragonfly", "openbsd", "netbsd":
			query.Set("host", "/tmp")
		default:
			hostname = "localhost"
		}
	}

	// host param overrides url hostname
	if query.Get("host") != "" {
		hostname = ""
	}

	// always specify a port
	if query.Get("port") != "" {
		port = query.Get("port")
		query.Del("port")
	}
	if port == "" {
		switch u.Scheme {
		case "redshift":
			port = "5439"
		default:
			port = "5432"
		}
	}

	// generate output URL
	out, _ := url.Parse(u.String())
	// force scheme back to postgres if there was another postgres-compatible scheme
	out.Scheme = "postgres"
	out.Host = fmt.Sprintf("%s:%s", hostname, port)
	out.RawQuery = query.Encode()

	return out.String()
}
