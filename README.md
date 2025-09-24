# Grape 🍇

Grape is a modern, zero-dependency HTTP library for Go.

It's a thin wrapper around the standard library, providing helper functions to
facilitate faster and easier development. Adding only a single dependency to
your projects.

## Features

- Zero-dependency, 100% compatible with the standard library
- Structured logging with [log/slog](https://pkg.go.dev/log/slog) using the`slogger` package
- Using new, improved [net/http](https://pkg.go.dev/net/http) router
- Group routes and scope-specific middlewares
- Read and write json via the [encoding/json](https://pkg.go.dev/encoding/json)
- Boosting modular and customizable architecture
- Featuring built-in `validator` and `errs` packages for validation and graceful
 error handling

## Installation

You need Go version 1.25 or higher.

```shell
go get -u github.com/hossein1376/grape@latest
```

## Usage

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
			ctx, w, errs.BadRequest(err, errs.WithMsg("invalid id")),
		)
		return
	}

	grape.Respond(ctx, w, http.StatusOK, grape.Map{"id": id})
}

```

It is possible customize and extend Grape for different use-cases. You can view
more code samples inside the [examples](/_examples) directory.

## Composability

Grape consists of several components independent of each other. Giving developers
**opt-in choice of features**.

### `*grape.Router`

Enable routing via HTTP named methods, with route grouping and scope-specific
middlewares. Create a new instance by calling `grape.NewRouter()`.  
All routes are registered on server's startup and the rest is handled by the
standard library, causing zero runtime overhead.

### `slogger` package

Acting as an abstraction over `log/slog` package, it creates a new logger with 
the provided functional options, and optionally set it as the default logger.  
It exposes wrapper functions around `slog.LogAttrs` for efficient and safe
logging.

### `errs` package

Used for effortlessly conveying error details, messages and relevant status code
among different layers and functions. Paired with `grape.RespondFromErr`,
responses are automatically derived and written.

### `validator` package

Presenting wide range of functions for data validation, start a new instance with
`validator.New()` and then `Check` each part of your data with as many `Case`s 
it's necessary.

## Why?

Go's standard library is awesome. It's fast, easy to use, and has a great API.  
With the addition of log/slog in go 1.21 and improved HTTP router in go 1.22, in
most cases there are not many reasons to look any further.  
Instead of breaking compatibility with the `net/http`, Grape aims to add commonly
used functions within the arm's reach of developer.
