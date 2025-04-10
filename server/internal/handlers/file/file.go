package file

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/spotdemo4/trevstack/server/internal/interceptors"
	"github.com/spotdemo4/trevstack/server/internal/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
)

type FileHandler struct {
	db  *bob.DB
	key []byte
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userid, ok := interceptors.GetUserContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	// Make sure this is a GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the file id from the path
	pathItems := strings.Split(r.URL.Path, "/")
	if len(pathItems) < 3 {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}
	id, err := strconv.Atoi(pathItems[2])
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get the file from the database
	file, err := models.Files.Query(
		sqlite.WhereAnd(
			models.SelectWhere.Files.ID.EQ(int64(id)),
			models.SelectWhere.Files.UserID.EQ(userid),
		),
	).One(context.Background(), h.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", http.DetectContentType(file.Data))
	w.Write(file.Data)
}

func NewFileHandler(db *bob.DB, key string) http.Handler {
	return interceptors.WithAuthRedirect(
		&FileHandler{
			db:  db,
			key: []byte(key),
		},
		key,
	)
}
