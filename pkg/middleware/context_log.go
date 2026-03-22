package middleware

import (
	"context"
	"net/http"

	"fixapp/pkg/ctxlog"

	"go.uber.org/zap"
)

// AttachLogger creates middleware that attaches a request-scoped logger to the context.
// The logger includes request ID, method, and path for correlation.
func AttachLogger(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := base.With(
				zap.String("req_id", GetRequestID(r.Context())),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
			ctx := ctxlog.NewContext(r.Context(), l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContextLogger retrieves the logger from the context.
// This is a convenience wrapper around ctxlog.FromContext.
func FromContextLogger(ctx context.Context) *zap.Logger {
	return ctxlog.FromContext(ctx)
}
