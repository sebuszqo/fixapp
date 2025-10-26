package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// prywatny typ -> brak kolizji
type ctxKeyReqID struct{}

var requestIDKey ctxKeyReqID

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		// zwróć do klienta
		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}
