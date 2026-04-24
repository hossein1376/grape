package slogger

import (
	"context"
	"fmt"
	"log/slog"
)

func Err(msg string, err error) slog.Attr {
	if err == nil {
		return slog.String(msg, "no-error")
	}
	return slog.String(msg, err.Error())
}

func String[T fmt.Stringer](msg string, s T) slog.Attr {
	return slog.String(msg, s.String())
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func Number[T number](msg string, num T) slog.Attr {
	return slog.String(msg, fmt.Sprintf("%v", num))
}

type slogAttr string

const slogAttrs slogAttr = "slog_attrs"

// ContextHandler embeds slog.Handler, overriding Handle method to log context
// attributes.
type ContextHandler struct {
	slog.Handler
}

// Handle adds contextual attributes to the Record before calling the underlying
// handler.
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogAttrs).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs adds one or more slog attributes to the provided context, so that
// they will be included in any Log Records created with such context. It relies
// on the caller to not pass a nil context.
func WithAttrs(parent context.Context, attr ...slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	if len(attr) == 0 {
		return parent
	}

	// if some slog attributes already exist, append to them
	if v, ok := parent.Value(slogAttrs).([]slog.Attr); ok {
		v = append(v, attr...)
		return context.WithValue(parent, slogAttrs, v)
	}

	var v []slog.Attr
	v = append(v, attr...)
	return context.WithValue(parent, slogAttrs, v)
}
