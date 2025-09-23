// Package grape is a modern, zero-dependency HTTP library for Go.
// Visit https://github.com/hossein1376/grape for more information.
package grape

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

type Map = map[string]any

const defaultMaxBodySize = 1_048_576 // 1mb

var ErrMissingParam = errors.New("parameter not found")

// Param attempts to extract the given parameter from path, or URL query params.
// Then, the parser function is used to parse it to the expected type.
// Example of the parser functions for primitive types are: [strconv.Atoi],
// [strconv.ParseBool], [strconv.ParseFloat], and [ParseInt64]. It can also be
// manually implemented for custom types.
func Param[T any](
	r *http.Request,
	name string,
	parser func(s string) (T, error),
) (T, error) {
	var t T
	param := r.PathValue(name)
	if param == "" {
		if param = r.URL.Query().Get(name); param == "" {
			return t, ErrMissingParam
		}
	}
	i, err := parser(param)
	if err != nil {
		return t, fmt.Errorf("parse: %w", err)
	}

	return i, nil
}

// ParseInt64 will attempt to parse the given input into an int64 number with
// the base of 10. It is meant to be used as an argument to the [Param] func.
func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Go spawns a new goroutine and will recover in case of panic; logging the
// error message in Error level. Using this function ensures panicking in other
// goroutines will not stop the main goroutine.
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
