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
	slogger.NewDefault(slogger.WithLevel(slog.LevelDebug))
	r := grape.NewRouter()
	r.UseAll(
		grape.RequestIDMiddleware,
		grape.RecoverMiddleware,
		grape.LoggerMiddleware,
	)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		grape.Respond(r.Context(), w, http.StatusOK, "Hello, World!")
	})
	group := r.Group("")
	group.Get("/{id}", paramHandler)

	// Alternatively, you can call r.Serve(":3000", nil)
	srv := &http.Server{Addr: ":3000", Handler: r}
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("start server failure", slogger.Err("error", err))
		return
	}
}

func paramHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slogger.Debug(ctx, "Param handler!")

	id, err := grape.Param(r, "id", strconv.Atoi)
	if err != nil {
		grape.RespondFromErr(
			ctx,
			w,
			errs.BadRequest(errs.WithErr(err), errs.WithMsg("invalid id")),
		)
		return
	}

	grape.Respond(ctx, w, http.StatusOK, grape.Map{"id": id})
}
