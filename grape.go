// Package grape is a modern, zero-dependency HTTP library for Go.
// Visit https://github.com/hossein1376/grape for more information.
package grape

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

// Map is an alias type and is intended for json response marshalling.
type Map = map[string]any

// Server is the main struct of Grape with three embedded types; Logger, Serializer and response.
// Main usage pattern is to included it in the same struct that your handlers are a method to,
// so the helper methods are accessible through the receiver.
type Server struct {
	Logger
	Serializer
	response
}

// Options is used to customize Grape's settings, by passing down an instance of it to grape.New().
//
// Log should implement grape.Logger. Since slog.Logger automatically does,
// all instances of it can be used. Alongside of any other custom types.
// Default logger displays text logs to the standard output and in `info` level.
//
// Serialize should implement grape.Serializer.
// Default serializer uses standard library `encoding/json` package.
//
// RequestMaxSize sets maximum request's body size in bytes. This value will be taken into effect only
// if Serialize field was not provided.
// Default size is 1_048_576 bytes (1 mb).
type Options struct {
	Log            Logger
	Serialize      Serializer
	RequestMaxSize int64
}

var defaultOptions = Options{
	Log:            slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	Serialize:      serializer{maxBytes: maxBodySize},
	RequestMaxSize: maxBodySize,
}

// New returns a new instance of grape.Server.
// It optionally accepts grape.Options to customize Grape's settings.
func New(opts ...Options) Server {
	var opt Options

	// if no options was provided, proceed with the default values.
	if len(opts) == 0 {
		opt = defaultOptions
	} else {
		opt = opts[0]

		// if Log field was not provided, use default logger.
		if opt.Log == nil {
			opt.Log = defaultOptions.Log
		}

		// if Serialize field was not provided, use default serializer.
		if opt.Serialize == nil {
			// if RequestMaxSize was not provided or an invalid value was given, use default value.
			if opt.RequestMaxSize <= 0 {
				opt.RequestMaxSize = defaultOptions.RequestMaxSize
			}

			opt.Serialize = serializer{maxBytes: opt.RequestMaxSize}
		}
	}

	return Server{
		Serializer: opt.Serialize,
		Logger:     opt.Log,
		response:   newResponse(opt.Log, opt.Serialize),
	}
}

// ParamInt extracts the parameter by its name from request and converts it to integer.
// It will return 0 and an error if no parameter was found, or there was an error converting it to int.
func (server Server) ParamInt(r *http.Request, name string) (int, error) {
	param, err := strconv.Atoi(r.PathValue(name))
	if err != nil {
		return 0, err
	}
	return param, nil
}

// ParamInt64 extracts the parameter by its name from request and converts it to 64-bit integer.
// It will return 0 and an error if no parameter was found, or there was an error converting it to int64.
func (server Server) ParamInt64(r *http.Request, name string) (int64, error) {
	param, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return param, nil
}

// Go spawns a new goroutine and will recover in case of panic; logging the error message in Error level.
// Using this function ensures that panic in goroutines will not stop the application's execution.
func (server Server) Go(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				server.Error("goroutine panic recovered", "error", err)
			}
		}()

		f()
	}()
}
