package middleware

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type ctxKeyLogger struct{}

var loggerKey ctxKeyLogger

func AttachLogger(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := base.With(
				zap.String("req_id", GetRequestID(r.Context())),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
			ctx := context.WithValue(r.Context(), loggerKey, l)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContextLogger(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.NewNop()
}
