package grape

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	logger
	serialize
	response
}

type Options struct {
	Log *slog.Logger
}

var defaultOptions = Options{
	Log: newTextLogger(os.Stdout, slog.LevelInfo),
}

func New(opts ...Options) *Server {
	var opt Options
	if len(opts) == 0 {
		opt = defaultOptions
	} else {
		opt = opts[0]
	}

	logging := logger{opt.Log}
	serializer := serialize{}
	return &Server{
		serialize: serializer,
		logger:    logging,
		response:  newResponse(logging, serializer),
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
