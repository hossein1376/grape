package grape

import (
	"log/slog"
	"os"
)

type Server struct {
	logger
	serialize
	*response
}

func New() *Server {
	logging := newTextLogger(os.Stdout, slog.LevelInfo)
	return &Server{
		serialize: serialize{},
		logger:    logging,
		response:  newResponse(logging, serialize{}),
	}
}
