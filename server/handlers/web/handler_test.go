package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandler(t *testing.T) {
	web := fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<h1>Web</h1>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('web')")},
		"sw.js":         &fstest.MapFile{Data: []byte("self.skipWaiting()")},
	}

	tests := []struct {
		name        string
		path        string
		status      int
		body        string
		contentType string
		cache       string
	}{
		{
			name:        "serves index",
			path:        "/",
			status:      http.StatusOK,
			body:        "<h1>Web</h1>",
			contentType: "text/html",
			cache:       "no-cache",
		},
		{
			name:        "serves asset",
			path:        "/assets/app.js",
			status:      http.StatusOK,
			body:        "console.log('web')",
			contentType: "text/javascript",
			cache:       "public, max-age=31536000, immutable",
		},
		{
			name:        "revalidates service worker",
			path:        "/sw.js",
			status:      http.StatusOK,
			body:        "self.skipWaiting()",
			contentType: "text/javascript",
			cache:       "no-cache",
		},
		{
			name:        "does not answer missing asset with index",
			path:        "/assets/old-build.js",
			status:      http.StatusNotFound,
			body:        "404 page not found\n",
			contentType: "text/plain",
		},
		{
			name:        "falls back to index",
			path:        "/settings/profile",
			status:      http.StatusOK,
			body:        "<h1>Web</h1>",
			contentType: "text/html",
			cache:       "no-cache",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			New(web).ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if response.Body.String() != test.body {
				t.Errorf("body = %q, want %q", response.Body.String(), test.body)
			}
			if !strings.HasPrefix(response.Header().Get("Content-Type"), test.contentType) {
				t.Errorf("Content-Type = %q, want prefix %q", response.Header().Get("Content-Type"), test.contentType)
			}
			if response.Header().Get("Cache-Control") != test.cache {
				t.Errorf("Cache-Control = %q, want %q", response.Header().Get("Cache-Control"), test.cache)
			}
		})
	}
}

func TestHandlerWithoutFilesystem(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	New(nil).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
