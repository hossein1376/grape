// Package grape is a modern, zero-dependency HTTP library for Go.
// Visit https://github.com/hossein1376/grape for more information.
package grape

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

type Map = map[string]any

const defaultMaxBodySize = 1_048_576 // 1mb

func Param[T any](
	r *http.Request,
	name string,
	parser func(s string) (T, error),
) (T, error) {
	var t T
	param := r.PathValue(name)
	i, err := parser(param)
	if err != nil {
		return t, fmt.Errorf("parse: %w", err)
	}
	if fmt.Sprintf("%v", i) != param {
		return t, fmt.Errorf("parser: expected %s, got %v", param, i)
	}

	return i, nil
}

func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Go spawns a new goroutine and will recover in case of panic; logging
// the error message in Error level. Using this function ensures that
// panic in goroutines will not stop the application's execution.
func Go(f func()) {
	go func() {
		defer func() {
			if msg := recover(); msg != nil {
				slog.Error("goroutine panic", slog.Any("message", msg))
			}
		}()
		f()
	}()
}
