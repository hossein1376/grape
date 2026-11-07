# Grape 🍇

Grape is a modern, zero-dependency HTTP library for Go.

It's a thin wrapper around the standard library, providing helper functions to
facilitate faster and easier development, while adding only a single dependency 
to your projects.

## Features

- Zero-dependency, 100% compatible with the standard library
- Structured logging with [log/slog](https://pkg.go.dev/log/slog)
- Using new, improved [net/http](https://pkg.go.dev/net/http) router
- Group routes and scope-specific middlewares
- Read and write JSON via the [encoding/json](https://pkg.go.dev/encoding/json)
- Boosting modular and customizable architecture
- Featuring `validator` and `errs` packages for validation and graceful error
  handling

## Installation

You need Go version 1.27 or higher.

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
		// Responds 400: {"message":"invalid path parameter \"id\""}
		grape.ExtractFromErr(ctx, w, err)
		return
	}

	grape.Respond(ctx, w, http.StatusOK, grape.Map{"id": id})
}
```

More code samples can be found inside the [examples](/_examples) directory.

## Composability

Grape consists of several components independent of each other. Giving developers
**opt-in** choice of features.

### `grape.Router`

Enable routing via HTTP named methods, with route grouping and scope-specific
middlewares. Create a new instance by calling `grape.NewRouter()`.  
All routes are registered on server's startup and the rest is handled by the
standard library, causing zero runtime overhead.

### `slogger` package

Acting as an abstraction over `log/slog` package, it creates a new logger with
the provided functional options, and optionally set it as the default logger.  
It also exposes wrapper functions around `slog.LogAttrs` for efficient and less
error-prone logging.

### `errs` package

Used for effortlessly conveying error details, messages and relevant status code
between different functions and application layers . Paired with `ExtractFromErr`,
JSON responses can be automatically derived and written. Grape's own helpers,
such as `ReadJSON`, `Param` and `Query`, return `errs.Error` values for client
mistakes, so their errors can be passed to `ExtractFromErr` as is. Panics
recovered by `RecoverMiddleware`, and the router's 404 and 405 responses, are
JSON too.

### `validator` package

Type-safe validation rule chains, such as
`v.String("name", req.Name).Required().MaxLen(50)` and
`v.Number("age", req.Age).Range(0, 130)`, with default error messages. Paired
with `ExtractFromErr`, field errors are sent as a 400 response. See the
[validator README](validator/README.md).

## Why?

Go standard library is awesome. It's fast, easy to use, and has a great API.  
With the addition of `log/slog` in go 1.21 and improved HTTP router in go 1.22,
there are not many reasons to look any further.  
Instead of breaking compatibility with the `net/http`, Grape aims to add commonly
used functions within the arm's reach of developers.
