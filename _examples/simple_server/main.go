package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/hossein1376/grape"
	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/slogger"
)

func main() {
	// Create a new logger and set it as the default logger
	_ = slogger.NewDefault()
	// Create an instance of grape.Router.
	r := grape.NewRouter()

	// Define your routes
	r.Use(grape.RequestIDMiddleware, grape.RecoverMiddleware, grape.LoggerMiddleware)
	r.Get("/", rootHandler)

	users := r.Group("/users")
	users.Post("/{id}", parameterHandler)
	users.Put("/ping", pingHandler)

	slog.Info("starting server on port 3000...")
	err := r.Serve(":3000", nil)
	if err != nil {
		slog.Error("failed to start server", "error", err)
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slogger.Debug(ctx, "Get request on root")
	grape.Respond(ctx, w, http.StatusOK, "Hello, world!")
}

func parameterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := grape.Param(r, "id", strconv.Atoi)
	if err != nil {
		err = errs.NotFound(errs.WithErr(err), errs.WithMsg("user id not found"))
		grape.RespondFromErr(ctx, w, err)
		return
	}
	grape.Respond(ctx, w, http.StatusOK, id)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Data string `json:"data"`
	}
	ctx := r.Context()

	req, err := grape.ReadJSON[request](w, r)
	if err != nil {
		slogger.Error(ctx, "reading request", slogger.Err("error", err))
		err = errs.BadRequest(errs.WithMsg("invalid request"))
		grape.RespondFromErr(ctx, w, err)
		return
	}

	slogger.Info(ctx, "ping handler", slog.Any("request", req.Data))
	grape.Respond(
		ctx, w, http.StatusOK, grape.Map{"ping": "pong", "data": req.Data},
	)
}
