package slogger

import (
	"context"
	"encoding/hex"
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

func UUID[T ~[16]byte](msg string, uuid T) slog.Attr {
	// Based on implementation in: github.com/google/uuid
	var dst [36]byte
	hex.Encode(dst[:], uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
	return slog.String(msg, string(dst[:]))
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
