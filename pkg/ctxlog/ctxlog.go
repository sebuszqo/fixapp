// Package ctxlog provides context-aware logging utilities.
// This package has no internal dependencies, so it can be used anywhere.
package ctxlog

import (
	"context"

	"go.uber.org/zap"
)

// ctxKeyLogger is a private type for the logger context key.
type ctxKeyLogger struct{}

// loggerKey is the context key for storing the logger.
var loggerKey ctxKeyLogger

// NewContext returns a new context with the logger attached.
func NewContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from the context.
// Returns a no-op logger if none is present.
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return zap.NewNop()
}

// With returns a new context with additional fields added to the logger.
func With(ctx context.Context, fields ...zap.Field) context.Context {
	logger := FromContext(ctx).With(fields...)
	return NewContext(ctx, logger)
}


