package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracingUsesIncomingTraceparent(t *testing.T) {
	const traceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	handler := Tracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	req.Header.Set("traceparent", traceparent)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-ID"); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("X-Trace-ID = %q", got)
	}
	if got := rec.Header().Get("traceparent"); got != traceparent {
		t.Fatalf("traceparent = %q", got)
	}
}

func TestTracingCreatesTraceWhenMissing(t *testing.T) {
	handler := Tracing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/graphql", nil))

	if got := rec.Header().Get("X-Trace-ID"); len(got) != 32 {
		t.Fatalf("X-Trace-ID length = %d, want 32", len(got))
	}
	if got := rec.Header().Get("traceparent"); got == "" {
		t.Fatal("expected traceparent header")
	}
}
