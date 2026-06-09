package graphql

import (
	"net/http"
	"os"
	"strings"
)

func CORSFromEnv() Middleware {
	allowed := splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(allowed) == 0 {
		allowed = []string{
			"http://localhost:8089",
			"http://127.0.0.1:8089",
			"https://raywall.github.io",
		}
	}
	return CORS(allowed)
}

func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Serial-Number, x-api-filter, traceparent, X-Trace-ID, X-Span-ID, X-Parent-Span-ID")
				w.Header().Set("Access-Control-Expose-Headers", "x-graphql-elapsed-time, traceparent, X-Trace-ID")
				w.Header().Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
