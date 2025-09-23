package slogger

import (
	"io"
	"log/slog"
)

type loggerOptions struct {
	textLogger bool
	level      slog.Level
	dst        io.Writer
	addSrc     bool
	replaceAtr func(groups []string, a slog.Attr) slog.Attr
}

type Option func(*loggerOptions)

func WithDestination(dst io.Writer) Option {
	return func(o *loggerOptions) {
		o.dst = dst
	}
}

func WithAddSource() Option {
	return func(o *loggerOptions) {
		o.addSrc = true
	}
}

func WithLevel(level slog.Level) Option {
	return func(o *loggerOptions) {
		o.level = level
	}
}

func WithReplaceAtr(f func(groups []string, a slog.Attr) slog.Attr) Option {
	return func(o *loggerOptions) {
		o.replaceAtr = f
	}
}

func WithTextLogger() Option {
	return func(o *loggerOptions) {
		o.textLogger = true
	}
}

func WithJSONLogger() Option {
	return func(o *loggerOptions) {
		o.textLogger = false
	}
}
