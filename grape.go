// Package grape is a modern, zero-dependency HTTP library for Go.
package grape

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

// Map is an alias type and is intended for json response marshalling.
type Map = map[string]any

// Server is main struct of Grape with three embedded private types; logger, serializer and response.
// It should be included in the same struct that your handlers are a method to,
// so the helper methods are accessible through the receiver.
type Server struct {
	logger
	serializer
	response
}

// Options is used to customize Grape's settings, namely Log and Serialize.
// If a field is not provided, the default will be used instead.
type Options struct {
	Log       *slog.Logger
	Serialize serializer
}

var defaultOptions = Options{
	// default logger displays text logs to the standard output and in `info` level
	Log: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	// default serialize uses standard library `encoding/json`
	Serialize: serialize{},
}

// New return an instance of grape.Server to be included in structs.
// It optionally accepts grape.Option to customize Grape's settings.
func New(opts ...Options) Server {
	var opt Options
	if len(opts) == 0 {
		opt = defaultOptions
	} else {
		opt = opts[0]

		if opt.Log == nil {
			opt.Log = defaultOptions.Log
		}

		if opt.Serialize == nil {
			opt.Serialize = defaultOptions.Serialize
		}
	}

	return Server{
		serializer: opt.Serialize,
		logger:     logger{opt.Log},
		response:   newResponse(logger{opt.Log}, opt.Serialize),
	}
}

// ParamInt extracts the parameter by its name from request and converts it to integer.
// It will return 0 if no parameter was found, or there was an error converting it to int.
func (server Server) ParamInt(r *http.Request, name string) int {
	param, err := strconv.Atoi(r.PathValue(name))
	if err != nil {
		return 0
	}
	return param
}

// ParamInt64 extracts the parameter by its name from request and converts it to integer.
// It will return 0 if no parameter was found, or there was an error converting it to int64.
func (server Server) ParamInt64(r *http.Request, name string) int64 {
	param, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return 0
	}
	return param
}
