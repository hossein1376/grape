package main

import (
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/validator"
)

type handler struct {
	grape.Server
}

func main() {
	h := handler{Server: grape.New()}
	r := grape.NewRouter()

	r.UseAll(h.LoggerMiddleware, h.RecoverMiddleware)
	r.Post("/users", h.createUserHandler)

	h.Info("starting server on port 3000...")
	err := r.Serve(":3000")
	if err != nil {
		h.Error("failed to start server", "error", err)
		return
	}
}

func (h *handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Age      int    `json:"age"`
	}

	var req request
	err := h.ReadJson(w, r, &req)
	if err != nil {
		h.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	v.Check("username",
		validator.Case{Cond: validator.NotEmpty(req.Username), Msg: "must not be empty"},
		validator.Case{Cond: validator.MaxLength(req.Username, 10), Msg: "must not be over 10 characters"},
	)
	v.Check("password",
		validator.Case{Cond: validator.NotEmpty(req.Password), Msg: "must not be empty"},
		validator.Case{Cond: validator.MinLength(req.Password, 6), Msg: "must be at least 6 characters"},
	)
	v.Check("age",
		validator.Case{Cond: validator.Range(req.Age, 0, 99), Msg: "must be between 0 and 99"},
	)
	if ok := v.Valid(); !ok {
		h.Response(w, http.StatusBadRequest, v.Errors)
		return
	}

	h.Info("ping handler", "request", req)
	h.CreatedResponse(w, req)
}
