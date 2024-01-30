package main

import (
	"net/http"

	"github.com/hossein1376/grape"
)

type handler struct {
	// data/models
	// settings
	grape.Server
}

func main() {
	// Instantiate the grape.Server inside your struct of choice.
	h := handler{
		// Models: models.New()
		// Settings: settings.Get()
		Server: grape.New(),
	}
	// Create an instance of grape.Router.
	r := grape.NewRouter()

	// Define your routes
	r.Use(h.LoggerMiddleware, h.RecoverMiddleware)
	r.Get("/", h.rootHandler)

	users := r.Group("/users")
	users.Post("/{id}", h.parameterHandler)
	users.Put("/ping", h.pingHandler)

	h.Info("starting server on port 3000...")
	err := r.Serve(":3000")
	if err != nil {
		h.Error("failed to start server", "error", err)
		return
	}
}

func (h *handler) rootHandler(w http.ResponseWriter, _ *http.Request) {
	h.Debug("Get request on root")
	h.OkResponse(w, "Hello, World!")
}

func (h *handler) parameterHandler(w http.ResponseWriter, r *http.Request) {
	id, err := h.ParamInt(r, "id")
	if err != nil {
		h.NotFoundResponse(w)
		return
	}
	h.CreatedResponse(w, id)
}

func (h *handler) pingHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Data string `json:"data"`
	}

	var req request
	err := h.ReadJson(w, r, &req)
	if err != nil {
		h.Error("ping handler", "error reading request", err)
		h.BadRequestResponse(w, err)
		return
	}

	h.Info("ping handler", "request", req.Data)
	h.OkResponse(w, grape.Map{"ping": "pong", "data": req.Data})
}
