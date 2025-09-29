package errs

type Options func(*Error)

func WithMsg(msg string) Options {
	return func(e *Error) {
		e.Message = msg
	}
}

func WithHTTPStatus(statusCode int) Options {
	return func(e *Error) {
		e.HTTPStatusCode = statusCode
	}
}
