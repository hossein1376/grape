// Package errs exposes Error object which implements the error interface.
// Helper functions describing different error scenarios are provided.
package errs

import (
	"fmt"
	"net/http"
)

// Error represents an error which will be passed up through the application
// layers.
type Error struct {
	Err            error
	HTTPStatusCode int
	Message        string
}

// New returns a new instance of the [Error] object.
func New(httpStatusCode int, opts ...Options) Error {
	e := Error{
		HTTPStatusCode: httpStatusCode,
	}
	for _, opt := range opts {
		opt(&e)
	}
	return e
}

// Error returns error's text, prefixed by its HTTP status code. If the underlying
// error is nil, then the default text of that status code is used.
// Example:
//
//	[400] Bad Request
//	[404] missing record
func (e Error) Error() string {
	var errMsg string
	if e.Err == nil {
		errMsg = http.StatusText(e.HTTPStatusCode)
	} else {
		errMsg = e.Err.Error()
	}
	return fmt.Sprintf("[%d] %s", e.HTTPStatusCode, errMsg)
}

// Unwrap returns the underlying Err.
func (e Error) Unwrap() error {
	return e.Err
}

// BadRequest indicates client has provided invalid arguments, and must
// correct them before retrying.
//
// HTTP: 400
func BadRequest(opts ...Options) Error {
	return New(http.StatusBadRequest, opts...)

}

// Unauthorized indicates the request does not have valid authentication
// credentials for the operation.
//
// HTTP: 401
func Unauthorized(opts ...Options) Error {
	return New(http.StatusUnauthorized, opts...)
}

// Forbidden indicates the caller does not have permission to execute
// the specified operation.
//
// HTTP: 403
func Forbidden(opts ...Options) Error {
	return New(http.StatusForbidden, opts...)
}

// NotFound means some requested entity was not found.
//
// HTTP: 404
func NotFound(opts ...Options) Error {
	return New(http.StatusNotFound, opts...)
}

// Conflict indicates operation was rejected because the request is in conflict
// with the system's current state.
//
// HTTP: 409
func Conflict(opts ...Options) Error {
	return New(http.StatusConflict, opts...)
}

// TooMany indicates some resource has been exhausted, and client may need to
// wait some time before retrying.
//
// HTTP: 429
func TooMany(opts ...Options) Error {
	return New(http.StatusTooManyRequests, opts...)
}

// Internal means something has gone wrong in the server's side.
//
// HTTP: 500
func Internal(opts ...Options) Error {
	return New(http.StatusInternalServerError, opts...)
}

// BadGateway indicates a remote service is currently unreachable.
//
// HTTP: 502
func BadGateway(opts ...Options) Error {
	return New(http.StatusBadGateway, opts...)
}

// Timeout means a timeout has been reached. The operation may have been
// completed successfully or not.
//
// HTTP: 504
func Timeout(opts ...Options) Error {
	return New(http.StatusGatewayTimeout, opts...)
}
