package slogger

import (
	"context"
	"log/slog"
	"os"
)

func New(opts ...Option) *slog.Logger {
	opt := &loggerOptions{
		level: slog.LevelInfo,
		dst:   os.Stdout,
	}
	for _, o := range opts {
		o(opt)
	}

	handlerOpts := &slog.HandlerOptions{
		Level:       opt.level,
		AddSource:   opt.addSrc,
		ReplaceAttr: opt.replaceAtr,
	}
	var handler slog.Handler
	if !opt.textLogger {
		handler = slog.NewJSONHandler(opt.dst, handlerOpts)
	} else {
		handler = slog.NewTextHandler(opt.dst, handlerOpts)
	}

	return slog.New(&ContextHandler{handler})
}

func NewDefault(opts ...Option) *slog.Logger {
	l := New(opts...)
	slog.SetDefault(l)
	return l
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
