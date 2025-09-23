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
	// Create new default logger for all calls to `slog` and `log` packages
	logger := slogger.NewDefault(slogger.WithLevel(slog.LevelDebug))
	// grape.Router for routing and starting the server
	r := grape.NewRouter()

	r.Use(
		grape.RequestIDMiddleware,
		grape.RecoverMiddleware,
		grape.LoggerMiddleware,
		grape.CORSMiddleware,
	)
	r.Get("/{id}", paramHandler)

	srv := &http.Server{Addr: ":3000", Handler: r}
	// Alternatively, calling r.Serve(":3000", nil) will do the same thing
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("start server failure", slogger.Err("error", err))
		return
	}
}

func paramHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slogger.Info(ctx, "Param handler!")

	// id is extracted and parsed into int
	id, err := grape.Param(r, "id", strconv.Atoi)
	if err != nil {
		err = errs.NotFound(err, errs.WithMsg("not found"))
		grape.RespondFromErr(ctx, w, err)
		return
	}

	grape.Respond(ctx, w, http.StatusOK, grape.Map{"id": id})
}
