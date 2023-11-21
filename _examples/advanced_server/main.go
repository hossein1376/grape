package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/hossein1376/grape"
)

type handler struct {
	// data/models
	// settings
	grape.Server
}

func main() {
	// Any valid *slog.Logger will do! You can configure the log format, destination, level and more
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Any valid grape.serializer will do!
	serializer := Jsoniter{}

	// If you don't provide a field, default value (grape.defaultOptions) will be used
	opts := grape.Options{Log: logger, Serialize: serializer}

	// Instantiate the grape.Server inside your struct of choice
	h := handler{
		// Models: models.New()
		// Settings: settings.Get()
		Server: grape.New(opts),
	}

	// You can define routes in a separate function
	r := h.router()

	h.Info("starting server on port 3000...")

	// You can optionally pass an instance of *http.Server to configure the running settings
	// Note that two fields `Addr` and `Handler` are set by the Serve method. If provided, they'll be overwritten.
	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	err := r.Serve(":3000", srv)
	if err != nil {
		h.Error("failed to start server", "error", err)
		return
	}
}
