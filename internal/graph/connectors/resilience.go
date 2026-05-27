package connectors

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	ErrorClassRetryable    = "retryable"
	ErrorClassNonRetryable = "non_retryable"
	ErrorClassTimeout      = "timeout"
	ErrorClassCircuitOpen  = "circuit_open"
	ErrorClassAuth         = "auth_error"
)

// ResilienceConfig controls connector retry, backoff and circuit breaker.
type ResilienceConfig struct {
	Backoff           string               `json:"backoff,omitempty"`
	InitialIntervalMS int                  `json:"initial_interval_ms,omitempty"`
	MaxIntervalMS     int                  `json:"max_interval_ms,omitempty"`
	Jitter            bool                 `json:"jitter,omitempty"`
	CircuitBreaker    CircuitBreakerConfig `json:"circuit_breaker,omitempty"`
}

// CircuitBreakerConfig controls when a connector stops calling a failing target.
type CircuitBreakerConfig struct {
	Enabled          bool `json:"enabled,omitempty"`
	FailureThreshold int  `json:"failure_threshold,omitempty"`
	OpenTimeoutMS    int  `json:"open_timeout_ms,omitempty"`
}

// CircuitState is a diagnostic snapshot of a connector circuit breaker.
type CircuitState struct {
	Enabled     bool      `json:"enabled"`
	State       string    `json:"state"`
	Failures    int       `json:"failures"`
	OpenedUntil time.Time `json:"opened_until,omitempty"`
}

type CircuitOpenError struct {
	Field       string
	OpenedUntil time.Time
}

func (e CircuitOpenError) Error() string {
	return fmt.Sprintf("connector %q circuit is open until %s", e.Field, e.OpenedUntil.Format(time.RFC3339))
}

type circuitBreaker struct {
	mu               sync.Mutex
	enabled          bool
	failureThreshold int
	openTimeout      time.Duration
	failures         int
	openedUntil      time.Time
}

func newCircuitBreaker(config CircuitBreakerConfig) *circuitBreaker {
	threshold := config.FailureThreshold
	if threshold <= 0 {
		threshold = 5
	}
	openTimeout := time.Duration(config.OpenTimeoutMS) * time.Millisecond
	if openTimeout <= 0 {
		openTimeout = 30 * time.Second
	}
	return &circuitBreaker{
		enabled:          config.Enabled,
		failureThreshold: threshold,
		openTimeout:      openTimeout,
	}
}

func (c *circuitBreaker) allow(field string, now time.Time) error {
	if c == nil || !c.enabled {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.openedUntil.IsZero() || !now.Before(c.openedUntil) {
		return nil
	}
	return CircuitOpenError{Field: field, OpenedUntil: c.openedUntil}
}

func (c *circuitBreaker) recordSuccess() {
	if c == nil || !c.enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = 0
	c.openedUntil = time.Time{}
}

func (c *circuitBreaker) recordFailure(now time.Time) {
	if c == nil || !c.enabled {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures++
	if c.failures >= c.failureThreshold {
		c.openedUntil = now.Add(c.openTimeout)
	}
}

func (c *circuitBreaker) state(now time.Time) CircuitState {
	if c == nil {
		return CircuitState{State: "disabled"}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state := "closed"
	if !c.enabled {
		state = "disabled"
	} else if !c.openedUntil.IsZero() && now.Before(c.openedUntil) {
		state = "open"
	} else if !c.openedUntil.IsZero() && !now.Before(c.openedUntil) {
		state = "half_open"
	}
	return CircuitState{
		Enabled:     c.enabled,
		State:       state,
		Failures:    c.failures,
		OpenedUntil: c.openedUntil,
	}
}

func classifyConnectorError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorClassTimeout
	}
	var circuitErr CircuitOpenError
	if errors.As(err, &circuitErr) {
		return ErrorClassCircuitOpen
	}
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "context deadline exceeded"), strings.Contains(text, "timeout"):
		return ErrorClassTimeout
	case strings.Contains(text, "status 401"), strings.Contains(text, "status 403"), strings.Contains(text, "authorization"), strings.Contains(text, "sts token"):
		return ErrorClassAuth
	case strings.Contains(text, "missing argument"), strings.Contains(text, "unwrapPath"), strings.Contains(text, "response contains errors"):
		return ErrorClassNonRetryable
	default:
		return ErrorClassRetryable
	}
}

func retryableConnectorError(class string) bool {
	return class == ErrorClassRetryable || class == ErrorClassTimeout
}

func (r ResilienceConfig) backoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	initial := time.Duration(r.InitialIntervalMS) * time.Millisecond
	if initial <= 0 {
		initial = 100 * time.Millisecond
	}
	maxInterval := time.Duration(r.MaxIntervalMS) * time.Millisecond
	if maxInterval <= 0 {
		maxInterval = time.Second
	}
	delay := initial
	switch strings.ToLower(strings.TrimSpace(r.Backoff)) {
	case "fixed":
	case "none":
		delay = 0
	default:
		for i := 1; i < attempt; i++ {
			delay *= 2
			if delay >= maxInterval {
				delay = maxInterval
				break
			}
		}
	}
	if delay > maxInterval {
		delay = maxInterval
	}
	if r.Jitter && delay > 0 {
		delay += time.Duration(rand.Int63n(int64(delay / 2)))
	}
	return delay
}
