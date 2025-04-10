package database

import (
	"database/sql"
	"net/url"
	"regexp"

	"github.com/stephenafamo/bob"
	_ "modernc.org/sqlite" // Sqlite
)

func NewSQLiteConnection(url *url.URL) (*bob.DB, error) {
	db, err := sql.Open("sqlite", sqliteConnectionString(url))
	if err != nil {
		return nil, err
	}

	// Create new bob db
	bobdb := bob.NewDB(db)

	return &bobdb, nil
}

// ConnectionString converts a URL into a valid connection string
func sqliteConnectionString(u *url.URL) string {
	// duplicate URL and remove scheme
	newURL := *u
	newURL.Scheme = ""

	if newURL.Opaque == "" && newURL.Path != "" {
		// When the DSN is in the form "scheme:/absolute/path" or
		// "scheme://absolute/path" or "scheme:///absolute/path", url.Parse
		// will consider the file path as :
		// - "absolute" as the hostname
		// - "path" (and the rest until "?") as the URL path.
		// Instead, when the DSN is in the form "scheme:", the (relative) file
		// path is stored in the "Opaque" field.
		// See: https://pkg.go.dev/net/url#URL
		//
		// While Opaque is not escaped, the URL Path is. So, if .Path contains
		// the file path, we need to un-escape it, and rebuild the full path.

		newURL.Opaque = "//" + newURL.Host + mustUnescapePath(newURL.Path)
		newURL.Path = ""
	}

	// trim duplicate leading slashes
	str := regexp.MustCompile("^//+").ReplaceAllString(newURL.String(), "/")

	return str
}

// MustUnescapePath unescapes a URL path, and panics if it fails.
// It is used during in cases where we are parsing a generated path.
func mustUnescapePath(s string) string {
	if s == "" {
		panic("missing path")
	}

	path, err := url.PathUnescape(s)
	if err != nil {
		panic(err)
	}

	return path
}
