package grape

import (
	"log/slog"
)

type logger struct {
	slog *slog.Logger
}

// Debug logs at LevelDebug.
func (log logger) Debug(msg string, args ...any) {
	log.slog.Debug(msg, args...)
}

// Info logs at LevelInfo.
func (log logger) Info(msg string, args ...any) {
	log.slog.Info(msg, args...)
}

// Warn logs at LevelWarn.
func (log logger) Warn(msg string, args ...any) {
	log.slog.Warn(msg, args...)
}

// Error logs at LevelError.
func (log logger) Error(msg string, args ...any) {
	log.slog.Error(msg, args...)
}
