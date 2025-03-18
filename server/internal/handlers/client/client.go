package client

import (
	"net/http"

	"github.com/spotdemo4/trevstack/server/internal/interceptors"
)

var embedfs *http.FileSystem

func NewClientHandler(key string) http.Handler {
	if embedfs != nil {
		return interceptors.WithAuthRedirect(http.FileServer(*embedfs), key)
	}

	return http.NotFoundHandler()
}
