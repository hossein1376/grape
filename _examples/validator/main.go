package main

import (
	"log/slog"
	"net/http"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/slogger"
	"github.com/hossein1376/grape/validator"
)

func main() {
	r := grape.NewRouter()

	r.Use(grape.LoggerMiddleware, grape.RecoverMiddleware)
	r.Post("/users", createUserHandler)

	slog.Info("starting server on port 5000...")
	err := r.Serve(":5000", nil)
	if err != nil {
		slog.Error("failed to start server", slogger.Err("error", err))
		return
	}
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Age      int    `json:"age"`
	}
	ctx := r.Context()

	var req request
	err := grape.ReadJson(w, r, &req)
	if err != nil {
		err = errs.BadRequest(err)
		grape.RespondFromErr(ctx, w, err)
		return
	}

	v := validator.New()
	v.Check("username",
		validator.Case{
			Cond: validator.NotEmpty(req.Username),
			Msg:  "must not be empty",
		},
		validator.Case{
			Cond: validator.MaxLength(req.Username, 10),
			Msg:  "must not be over 10 characters",
		},
	)
	v.Check("password",
		validator.Case{
			Cond: validator.NotEmpty(req.Password),
			Msg:  "must not be empty",
		},
		validator.Case{
			Cond: validator.MinLength(req.Password, 6),
			Msg:  "must be at least 6 characters",
		},
	)
	v.Check("age",
		validator.Case{
			Cond: validator.Range(req.Age, 0, 99),
			Msg:  "must be between 0 and 99",
		},
	)
	if ok := v.Valid(); !ok {
		grape.Respond(ctx, w, http.StatusBadRequest, err)
		return
	}

	slogger.Info(ctx, "create user handler", slog.Any("request", req))
	grape.Respond(ctx, w, http.StatusNoContent, nil)
}
