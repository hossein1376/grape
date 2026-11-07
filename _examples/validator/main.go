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

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Age      int    `json:"age"`
}

func (req createUserRequest) Validate() error {
	v := validator.New()
	v.String("username", req.Username).Required().MaxLen(10)
	v.String("password", req.Password).Required().MinLen(6)
	v.Number("age", req.Age).Range(0, 99)
	return v.Err()
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := grape.ReadJSON[createUserRequest](w, r)
	if err != nil {
		// Decoding errors are written for humans, so they're safe to show.
		grape.ExtractFromErr(ctx, w, errs.BadRequest(errs.WithErrMsg(err)))
		return
	}
	if err := req.Validate(); err != nil {
		// Responds 400 with the field errors, e.g.
		// {"message":"Bad Request","data":{"age":["must be between 0 and 99"]}}
		grape.ExtractFromErr(ctx, w, err)
		return
	}

	slogger.Info(ctx, "create user handler", slog.String("username", req.Username))
	grape.Respond(ctx, w, http.StatusNoContent, nil)
}
