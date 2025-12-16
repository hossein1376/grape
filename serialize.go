package grape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/hossein1376/grape/slogger"
)

type writeOption struct {
	status  int
	data    any
	headers http.Header
}

func defaultWriteOptions() *writeOption {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Date", time.Now().Format(http.TimeFormat))
	return &writeOption{
		status:  http.StatusOK,
		headers: headers,
	}
}

type WriteOpts func(*writeOption)

func WithStatus(statusCode int) func(*writeOption) {
	return func(o *writeOption) {
		o.status = statusCode
	}
}

func WithData(data any) func(*writeOption) {
	return func(o *writeOption) {
		o.data = data
	}
}

func WithHeaders(headers http.Header) func(*writeOption) {
	return func(o *writeOption) {
		maps.Copy(o.headers, headers)
	}
}

// WriteJSON will write back data in json format with the provided status code
// and headers. It automatically sets content-type and date headers. To override,
// provide them as headers.
func WriteJSON(ctx context.Context, w http.ResponseWriter, opts ...WriteOpts) {
	opt := defaultWriteOptions()
	for _, o := range opts {
		o(opt)
	}
	maps.Copy(w.Header(), opt.headers)

	if opt.data == nil {
		w.WriteHeader(opt.status)
		return
	}

	js, err := json.Marshal(opt.data)
	if err != nil {
		slogger.Error(ctx, "marshal data", slogger.Err("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(opt.status)
	if _, err = w.Write(js); err != nil {
		slogger.Error(ctx, "write response", slogger.Err("error", err))
		return
	}
}

type readOptions[T any] struct {
	maxBodySize int64
}

type ReadOpts[T any] func(*readOptions[T])

func WithMaxBodySize(size int64) func(*readOptions[any]) {
	return func(o *readOptions[any]) {
		o.maxBodySize = size
	}
}

// ReadJSON will decode incoming json requests. It will return a human-readable
// error in case of failure. If [T] implements the following method:
//
//	Validate() error
//
// it will be called after decoding. By default, the maximum body size is 1MB,
// which can be changed using WithMaxBodySize option.
func ReadJSON[T any](
	w http.ResponseWriter, r *http.Request, opts ...ReadOpts[T],
) (*T, error) {
	ct := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if !strings.HasPrefix(ct, "application/json") {
		return nil, errors.New("content type is not application/json")
	}
	opt := &readOptions[T]{maxBodySize: defaultMaxBodySize}
	for _, o := range opts {
		o(opt)
	}
	r.Body = http.MaxBytesReader(w, r.Body, opt.maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	dst := new(T)
	err := dec.Decode(dst)
	if err == nil {
		if err = dec.Decode(&struct{}{}); err != io.EOF {
			return nil, errors.New("body must only contain a single JSON value")
		}
		if v, ok := any(dst).(interface{ Validate() error }); ok {
			err = v.Validate()
		}
		return dst, err
	}

	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	switch {
	case errors.Is(err, io.EOF):
		return nil, errors.New("body must not be empty")
	case errors.Is(err, io.ErrUnexpectedEOF):
		return nil, errors.New("body contains badly-formed JSON")
	case errors.As(err, &syntaxError):
		return nil, fmt.Errorf(
			"body contains badly-formed JSON (at character %d)",
			syntaxError.Offset,
		)
	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return nil, fmt.Errorf(
				"body contains incorrect JSON type for field %q",
				unmarshalTypeError.Field,
			)
		}
		return nil, fmt.Errorf(
			"body contains incorrect JSON type (at character %d)",
			unmarshalTypeError.Offset,
		)
	case err.Error() == "http: request body too large":
		return nil, fmt.Errorf(
			"body must not be larger than %d bytes", opt.maxBodySize,
		)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.Trim(
			strings.TrimPrefix(err.Error(), "json: unknown field "), "\"")
		return nil, fmt.Errorf("body contains unknown key %q", fieldName)
	default:
		return nil, fmt.Errorf("unable to parse body: %w", err)
	}
}
