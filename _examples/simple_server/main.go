package main

import (
	"net/http"

	"github.com/hossein1376/grape"
)

type handler struct {
	// data/models
	// settings
	*grape.Server
}

func main() {
	h := handler{Server: grape.New()}
	r := grape.NewRouter()

	r.UseAll(h.LoggerMiddleware, h.RecoverMiddleware)
	r.Get("/", h.rootHandler)
	r.Post("/ping", h.pingHandler)
	r.Put("/{id}", h.parameterHandler)

	h.Info("starting server on port 3000...")
	err := r.Serve(":3000")
	if err != nil {
		h.Error("failed to start server", "error", err)
		return
	}
}

func (h *handler) rootHandler(w http.ResponseWriter, _ *http.Request) {
	h.Warn("Get request on root")
	h.OkResponse(w, "Hello, World!")
}

func (h *handler) pingHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Ping string `json:"ping"`
	}

	var req request
	err := h.ReadJson(w, r, &req)
	if err != nil {
		h.BadRequestResponse(w, err)
		return
	}

	h.Debug("ping handler", "request", req.Ping)
	h.NoContentResponse(w)
}

func (h *handler) parameterHandler(w http.ResponseWriter, r *http.Request) {
	id := h.ParamInt(r, "id")
	if id == 0 {
		h.NotFoundResponse(w)
		return
	}
	h.CreatedResponse(w, grape.Map{"id": id})
}
