package grape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/slogger"
	"github.com/hossein1376/grape/validator"
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

type readOptions struct {
	maxBodySize int64
}

type ReadOpts func(*readOptions)

// WithMaxBodySize sets the maximum accepted body size in bytes for [ReadJSON]
// and [ReadValidateJSON].
func WithMaxBodySize(size int64) ReadOpts {
	return func(o *readOptions) {
		o.maxBodySize = size
	}
}

// ReadJSON will decode incoming json requests. By default, the maximum body
// size is 1MB, which can be changed using WithMaxBodySize option.
//
// Errors are [errs.Error] values with a human-readable message, so
// [ExtractFromErr] responds with it: 415 if the content type isn't JSON, 413
// if the body is too large, and 400 for anything else wrong with the body.
// The underlying error, e.g. a [*json.SyntaxError], is kept in the chain.
func ReadJSON[T any](
	w http.ResponseWriter, r *http.Request, opts ...ReadOpts,
) (*T, error) {
	if !isJSON(r.Header.Get("Content-Type")) {
		return nil, bodyErr(
			http.StatusUnsupportedMediaType,
			"content type must be application/json",
			nil,
		)
	}
	opt := &readOptions{maxBodySize: defaultMaxBodySize}
	for _, o := range opts {
		o(opt)
	}
	r.Body = http.MaxBytesReader(w, r.Body, opt.maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	dst := new(T)
	if err := dec.Decode(dst); err != nil {
		return nil, decodeErr(err, opt.maxBodySize)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if isTooLarge(err) {
			return nil, decodeErr(err, opt.maxBodySize)
		}
		return nil, bodyErr(
			http.StatusBadRequest,
			"body must only contain a single JSON value",
			err,
		)
	}
	return dst, nil
}

// isJSON reports whether the content type is application/json, or a JSON
// based one such as application/merge-patch+json.
func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == "application/json" ||
		strings.HasPrefix(mediaType, "application/") &&
			strings.HasSuffix(mediaType, "+json")
}

func isTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// bodyErr returns an [errs.Error] with msg as its message. The underlying
// error, if any, is wrapped for logs and errors.Is and errors.As.
func bodyErr(status int, msg string, err error) error {
	cause := errors.New(msg)
	if err != nil {
		cause = fmt.Errorf("%s: %w", msg, err)
	}
	return errs.New(status, errs.WithErr(cause), errs.WithMsg(msg))
}

func decodeErr(err error, maxBodySize int64) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	switch {
	case isTooLarge(err):
		return bodyErr(
			http.StatusRequestEntityTooLarge,
			fmt.Sprintf("body must not be larger than %d bytes", maxBodySize),
			err,
		)
	case errors.Is(err, io.EOF):
		return bodyErr(http.StatusBadRequest, "body must not be empty", err)
	case errors.Is(err, io.ErrUnexpectedEOF):
		return bodyErr(
			http.StatusBadRequest, "body contains badly-formed JSON", err,
		)
	case errors.As(err, &syntaxError):
		return bodyErr(
			http.StatusBadRequest,
			fmt.Sprintf(
				"body contains badly-formed JSON (at character %d)",
				syntaxError.Offset,
			),
			err,
		)
	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			return bodyErr(
				http.StatusBadRequest,
				fmt.Sprintf(
					"body contains incorrect JSON type for field %q",
					unmarshalTypeError.Field,
				),
				err,
			)
		}
		return bodyErr(
			http.StatusBadRequest,
			fmt.Sprintf(
				"body contains incorrect JSON type (at character %d)",
				unmarshalTypeError.Offset,
			),
			err,
		)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.Trim(
			strings.TrimPrefix(err.Error(), "json: unknown field "), "\"",
		)
		return bodyErr(
			http.StatusBadRequest,
			fmt.Sprintf("body contains unknown key %q", fieldName),
			err,
		)
	default:
		return bodyErr(http.StatusBadRequest, "unable to parse body", err)
	}
}

// ReadValidateJSON reads and validates the json input. If T is a pointer type,
// a JSON null body is rejected rather than calling Validate on a nil pointer.
//
// Errors from reading are as described in [ReadJSON]. Errors from Validate are
// client errors too: a [validator.ValidationError] is kept, so
// [ExtractFromErr] sends its field errors, and an [errs.Error] is kept as is.
// Any other error becomes a 400 [errs.Error] with its text as the message.
func ReadValidateJSON[T Validator](
	w http.ResponseWriter, r *http.Request, opts ...ReadOpts,
) (*T, error) {
	t, err := ReadJSON[T](w, r, opts...)
	if err != nil {
		return nil, err
	}
	if v := reflect.ValueOf(*t); v.Kind() == reflect.Pointer && v.IsNil() {
		return nil, bodyErr(http.StatusBadRequest, "body must not be null", nil)
	}
	if err := (*t).Validate(); err != nil {
		var fields validator.ValidationError
		if _, ok := asError(err); ok || errors.As(err, &fields) {
			return nil, fmt.Errorf("validate: %w", err)
		}
		return nil, errs.BadRequest(
			errs.WithErr(fmt.Errorf("validate: %w", err)),
			errs.WithMsg(err.Error()),
		)
	}
	return t, nil
}
