package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"
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
		// Add tracing logic here (e.g., extract/propagate trace context)
		// Example: traceID := r.Header.Get("X-Trace-ID")
		next.ServeHTTP(w, r)
	})
}
