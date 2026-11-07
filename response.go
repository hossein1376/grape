package grape

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/reqid"
	"github.com/hossein1376/grape/slogger"
	"github.com/hossein1376/grape/validator"
)

type Response struct {
	Message any `json:"message,omitempty"`
	Data    any `json:"data,omitempty"`
}

// Respond is a general function which responses with the provided message
// and status code. It acts as an abstraction over WriteJson.
func Respond(
	ctx context.Context, w http.ResponseWriter, statusCode int, data any,
) {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := []WriteOpts{WithStatus(statusCode)}
	if data != nil {
		opts = append(opts, WithData(data))
	}

	WriteJSON(ctx, w, opts...)
}

// ExtractFromErr extracts a response from the given error. If nil, 204 response
// is returned. If error is of type [errs.Error] or *[errs.Error], with a valid
// HTTP status code, the status code and response message are filled
// accordingly. Otherwise, if error is a [validator.ValidationError], a 400
// response with the field errors as data is returned:
//
//	{"message": "Bad Request", "data": {"age": ["must be at least 0"]}}
//
// Otherwise, a 500 response with the request ID are returned.
//
// [errs.Error] takes precedence: a ValidationError wrapped in it with
// [errs.WithErr] is considered internal, and its fields aren't sent.
func ExtractFromErr(ctx context.Context, w http.ResponseWriter, err error) {
	if err == nil {
		Respond(ctx, w, http.StatusNoContent, nil)
		return
	}

	if e, ok := asError(err); ok {
		msg := e.Message
		if msg == "" {
			msg = http.StatusText(e.HTTPStatusCode)
		}
		Respond(ctx, w, e.HTTPStatusCode, Response{Message: msg})
		slogger.Debug(
			ctx,
			"failed request",
			slogger.Err("error", err),
			slog.Int("status_code", e.HTTPStatusCode),
			slog.String("message", msg),
		)
		return
	}

	var fields validator.ValidationError
	if errors.As(err, &fields) && len(fields) != 0 {
		Respond(
			ctx,
			w,
			http.StatusBadRequest,
			Response{
				Message: http.StatusText(http.StatusBadRequest),
				Data:    fields,
			},
		)
		slogger.Debug(ctx, "validation failed", slogger.Err("error", err))
		return
	}

	slogger.Error(ctx, "internal error", slogger.Err("error", err))
	respondInternal(ctx, w)
}

// respondInternal writes a 500 response with the request ID as data, so it
// can be matched with the logs.
func respondInternal(ctx context.Context, w http.ResponseWriter) {
	reqID, _ := reqid.RequestID(ctx)
	Respond(
		ctx,
		w,
		http.StatusInternalServerError,
		Response{
			Message: http.StatusText(http.StatusInternalServerError),
			Data:    reqID,
		},
	)
}

// asError finds the first [errs.Error] or *[errs.Error] in err's tree. It
// reports false if there's none, or its status code can't be written, which
// [http.ResponseWriter.WriteHeader] would panic on.
func asError(err error) (errs.Error, bool) {
	var e errs.Error
	if !errors.As(err, &e) {
		var pe *errs.Error
		if !errors.As(err, &pe) || pe == nil {
			return errs.Error{}, false
		}
		e = *pe
	}
	if e.HTTPStatusCode < 100 || e.HTTPStatusCode > 999 {
		return errs.Error{}, false
	}
	return e, true
}
