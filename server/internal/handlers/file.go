package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/spotdemo4/trevstack/server/internal/interceptors"
	"github.com/spotdemo4/trevstack/server/internal/models"
	"gorm.io/gorm"
)

type FileHandler struct {
	db  *gorm.DB
	key []byte
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userid, ok := interceptors.GetUserContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	// Get the file id from the path
	pathItems := strings.Split(r.URL.Path, "/")
	if len(pathItems) < 3 {
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}
	id := pathItems[2]

	// Get the file from the database
	file := models.File{}
	if err := h.db.First(&file, "id = ? AND user_id = ?", id, userid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Serve the file
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", http.DetectContentType(file.Data))
		w.Write(file.Data)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func NewFileHandler(db *gorm.DB, key string) http.Handler {
	return interceptors.WithAuthRedirect(
		&FileHandler{
			db:  db,
			key: []byte(key),
		},
		key,
	)
}
