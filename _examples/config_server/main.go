package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/hossein1376/grape"
)

type handler struct {
	// data/models
	// settings
	*grape.Server
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	serializer := Jsoniter{}
	opts := grape.Options{Log: logger, Serialize: serializer}

	h := handler{Server: grape.New(opts)}
	r := grape.NewRouter()

	r.UseAll(h.LoggerMiddleware, h.RecoverMiddleware)
	r.Post("/{pid}", h.permitHandler)

	h.Info("starting server on port 3000...")
	err := r.Serve(":3000")
	if err != nil {
		h.Error("failed to start server", "error", err)
		return
	}
}

func (h *handler) permitHandler(w http.ResponseWriter, r *http.Request) {
	pid := h.ParamInt(r, "pid")
	if pid == 0 {
		h.NotFoundResponse(w)
		return
	}

	type request struct {
		Name string `json:"name"`
	}

	var req request
	err := h.ReadJson(w, r, &req)
	if err != nil {
		h.BadRequestResponse(w, err)
		return
	}

	h.OkResponse(w, grape.Map{"pid": pid, "name": req.Name})
}
