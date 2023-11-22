package grape

import (
	"net/http"
)

type response struct {
	logger
	serializer
}

func newResponse(logger logger, json serializer) response {
	return response{logger: logger, serializer: json}
}

type resp struct {
	Message any `json:"message,omitempty"`
}

// Response is a general function which responses with the provided message and status code,
// it will return 500 if case of failure
func (res response) Response(w http.ResponseWriter, statusCode int, message any) {
	err := res.WriteJson(w, statusCode, message, nil)
	if err != nil {
		res.Error("writing json response", "error", err, "data", message)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	}
}

// OkResponse means everything went as expected
func (res response) OkResponse(w http.ResponseWriter, data any) {
	res.Response(w, http.StatusOK, data)
}

// CreatedResponse indicates that requested resource(s) have been successfully created
func (res response) CreatedResponse(w http.ResponseWriter, data any) {
	res.Response(w, http.StatusCreated, data)
}

// NoContentResponse means the operation was successful, and server has nothing more to say about it
func (res response) NoContentResponse(w http.ResponseWriter) {
	res.Response(w, http.StatusNoContent, nil)
}

// BadRequestResponse indicates that the request has been deemed unacceptable by server
func (res response) BadRequestResponse(w http.ResponseWriter, err ...error) {
	msg := http.StatusText(http.StatusBadRequest)
	if len(err) != 0 {
		msg = err[0].Error()
	}

	r := resp{Message: msg}
	res.Response(w, http.StatusBadRequest, r)
}

// UnauthorizedResponse responds when user is not authorized
func (res response) UnauthorizedResponse(w http.ResponseWriter) {
	r := resp{Message: http.StatusText(http.StatusUnauthorized)}
	res.Response(w, http.StatusUnauthorized, r)
}

// ForbiddenResponse indicates that the action is not allowed
func (res response) ForbiddenResponse(w http.ResponseWriter) {
	r := resp{Message: http.StatusText(http.StatusForbidden)}
	res.Response(w, http.StatusForbidden, r)
}

// NotFoundResponse will return with classic 404 error message.
// if error message is provided, it will return that instead.
func (res response) NotFoundResponse(w http.ResponseWriter, err ...error) {
	msg := http.StatusText(http.StatusNotFound)
	if len(err) != 0 {
		msg = err[0].Error()
	}

	r := resp{Message: msg}
	res.Response(w, http.StatusNotFound, r)
}

// InternalServerErrorResponse indicates something has gone wrong unexpectedly.
func (res response) InternalServerErrorResponse(w http.ResponseWriter) {
	r := resp{Message: http.StatusText(http.StatusInternalServerError)}
	res.Response(w, http.StatusInternalServerError, r)
}
