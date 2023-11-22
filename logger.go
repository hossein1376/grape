package grape

import (
	"log/slog"
)

type logger struct {
	slog *slog.Logger
}

// Debug logs at LevelDebug.
func (l logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info logs at LevelInfo.
func (l logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn logs at LevelWarn.
func (l logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error logs at LevelError.
func (l logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}
