package graphql

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestManagedTokenSendsClientCredentialsHeadersAndCachesToken(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if got := r.Header.Get("x-serial-number"); got != "serial-123" {
			t.Fatalf("x-serial-number = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("Content-Type = %q", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		rawBody := string(body)
		if !strings.HasPrefix(rawBody, "grant_type=client_credentials&client_id=") {
			t.Fatalf("body order = %q", rawBody)
		}
		values, err := url.ParseQuery(rawBody)
		if err != nil {
			t.Fatalf("ParseQuery returned error: %v", err)
		}
		if values.Get("client_id") != "client-1" || values.Get("client_secret") != "secret-1" {
			t.Fatalf("credentials = %v", values)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "token-abc",
			"expires_in":   300,
		})
	}))
	defer server.Close()

	manager := newManagedToken(server.URL, "client-1", "secret-1", map[string]string{
		"x-serial-number": "serial-123",
	})

	token, err := manager.GetToken()
	if err != nil {
		t.Fatalf("GetToken returned error: %v", err)
	}
	if token != "token-abc" {
		t.Fatalf("token = %q", token)
	}
	token, err = manager.GetToken()
	if err != nil {
		t.Fatalf("second GetToken returned error: %v", err)
	}
	if token != "token-abc" {
		t.Fatalf("second token = %q", token)
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
}
