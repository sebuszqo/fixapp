package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Global logger instance
var Log *zap.Logger

// Initialize sets up the global logger with default handling
func Initialize(level string) error {
	if level == "" {
		level = "info"
	}

	var err error
	Log, err = New(level)
	return err
}

// New creates a new zap logger instance
func New(level string) (*zap.Logger, error) {
	var cfg zap.Config

	// Choose config based on environment
	if level == "debug" || level == "development" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// Parse log level
	parsedLevel, err := zapcore.ParseLevel(level)
	if err == nil {
		cfg.Level = zap.NewAtomicLevelAt(parsedLevel)
	}

	// Customize encoding
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Build logger
	logger, err := cfg.Build(
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
