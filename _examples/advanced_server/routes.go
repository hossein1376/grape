package main

import (
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/slogger"
)

func createPermitHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Name string `json:"name"`
	}
	ctx := r.Context()

	var req request
	err := grape.ReadJson(w, r, &req)
	if err != nil {
		slogger.Info(ctx, "reda request", slogger.Err("error", err))
		err = errs.BadRequest(errs.WithErr(err))
		return
	}

	grape.Respond(ctx, w, http.StatusCreated, grape.Map{"name": req.Name})
}

func getPermitByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slogger.Debug(ctx, "getPermitByID handler")

	pid, err := grape.Param(r, "pid", grape.ParseInt[int64]())
	if err != nil {
		slogger.Info(ctx, "invalid parameter", slogger.Err("error", err))
		err = errs.BadRequest(errs.WithMsg("invalid id"))
		return
	}

	grape.Respond(ctx, w, http.StatusCreated, grape.Map{"id": pid})
}

func getUserPermits(w http.ResponseWriter, r *http.Request) {
	grape.Respond(r.Context(), w, http.StatusOK, "users endpoint")
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	grape.Respond(r.Context(), w, http.StatusOK, "login endpoint")
}
