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
// is returned. If error is of type [errs.Error], the status code and response
// message are filled accordingly. Otherwise, a 500 response with the request ID
// are returned.
func ExtractFromErr(ctx context.Context, w http.ResponseWriter, err error) {
	if err == nil {
		Respond(ctx, w, http.StatusNoContent, nil)
		return
	}

	var e errs.Error
	if errors.As(err, &e) {
		msg := e.Message
		if msg == "" {
			msg = e.Err.Error()
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
	)
}
