package grape

import (
	"io"
	"log/slog"
)

type logger struct {
	slog *slog.Logger
}

func (l logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

func (l logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

func (l logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

func (l logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

func newTextLogger(dst io.Writer, level slog.Level) *slog.Logger {
	textLogger := slog.NewTextHandler(dst, &slog.HandlerOptions{Level: level})
	return slog.New(textLogger)
}
