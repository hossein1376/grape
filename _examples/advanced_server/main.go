package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/slogger"
)

func main() {
	logger := slogger.New(
		slogger.WithLevel(slog.LevelDebug),
		slogger.WithDestination(os.Stdout),
		slogger.WithAddSource(),
		slogger.WithTextLogger(),
	)
	// You can use [slogger.NewDefault] too
	slog.SetDefault(logger)

	// Routes can be defined in a separate function
	r := router()

	logger.Info("starting server on port 3000...")

	// Router implements
	srv := &http.Server{
		Handler:      r,
		Addr:         ":3000",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		logger.Error("failed to start server", slogger.Err("error", err))
	}
}

func router() *grape.Router {
	// Create an instance of the router.
	r := grape.NewRouter()

	// Define middlewares to be used by all endpoints.
	r.Use(
		grape.RequestIDMiddleware,
		grape.RecoverMiddleware,
		grape.LoggerMiddleware,
		grape.CORSMiddleware,
	)

	// if you want the middleware to affect default handlers such as NotFound or MethodNotAllowed as well, use UseAll.
	// r.UseAll(h.LoggerMiddleware)

	r.Post("/login", authHandler)

	// Previously declared endpoints will not be impacted by these middlewares.
	r.Use(checkAuth)

	permits := r.Group("/permits")
	permits.Post("/", createPermitHandler)
	permits.Get("/{pid}", getPermitByID)

	users := permits.Group("/users") // /permits/users/
	users.Use(usersMiddleware)       // scope specific middleware
	users.Get("/", getUserPermits)

	return r
}
