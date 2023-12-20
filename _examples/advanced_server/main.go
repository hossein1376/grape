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
	// Any valid *slog.Logger will do! You can configure the destination, log format, level and more.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Any valid grape.serializer will do! (check out serializer.go for the implementation).
	serializer := Jsoniter{}

	// If you don't provide a field, default value (grape.defaultOptions) will be used.
	opts := grape.Options{Log: logger, Serialize: serializer}

	// Instantiate the grape.Server inside your struct of choice.
	h := handler{
		// Models: models.New()
		// Settings: settings.Get()
		Server: grape.New(opts),
	}

	// Routes can be defined in a separate function.
	r := h.router()

	h.Info("starting server on port 3000...")

	// You can optionally pass an instance of *http.Server to configure the running settings.
	// Note that two fields `Addr` and `Handler` are set by the Serve method. If provided, they'll be ignored.
	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err := r.Serve(":3000", srv)
	if err != nil {
		h.Error("failed to start server", "error", err)
	}
}

func (h *handler) router() *grape.Router {
	// Create an instance of the router.
	r := grape.NewRouter()

	// Define middlewares to be used by all endpoints.
	r.Use(h.LoggerMiddleware, h.RecoverMiddleware)

	// if you want the middleware to affect default handlers such as NotFound or MethodNotAllowed as well, use UseAll.
	// r.UseAll(h.LoggerMiddleware)

	r.Post("/login", h.authHandler)

	// Previously declared endpoints will not be impacted by these middlewares.
	r.Use(h.checkAuth)

	permits := r.Group("/permits")
	permits.Post("/", h.createPermitHandler)
	permits.Get("/{pid}", h.getPermitByID)

	users := permits.Group("/users") // /permits/users/
	users.Use(h.usersMiddleware)     // scope specific middleware
	users.Get("/", h.getUserPermits)

	return r
}
