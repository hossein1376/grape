package slogger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func logLine(t *testing.T, log func(*slog.Logger)) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	log(New(WithDestination(&buf)))
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal %q: %v", buf.String(), err)
	}
	return m
}

// Sibling contexts must not overwrite each other's attributes through a
// shared backing array.
func TestWithAttrs_SiblingsDontShareAttrs(t *testing.T) {
	ctx := context.Background()
	for _, k := range []string{"a", "b", "c"} {
		ctx = WithAttrs(ctx, slog.String(k, k))
	}
	x := WithAttrs(ctx, slog.String("child", "x"))
	_ = WithAttrs(ctx, slog.String("child", "y"))

	m := logLine(t, func(l *slog.Logger) { l.InfoContext(x, "msg") })
	if m["child"] != "x" || m["a"] != "a" || m["c"] != "c" {
		t.Fatalf("unexpected attributes: %v", m)
	}
}

func TestContextHandler_DerivedLoggersKeepContextAttrs(t *testing.T) {
	ctx := WithAttrs(context.Background(), slog.String("request_id", "abc"))

	m := logLine(t, func(l *slog.Logger) { l.With("k", "v").InfoContext(ctx, "msg") })
	if m["request_id"] != "abc" || m["k"] != "v" {
		t.Fatalf("With: unexpected attributes: %v", m)
	}

	m = logLine(t, func(l *slog.Logger) { l.WithGroup("g").InfoContext(ctx, "msg") })
	g, _ := m["g"].(map[string]any)
	if g["request_id"] != "abc" {
		t.Fatalf("WithGroup: unexpected attributes: %v", m)
	}
}
