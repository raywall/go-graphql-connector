package graphql

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestNewHandlerAddsElapsedTimeHeader(t *testing.T) {
	api, err := New(&Config{
		Schema: `{
			"types": [
				{
					"name": "CombinedData",
					"fields": [
						{"name": "status", "type": "String"}
					]
				}
			],
			"query": {
				"name": "Query",
				"fields": [
					{"name": "dataSources", "type": "Object", "ofType": "CombinedData"}
				]
			}
		}`,
		Connectors: `{"connectors":[]}`,
	}, testResources(), "us-east-1", "http://localhost:4566")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	rec := httptest.NewRecorder()

	api.NewHandler(true).ServeHTTP(rec, req)

	got := rec.Header().Get("x-graphql-elapsed-time")
	if got == "" {
		t.Fatal("expected x-graphql-elapsed-time header")
	}
	if _, err := strconv.Atoi(got); err != nil {
		t.Fatalf("expected x-graphql-elapsed-time to be numeric, got %q", got)
	}
}
