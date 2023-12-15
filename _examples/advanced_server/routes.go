package main

import (
	"net/http"

	"github.com/hossein1376/grape"
)

func (h *handler) createPermitHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Name string `json:"name"`
	}

	var req request
	err := h.ReadJson(w, r, &req)
	if err != nil {
		h.Info("create_permit_handler", "error", err)
		h.BadRequestResponse(w, err)
		return
	}

	h.CreatedResponse(w, grape.Map{"name": req.Name})
}

func (h *handler) getPermitByID(w http.ResponseWriter, r *http.Request) {
	pid := h.ParamInt64(r, "pid")
	if pid == 0 {
		h.Info("get_permit_handler", "error", "invalid parameter")
		h.NotFoundResponse(w)
		return
	}

	h.CreatedResponse(w, grape.Map{"id": pid})
}

func (h *handler) getUserPermits(w http.ResponseWriter, _ *http.Request) {
	h.Response(w, http.StatusOK, "users endpoint") // equivalent of using h.OkResponse()
}

func (h *handler) authHandler(w http.ResponseWriter, _ *http.Request) {
	h.OkResponse(w, "login endpoint")
}
