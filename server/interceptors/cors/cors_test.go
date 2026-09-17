package cors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSExposesRetryAfter(t *testing.T) {
	handler := WithCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	exposed := strings.ToLower(response.Header().Get("Access-Control-Expose-Headers"))
	if !strings.Contains(exposed, "retry-after") {
		t.Errorf("Access-Control-Expose-Headers = %q, want Retry-After", exposed)
	}
}
