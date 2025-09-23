package grape

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/reqid"
	"github.com/hossein1376/grape/slogger"
)

type Response struct {
	Message any `json:"message"`
	Data    any `json:"data,omitempty"`
}

type ResponseFormat int

const (
	JSON ResponseFormat = iota
)

type responseOptions struct {
	format ResponseFormat
}

type ResponseOption func(*responseOptions)

func WithResponseFormat(format ResponseFormat) ResponseOption {
	return func(o *responseOptions) {
		o.format = format
	}
}

// Respond is a general function which responses with the provided message
// and status code. It acts as an abstraction over WriteJson.
func Respond(
	ctx context.Context,
	w http.ResponseWriter,
	statusCode int,
	data any,
	opts ...ResponseOption,
) {
	if ctx == nil {
		ctx = context.Background()
	}

	opt := responseOptions{}
	for _, o := range opts {
		o(&opt)
	}

	wOpts := []WriteOpts{WithStatus(statusCode)}
	if data != nil {
		wOpts = append(wOpts, WithData(data))
	}

	switch opt.format {
	case JSON:
		WriteJson(ctx, w, wOpts...)
	}
}

// RespondFromErr extracts a response from the given error. If nil, 204 response
// is returned. If error is of type [errs.Error], the status code and response
// message are filled accordingly. Otherwise, a 500 response with the request ID
// are returned.
func RespondFromErr(
	ctx context.Context,
	w http.ResponseWriter,
	err error,
	opts ...ResponseOption,
) {
	if err == nil {
		Respond(ctx, w, http.StatusNoContent, nil, opts...)
		return
	}

	var e errs.Error
	if errors.As(err, &e) {
		msg := e.Message
		if msg == "" {
			msg = http.StatusText(e.HTTPStatusCode)
		}
		Respond(ctx, w, e.HTTPStatusCode, Response{Message: msg}, opts...)
		slogger.Debug(
			ctx,
			"failed request",
			slogger.Err("error", err),
			slog.Int("status_code", e.HTTPStatusCode),
			slog.String("message", msg),
		)
		return
	}

	slogger.Error(ctx, "internal error", slogger.Err("error", err))
	reqID, _ := reqid.RequestID(ctx)
	Respond(
		ctx,
		w,
		http.StatusInternalServerError,
		Response{
			Message: http.StatusText(http.StatusInternalServerError),
			Data:    reqID,
		},
		opts...,
	)
}
