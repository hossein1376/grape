package slogger

import (
	"context"
	"io"
	"log/slog"
)

// NewJSONLogger creates a new slog.Logger instance which logs to the stdout.
// Level must be one of: "trace", "debug", "info", "warn", "error", or "fatal".
// It will default to Info level if the given level is invalid.
func NewJSONLogger(level slog.Level, w io.Writer) *slog.Logger {
	h := &ContextHandler{slog.NewJSONHandler(
		w, &slog.HandlerOptions{
			Level: level,
		}),
	}
	return slog.New(h)
}

func Err(msg string, err error) slog.Attr {
	if err == nil {
		return slog.String(msg, "no-error")
	}
	return slog.String(msg, err.Error())
}

func Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

func Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

func Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

func Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	slog.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}
