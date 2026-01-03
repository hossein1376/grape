package errs

type Options func(*Error)

// WithErr sets the underlying error for the custom [Error] type. This error is
// considered to be internal and not exposed to the end user.
// It can be used to wrap an existing error with additional context.
func WithErr(err error) Options {
	return func(e *Error) {
		e.Err = err
	}
}

// WithMsg sets the message for the custom [Error] type. This message is intended
// to be user-friendly and can be displayed to the end user.
func WithMsg(msg string) Options {
	return func(e *Error) {
		e.Message = msg
	}
}

// WithHTTPStatus sets the HTTP status code for the custom [Error] type. This
// status code can be used when returning the error in an HTTP response.
func WithHTTPStatus(statusCode int) Options {
	return func(e *Error) {
		e.HTTPStatusCode = statusCode
	}
}

// WithErrMsg sets the message of the custom [Error] type to the message of the
// provided error. This message is intended to be user-friendly and can be
// displayed to the end user.[]
// If the provided error is nil, the message is set to an empty string.
func WithErrMsg(err error) Options {
	return func(e *Error) {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		e.Message = msg
	}
}
