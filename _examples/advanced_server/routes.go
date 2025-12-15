package main

import (
	"fmt"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/slogger"
)

type createPermitRequest struct {
	Name string `json:"name"`
}

func (r createPermitRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func createPermitHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, err := grape.ReadJSON[createPermitRequest](w, r)
	if err != nil {
		err = errs.BadRequest(errs.WithErr(err))
		grape.ExtractFromErr(ctx, w, err)
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
