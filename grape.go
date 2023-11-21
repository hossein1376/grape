package grape

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type Map = map[string]any

type Server struct {
	logger
	serializer
	response
}

type Options struct {
	Log       *slog.Logger
	Serialize serializer
}

var defaultOptions = Options{
	Log:       newTextLogger(os.Stdout, slog.LevelInfo),
	Serialize: serialize{},
}

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

func (s *Server) ParamInt(r *http.Request, name string) int {
	param, err := strconv.Atoi(r.PathValue(name))
	if err != nil {
		return 0
	}
	return param
}

func (s *Server) ParamInt64(r *http.Request, name string) int64 {
	param, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return 0
	}
	return param
}
