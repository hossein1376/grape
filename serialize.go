package grape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		for key, values := range headers {
			o.headers[key] = values
		}
	}
}

// WriteJson will write back data in json format with the provided status code
// and headers. It automatically sets content-type and date headers. To override,
// provide them as headers.
func WriteJson(ctx context.Context, w http.ResponseWriter, opts ...WriteOpts) {
	opt := defaultWriteOptions()
	for _, o := range opts {
		o(opt)
	}
	for key, value := range opt.headers {
		w.Header()[key] = value
	}

	if opt.data == nil {
		w.WriteHeader(opt.status)
		return
	}

	js, err := json.Marshal(opt.data)
	if err != nil {
		slogger.Error(ctx, "marshal data", slogger.Err("error", err))
		return
	}
	w.WriteHeader(opt.status)
	if _, err = w.Write(js); err != nil {
		slogger.Error(ctx, "write response", slogger.Err("error", err))
		return
	}
	return
}

type readOptions struct {
	maxBodySize int64
}

type ReadOpts func(*readOptions)

func WithMaxBodySize(size int64) func(*readOptions) {
	return func(o *readOptions) {
		o.maxBodySize = size
	}
}

func defaultReadOptions() *readOptions {
	return &readOptions{maxBodySize: defaultMaxBodySize}
}

// ReadJson will decode incoming json requests. It will return a
// human-readable error in case of failure.
func ReadJson(
	w http.ResponseWriter, r *http.Request, dst any, opts ...ReadOpts,
) error {
	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
		return errors.New("content type is not application/json")
	}

	opt := defaultReadOptions()
	for _, o := range opts {
		o(opt)
	}
	r.Body = http.MaxBytesReader(w, r.Body, opt.maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err == nil {
		if err = dec.Decode(&struct{}{}); err != io.EOF {
			return errors.New("body must only contain a single JSON value")
		}
		return nil
	}

	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	switch {
	case errors.Is(err, io.EOF):
		return errors.New("body must not be empty")
	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("body contains badly-formed JSON")
	case errors.As(err, &syntaxError):
		return fmt.Errorf(
			"body contains badly-formed JSON (at character %d)",
			syntaxError.Offset,
		)
	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf(
				"body contains incorrect JSON type for field %q",
				unmarshalTypeError.Field,
			)
		}
		return fmt.Errorf(
			"body contains incorrect JSON type (at character %d)",
			unmarshalTypeError.Offset,
		)
	case err.Error() == "http: request body too large":
		return fmt.Errorf("body must not be larger than %d bytes", opt.maxBodySize)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return fmt.Errorf("body contains unknown key %s", fieldName)
	default:
		return err
	}
}
