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
func New(err error, httpStatusCode int, opts ...Options) Error {
	e := Error{
		Err:            err,
		HTTPStatusCode: httpStatusCode,
	}
	for _, opt := range opts {
		opt(&e)
	}
	return e
}

// Error returns error's text, prefixed by its HTTP status code.
// Example:
//
//	[400] Bad Input
func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.HTTPStatusCode, e.Err)
}

// Unwrap returns the underlying Err.
func (e Error) Unwrap() error {
	return e.Err
}

// BadRequest indicates client has provided invalid arguments, and must
// correct them before retrying.
//
// HTTP: 400
func BadRequest(err error, opts ...Options) Error {
	return New(err, http.StatusBadRequest, opts...)

}

// Unauthorized indicates the request does not have valid authentication
// credentials for the operation.
//
// HTTP: 401
func Unauthorized(err error, opts ...Options) Error {
	return New(err, http.StatusUnauthorized, opts...)
}

// Forbidden indicates the caller does not have permission to execute
// the specified operation.
//
// HTTP: 403
func Forbidden(err error, opts ...Options) Error {
	return New(err, http.StatusForbidden, opts...)
}

// NotFound means some requested entity was not found.
//
// HTTP: 404
func NotFound(err error, opts ...Options) Error {
	return New(err, http.StatusNotFound, opts...)
}

// Conflict indicates operation was rejected because the request is in conflict
// with the system's current state.
//
// HTTP: 409
func Conflict(err error, opts ...Options) Error {
	return New(err, http.StatusConflict, opts...)
}

// TooMany indicates some resource has been exhausted, and client may need to
// wait some time before retrying.
//
// HTTP: 429
func TooMany(err error, opts ...Options) Error {
	return New(err, http.StatusTooManyRequests, opts...)
}

// Internal means something has gone wrong in the server's side.
//
// HTTP: 500
func Internal(err error, opts ...Options) Error {
	return New(err, http.StatusInternalServerError, opts...)
}

// Timeout means a timeout has been reached. The operation may have been
// completed successfully or not.
//
// HTTP: 504
func Timeout(err error, opts ...Options) Error {
	return New(err, http.StatusGatewayTimeout, opts...)
}
