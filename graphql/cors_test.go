package graphql

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHandlesPreflightForAllowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:8089", "https://raywall.github.io"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/graphql", nil)
	req.Header.Set("Origin", "http://localhost:8089")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8089" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestCORSHandlesGitHubPagesOrigin(t *testing.T) {
	handler := CORSFromEnv()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/graphql", nil)
	req.Header.Set("Origin", "https://raywall.github.io")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://raywall.github.io" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestCORSDoesNotAllowUnknownOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:8089"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}
