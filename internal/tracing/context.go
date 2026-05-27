package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type contextKey struct{}

// Context stores the W3C trace identifiers propagated by HTTP and connectors.
type Context struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	TraceFlags   string
}

// FromContext returns the trace data attached to ctx.
func FromContext(ctx context.Context) (Context, bool) {
	trace, ok := ctx.Value(contextKey{}).(Context)
	return trace, ok && trace.TraceID != "" && trace.SpanID != ""
}

// WithContext attaches trace data to ctx.
func WithContext(ctx context.Context, trace Context) context.Context {
	return context.WithValue(ctx, contextKey{}, trace)
}

// FromHeaders extracts W3C trace data from request headers.
func FromHeaders(header http.Header) (Context, bool) {
	if trace, ok := ParseTraceparent(header.Get("traceparent")); ok {
		return trace, true
	}
	traceID := firstNonEmpty(header.Get("X-Trace-ID"), header.Get("trace_id"))
	if traceID == "" {
		return Context{}, false
	}
	spanID := firstNonEmpty(header.Get("X-Span-ID"), header.Get("span_id"))
	if spanID == "" {
		spanID = newHexID(8)
	}
	return Context{TraceID: traceID, SpanID: spanID, TraceFlags: "01"}, true
}

// Ensure returns an existing trace or creates a new root trace.
func Ensure(ctx context.Context) (context.Context, Context) {
	if trace, ok := FromContext(ctx); ok {
		return ctx, trace
	}
	trace := Context{TraceID: newHexID(16), SpanID: newHexID(8), TraceFlags: "01"}
	return WithContext(ctx, trace), trace
}

// Child creates a child span context while keeping the same trace id.
func Child(ctx context.Context) (context.Context, Context) {
	ctx, parent := Ensure(ctx)
	trace := Context{
		TraceID:      parent.TraceID,
		ParentSpanID: parent.SpanID,
		SpanID:       newHexID(8),
		TraceFlags:   parent.TraceFlags,
	}
	return WithContext(ctx, trace), trace
}

// Inject writes trace headers into an outgoing request.
func Inject(req *http.Request) {
	if req == nil {
		return
	}
	trace, ok := FromContext(req.Context())
	if !ok {
		return
	}
	req.Header.Set("traceparent", Traceparent(trace))
	req.Header.Set("X-Trace-ID", trace.TraceID)
	req.Header.Set("X-Span-ID", trace.SpanID)
	if trace.ParentSpanID != "" {
		req.Header.Set("X-Parent-Span-ID", trace.ParentSpanID)
	}
}

// Traceparent serializes trace data as a W3C traceparent header.
func Traceparent(trace Context) string {
	flags := trace.TraceFlags
	if flags == "" {
		flags = "01"
	}
	return fmt.Sprintf("00-%s-%s-%s", trace.TraceID, trace.SpanID, flags)
}

// ParseTraceparent reads a W3C traceparent header.
func ParseTraceparent(value string) (Context, bool) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 || parts[0] != "00" || len(parts[1]) != 32 || len(parts[2]) != 16 {
		return Context{}, false
	}
	return Context{TraceID: parts[1], SpanID: parts[2], TraceFlags: parts[3]}, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func newHexID(bytesLen int) string {
	data := make([]byte, bytesLen)
	if _, err := rand.Read(data); err != nil {
		return strings.Repeat("0", bytesLen*2)
	}
	return hex.EncodeToString(data)
}
