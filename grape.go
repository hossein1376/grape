// Package grape is a modern, zero-dependency HTTP library for Go.
// Visit https://github.com/hossein1376/grape for more information.
package grape

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

const defaultMaxBodySize = 1_048_576 // 1mb

var (
	ErrMissingParam = errors.New("path parameter not found")
	ErrMissingQuery = errors.New("query parameter not found")
	ErrOverflow     = errors.New("value overflow")
)

type (
	Map           = map[string]any
	Parser[T any] func(s string) (T, error)
)

// Param attempts to extract the given parameter from path. Then, the parser
// function is used to parse it to the expected type. If nil parser is provided,
// it will attempt to extract and use the Parse method on the type T.
//
// Some of the parser functions for primitive types are: [strconv.Atoi],
// [strconv.ParseBool], [ParseInt], [ParseUint], and [ParseFloat]. It can also
// be manually implemented for custom types.
//
// Examples:
//
//		grape.Param(r, "boolean", strconv.ParseBool)
//		grape.Param(r, "integer", grape.ParseInt[int16]())
//		grape.Param(r, "float", grape.ParseFloat[float64]())
//
//		grape.Param[Custom](r, "custom", nil)
//	 /*
//		type Custom struct {}
//
//		func (c Custom) Parse(s string) (Custom, error) {
//			// custom parsing logic
//		}
//	 */
func Param[T any](r *http.Request, name string, parser Parser[T]) (T, error) {
	var t T
	param := r.PathValue(name)
	if param == "" {
		return t, fmt.Errorf("%w: %s", ErrMissingParam, name)
	}
	return parse(t, param, parser)
}

// Query attempts to extract and parse the given query parameter. For more
// details, refer to the [Param] func.
func Query[T any](query url.Values, name string, parser Parser[T]) (T, error) {
	var t T
	param := query.Get(name)
	if param == "" {
		return t, fmt.Errorf("%w: %s", ErrMissingQuery, name)
	}
	return parse(t, param, parser)
}

func parse[T any](t T, param string, parser Parser[T]) (T, error) {
	if parser == nil {
		p, ok := any(t).(interface{ Parse(s string) (T, error) })
		if !ok {
			return t, fmt.Errorf("no parser provided for type %T", t)
		}
		return p.Parse(param)
	}
	i, err := parser(param)
	if err != nil {
		return t, fmt.Errorf("parse: %w", err)
	}

	return i, nil
}

// ParseInt will attempt to parse the given input into an integer of the
// specified type. It can be used as an argument to the [Param] function.
func ParseInt[
	T ~int | ~int64 | ~int32 | ~int16 | ~int8,
]() func(string) (T, error) {
	return func(s string) (T, error) {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return checkOverflow(s, T(i))
	}
}

// ParseUint will attempt to parse the given input into an unsigned integer of
// the specified type. It can be used as an argument to the [Param] function.
func ParseUint[
	T ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8,
]() func(string) (T, error) {
	return func(s string) (T, error) {
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return checkOverflow(s, T(i))
	}
}

// ParseFloat will attempt to parse the given input into a floating point number
// of the specified type. It can be used as an argument to the [Param] function.
func ParseFloat[T ~float64 | ~float32]() func(string) (T, error) {
	return func(s string) (T, error) {
		i, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return checkOverflow(s, T(i))
	}
}

func checkOverflow[T any](s string, t T) (T, error) {
	if fmt.Sprintf("%v", t) != s {
		return t, ErrOverflow
	}
	return t, nil
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
