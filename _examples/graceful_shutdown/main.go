package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hossein1376/grape"
)

type application struct {
	grape.Server
}

func main() {
	// principal is the same, naming is different.
	app := application{Server: grape.New()}
	router := grape.NewRouter()

	router.Get("/", app.rootHandler)
	group := router.Group("/group")
	group.Get("/", app.groupHandler)

	// create an instance of *http.Server; and pass it down to the Serve method.
	srv := &http.Server{}

	// quit channel listens for the interrupt signal.
	// failure channel is used to communicate server's startup error.
	quit, failure := make(chan os.Signal, 1), make(chan error, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		app.Info("starting server on port 8000...")
		err := router.Serve(":8000", srv) // passing srv as an optional argument
		if err != nil {
			failure <- err
			return
		}
	}()

	select {
	case <-quit:
		// after receiving the signal, gracefully stop the server.
		if err := srv.Shutdown(context.Background()); err != nil {
			app.Error("graceful shutdown failed", "err", err)
			return
		}
		app.Info("server was gracefully shutdown")

	case err := <-failure:
		app.Error("failed to start server", "error", err)
	}
}

func (app *application) rootHandler(w http.ResponseWriter, _ *http.Request) {
	app.OkResponse(w, "root endpoint")
}

func (app *application) groupHandler(w http.ResponseWriter, _ *http.Request) {
	app.OkResponse(w, "group endpoint")
}
