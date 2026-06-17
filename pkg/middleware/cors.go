package middleware

import (
	"net/http"
	"os"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
// In development, it allows requests from localhost:5173 (Vite dev server).
// In production, it uses the CORS_ORIGIN environment variable.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := getAllowedOrigin(origin)

		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-Request-ID")
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
	// Check CORS_ORIGIN env var first (production)
	if corsOrigin := os.Getenv("CORS_ORIGIN"); corsOrigin != "" {
		if origin == corsOrigin {
			return origin
		}
		return ""
	}

	// Development: allow localhost origins
	allowedDevOrigins := map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:3000": true,
		"http://localhost:4173": true, // Vite preview
		"http://127.0.0.1:5173": true,
	}

	if allowedDevOrigins[origin] {
		return origin
	}

	return ""
}
