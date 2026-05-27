package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/raywall/go-graphql-connector/internal/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, mw := range middlewares {
		h = mw(h)
	}
	return h
}

func ElapsedTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &elapsedTimeWriter{
			ResponseWriter: w,
			start:          time.Now(),
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(writer, r)
		if !writer.wroteHeader {
			writer.WriteHeader(writer.statusCode)
		}
	})
}

type elapsedTimeWriter struct {
	http.ResponseWriter
	start       time.Time
	statusCode  int
	wroteHeader bool
}

func (w *elapsedTimeWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.statusCode = statusCode
	w.Header().Set("x-graphql-elapsed-time", fmt.Sprintf("%d", time.Since(w.start).Milliseconds()))
	w.ResponseWriter.WriteHeader(statusCode)
	w.wroteHeader = true
}

func (w *elapsedTimeWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(w.statusCode)
	}
	return w.ResponseWriter.Write(data)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace, ok := tracing.FromHeaders(r.Header)
		if !ok {
			_, trace = tracing.Ensure(r.Context())
		}
		ctx := tracing.WithContext(r.Context(), trace)
		tracer := otel.Tracer("go-graphql-connector/http")
		ctx, span := tracer.Start(ctx, "graphql.request")
		span.SetAttributes(
			attribute.String("graphql.route", r.URL.Path),
			attribute.String("http.method", r.Method),
			attribute.String("graphql.trace_id", trace.TraceID),
			attribute.String("graphql.span_id", trace.SpanID),
		)
		defer span.End()
		w.Header().Set("X-Trace-ID", trace.TraceID)
		w.Header().Set("traceparent", tracing.Traceparent(trace))
		next.ServeHTTP(w, r.WithContext(ctx))
		span.SetStatus(codes.Ok, "")
	})
}
