package docs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandler(t *testing.T) {
	docs := fstest.MapFS{
		"index.html":           &fstest.MapFile{Data: []byte("<h1>API Reference</h1>")},
		"assets/app.js":        &fstest.MapFile{Data: []byte("console.log('docs')")},
		"assets/openapi.yaml": &fstest.MapFile{Data: []byte("openapi: 3.1.0")},
	}

	mux := http.NewServeMux()
	mux.Handle("/docs/", New(docs))

	tests := []struct {
		name        string
		path        string
		status      int
		body        string
		contentType string
		location    string
	}{
		{
			name:     "redirects to trailing slash",
			path:     "/docs",
			status:   http.StatusTemporaryRedirect,
			location: "/docs/",
		},
		{
			name:        "serves index",
			path:        "/docs/",
			status:      http.StatusOK,
			body:        "<h1>API Reference</h1>",
			contentType: "text/html",
		},
		{
			name:        "serves JavaScript asset",
			path:        "/docs/assets/app.js",
			status:      http.StatusOK,
			body:        "console.log('docs')",
			contentType: "text/javascript",
		},
		{
			name:        "serves OpenAPI asset",
			path:        "/docs/assets/openapi.yaml",
			status:      http.StatusOK,
			body:        "openapi: 3.1.0",
			contentType: "text/plain",
		},
		{
			name:   "rejects directory",
			path:   "/docs/assets/",
			status: http.StatusNotFound,
		},
		{
			name:   "rejects missing file",
			path:   "/docs/missing",
			status: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			mux.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if test.body != "" && response.Body.String() != test.body {
				t.Errorf("body = %q, want %q", response.Body.String(), test.body)
			}
			if test.contentType != "" && !strings.HasPrefix(response.Header().Get("Content-Type"), test.contentType) {
				t.Errorf("Content-Type = %q, want prefix %q", response.Header().Get("Content-Type"), test.contentType)
			}
			if test.location != "" && response.Header().Get("Location") != test.location {
				t.Errorf("Location = %q, want %q", response.Header().Get("Location"), test.location)
			}
		})
	}
}

func TestHandlerRejectsInvalidPath(t *testing.T) {
	handler := New(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("index")},
	})
	request := httptest.NewRequest(http.MethodGet, "/docs/secret", nil)
	request.URL.Path = "/docs/../secret"
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestHandlerWithoutFilesystem(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	response := httptest.NewRecorder()

	New(nil).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
