package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/slogger"
)

func main() {
	_ = slogger.NewDefault(slogger.WithLevel(slog.LevelDebug))
	router := grape.NewRouter()
	router.Get("/", rootHandler)

	// create an instance of *http.Server; and pass it down to the Serve method.
	srv := &http.Server{}

	// quit channel listens for the interrupt signal.
	// failure channel is used to communicate server's startup error.
	quit, failure := make(chan os.Signal, 1), make(chan error, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	slog.Debug("starting server")

	go func() {
		slog.Info("starting server on port 8000...")
		err := router.Serve(":8000", srv) // passing srv instead of nil
		if err != nil {
			failure <- err
			return
		}
	}()

	select {
	case <-quit:
		// after receiving the signal, gracefully stop the server.
		if err := srv.Shutdown(context.Background()); err != nil {
			slog.Error("graceful shutdown failed", slogger.Err("error", err))
			return
		}
		slog.Info("server was gracefully shutdown")

	case err := <-failure:
		slog.Error("failed to start server", "error", err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	grape.Respond(r.Context(), w, http.StatusOK, "root endpoint")
}
