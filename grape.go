// Package grape is a modern, zero-dependency HTTP library for Go.
// Visit https://github.com/hossein1376/grape for more information.
package grape

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

const defaultMaxBodySize = 1_048_576 // 1mb

var (
	ErrMissingParam = errors.New("path parameter not found")
	ErrMissingQuery = errors.New("query parameter not found")
	ErrNilParser    = errors.New("nil parser provided")
	ErrOverflow     = errors.New("value overflow")
)

type (
	Map           = map[string]any
	Parser[T any] func(s string) (T, error)
	Validator     interface{ Validate() error }
)

// Param attempts to extract the given parameter from path. Then, the parser
// function is used to parse it to the expected type.
//
// Some of the parser functions for primitive types are: [strconv.Atoi],
// [strconv.ParseBool], [ParseInt], [ParseUint], and [ParseFloat].
//
// Examples:
//
//		grape.Param(r, "boolean", strconv.ParseBool)
//		grape.Param(r, "integer", grape.ParseInt[int16]())
//		grape.Param(r, "float", grape.ParseFloat[float64]())
//
//		grape.Param(r, "custom", ParseCustom)
//	 /*
//		type Custom struct {}
//		func ParseCustom(s string) (Custom, error) {}
//	 */
func Param[T any](r *http.Request, name string, parser Parser[T]) (T, error) {
	var t T
	if parser == nil {
		return t, ErrNilParser
	}
	param := r.PathValue(name)
	if param == "" {
		return t, fmt.Errorf("%w: %s", ErrMissingParam, name)
	}
	value, err := parser(param)
	if err != nil {
		return t, fmt.Errorf("parse: %w", err)
	}
	return value, nil
}

// ParamOrDefault calls [Param] under the hood, and will return the provided
// default value if there's any error.
func ParamOrDefault[T any](
	r *http.Request, name string, parser Parser[T], def T,
) T {
	if value, err := Param(r, name, parser); err == nil {
		return value
	}
	return def
}

// Query attempts to extract and parse the given query parameter. For more
// details, refer to the [Param] func.
func Query[T any](query url.Values, name string, parser Parser[T]) (T, error) {
	var t T
	if parser == nil {
		return t, ErrNilParser
	}
	param := query.Get(name)
	if param == "" {
		return t, fmt.Errorf("%w: %s", ErrMissingQuery, name)
	}
	value, err := parser(param)
	if err != nil {
		return t, fmt.Errorf("parse: %w", err)
	}
	return value, nil
}

// QueryOrDefault calls [Query] under the hood, and will return the provided
// default value if there's any error.
func QueryOrDefault[T any](
	query url.Values, name string, parser Parser[T], def T,
) T {
	if value, err := Query(query, name, parser); err == nil {
		return value
	}
	return def
}

// ParseInt will attempt to parse the given input into an integer of the
// specified type. It can be used as an argument to the [Param] or similar
// functions. [ErrOverflow] is returned if the value doesn't fit in the type.
func ParseInt[
	T ~int | ~int64 | ~int32 | ~int16 | ~int8,
]() func(string) (T, error) {
	bits := reflect.TypeFor[T]().Bits()
	return func(s string) (T, error) {
		i, err := strconv.ParseInt(s, 10, bits)
		if err != nil {
			return 0, rangeErr[T](s, err)
		}
		return T(i), nil
	}
}

// ParseUint will attempt to parse the given input into an unsigned integer of
// the specified type. It can be used as an argument to the [Param] or similar
// functions. [ErrOverflow] is returned if the value doesn't fit in the type.
func ParseUint[
	T ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8,
]() func(string) (T, error) {
	bits := reflect.TypeFor[T]().Bits()
	return func(s string) (T, error) {
		i, err := strconv.ParseUint(s, 10, bits)
		if err != nil {
			return 0, rangeErr[T](s, err)
		}
		return T(i), nil
	}
}

// ParseFloat will attempt to parse the given input into a floating point number
// of the specified type. It can be used as an argument to the [Param] or
// similar functions. [ErrOverflow] is returned if the value is out of the
// type's range. The value is rounded to the type's precision, as
// [strconv.ParseFloat] does.
func ParseFloat[T ~float64 | ~float32]() func(string) (T, error) {
	bits := reflect.TypeFor[T]().Bits()
	return func(s string) (T, error) {
		f, err := strconv.ParseFloat(s, bits)
		if err != nil {
			return 0, rangeErr[T](s, err)
		}
		return T(f), nil
	}
}

// rangeErr reports [strconv.ErrRange] as [ErrOverflow], and returns other
// errors as is.
func rangeErr[T any](s string, err error) error {
	if errors.Is(err, strconv.ErrRange) {
		var t T
		return fmt.Errorf("%w: %s doesn't fit in %T", ErrOverflow, s, t)
	}
	return err
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
