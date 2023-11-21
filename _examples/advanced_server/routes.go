package main

import (
	"net/http"

	"github.com/hossein1376/grape"
)

func (h *handler) createPermitHandler(w http.ResponseWriter, r *http.Request) {
	pid := h.ParamInt64(r, "pid")
	if pid == 0 {
		h.Debug("permit_handler", "error", "invalid parameter")
		h.NotFoundResponse(w)
		return
	}

	type request struct {
		Name string `json:"name"`
	}

	var req request
	err := h.ReadJson(w, r, &req)
	if err != nil {
		h.Info("permit_handler", "error", err)
		h.BadRequestResponse(w, err)
		return
	}

	h.CreatedResponse(w, grape.Map{"pid": pid, "name": req.Name})
}

func (h *handler) authHandler(http.ResponseWriter, *http.Request) {}

func (h *handler) getAllPermitsHandler(http.ResponseWriter, *http.Request) {}

func (h *handler) createNewUser(http.ResponseWriter, *http.Request) {}

func (h *handler) getUserByID(http.ResponseWriter, *http.Request) {}

func (h *handler) getUserPermits(http.ResponseWriter, *http.Request) {}
