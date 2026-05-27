package connectors

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeAdapter struct {
	failures int
	calls    int
}

func (f *fakeAdapter) GetData(_ context.Context, _ string) (map[string]interface{}, error) {
	f.calls++
	if f.calls <= f.failures {
		return nil, errors.New("temporary network error")
	}
	return map[string]interface{}{"ok": true}, nil
}

func TestConnectorRetriesRetryableErrors(t *testing.T) {
	adapter := &fakeAdapter{failures: 2}
	conn := &connector{
		field:      "orders",
		adapter:    adapter,
		keyPattern: "/orders/{id}",
		timeout:    time.Second,
		retries:    2,
		resilience: ResilienceConfig{Backoff: "none"},
		circuit:    newCircuitBreaker(CircuitBreakerConfig{}),
	}

	got, err := conn.GetData(context.Background(), map[string]interface{}{"id": "1"})
	if err != nil {
		t.Fatalf("GetData returned error: %v", err)
	}
	if adapter.calls != 3 {
		t.Fatalf("calls = %d, want 3", adapter.calls)
	}
	if gotMap, ok := got.(map[string]interface{}); !ok || gotMap["ok"] != true {
		t.Fatalf("unexpected response: %#v", got)
	}
}

func TestConnectorCircuitBreakerOpensAfterFailures(t *testing.T) {
	adapter := &fakeAdapter{failures: 10}
	conn := &connector{
		field:      "orders",
		adapter:    adapter,
		keyPattern: "/orders/{id}",
		timeout:    time.Second,
		retries:    0,
		resilience: ResilienceConfig{Backoff: "none", CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 1,
			OpenTimeoutMS:    1000,
		}},
		circuit: newCircuitBreaker(CircuitBreakerConfig{
			Enabled:          true,
			FailureThreshold: 1,
			OpenTimeoutMS:    1000,
		}),
	}

	if _, err := conn.GetData(context.Background(), map[string]interface{}{"id": "1"}); err == nil {
		t.Fatal("expected first call to fail")
	}
	_, err := conn.GetData(context.Background(), map[string]interface{}{"id": "1"})
	var circuitErr CircuitOpenError
	if !errors.As(err, &circuitErr) {
		t.Fatalf("expected CircuitOpenError, got %v", err)
	}
	if state := conn.CircuitState(); state.State != "open" {
		t.Fatalf("circuit state = %q, want open", state.State)
	}
}

func TestClassifyConnectorError(t *testing.T) {
	tests := map[string]string{
		"REST API returned status 401":       ErrorClassAuth,
		"context deadline exceeded":          ErrorClassTimeout,
		"missing argument(s) for template":   ErrorClassNonRetryable,
		"connector response contains errors": ErrorClassNonRetryable,
		"dial tcp connection refused":        ErrorClassRetryable,
	}
	for input, want := range tests {
		if got := classifyConnectorError(errors.New(input)); got != want {
			t.Fatalf("classifyConnectorError(%q) = %q, want %q", input, got, want)
		}
	}
}
