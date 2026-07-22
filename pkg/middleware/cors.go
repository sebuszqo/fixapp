package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := getAllowedOrigin(origin)

		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

			reqHeaders := r.Header.Get("Access-Control-Request-Headers")
			if reqHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-Request-ID")
			}
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		}

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getAllowedOrigin checks if the given origin is allowed.
func getAllowedOrigin(origin string) string {
	if origin == "" {
		return ""
	}

	// Check CORS_ORIGIN env var first
	if corsOrigin := os.Getenv("CORS_ORIGIN"); corsOrigin != "" {
		corsOrigin = strings.ReplaceAll(corsOrigin, "\r", "")
		corsOrigin = strings.ReplaceAll(corsOrigin, "\n", "")
		origins := strings.Split(corsOrigin, ",")
		for _, o := range origins {
			o = strings.TrimSpace(o)
			if o == "*" || origin == o {
				return origin
			}
		}
	}

	// Development: allow any localhost origins or 127.0.0.1 origins
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") || strings.HasPrefix(origin, "https://localhost:") {
		return origin
	}

	return ""
}
