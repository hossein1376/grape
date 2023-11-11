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

func newJsonLogger(dst io.Writer, level slog.Level) logger {
	jsonHandler := slog.NewJSONHandler(dst, &slog.HandlerOptions{Level: level})
	return logger{slog.New(jsonHandler)}
}

func newTextLogger(dst io.Writer, level slog.Level) logger {
	textLogger := slog.NewTextHandler(dst, &slog.HandlerOptions{Level: level})
	return logger{slog.New(textLogger)}
}
