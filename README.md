# Grape 🍇

Grape is a modern, zero-dependency HTTP library for Go.

It's a thin wrapper around the standard library, providing helper functions to
facilitate faster and easier development. Adding only a single dependency to
your projects.

## Features

- Zero-dependency, 100% compatible with the standard library
- Structured logging with [log/slog](https://pkg.go.dev/log/slog)
- Using new, improved [net/http](https://pkg.go.dev/net/http) router
- Group routes and scope-specific middlewares
- Read and write json via the [encoding/json](https://pkg.go.dev/encoding/json)
- Highly customizable; bring your own logger and serializer!
- Helper functions for commonly used HTTP status code responses
- Featuring a built-in `validator` package for data validation

## Installation

You need Go version 1.25 or higher.

```shell
go get -u github.com/hossein1376/grape@latest
```

## Usage

The following is a simple example. For more, check out the [examples](/_examples)
directory.

```go
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

```

It is possible customize Grape for different use-cases. You can view more inside
the [examples](/_examples) directory.

## Composability

Grape consists of several components independent of each other. Giving developers
**opt-in choice of features**.

### `*grape.Router`

Enable routing via methods named after HTTP verbs, with route grouping and
scope-specific middlewares. Create a new instance by running `grape.NewRouter()`.  
All routes are registered on server's startup and the rest is handled by the
standard library, causing zero runtime overhead.

### `slogger` package

TODO

### `errs` package

TODO

### `validator` package

Presenting wide range of functions for data validation. Start a new instance with
`validator.New()`, then `Check` each part of your data with as many `Case`s 
it's necessary.

## Why?

Go standard library is awesome. It's fast, easy to use, and has a great API.  
With the addition of log/slog in go 1.21 and improved HTTP router in go 1.22, in
most cases there are not many reasons to look any further.  
Instead of breaking compatibility with net/http, Grape aims to add commonly used
functions within the arm's reach.
