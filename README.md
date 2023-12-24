# Grape 🍇

Grape is a modern, zero-dependency HTTP library for Go.

It's a thin wrapper around the standard library, providing helper functions to facilitate faster and easier development.
Adding only a single dependency to your projects.

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

You need Go version 1.22 or higher.

```shell
go get -u github.com/hossein1376/grape@latest
```

## Usage

Main usage pattern is to embed Grape into the struct which handlers are a method to it, next to other fields like
models, settings, etc.  
In this approach, instead of changing handlers' argument to accept a specific context;
all the helper methods are available through the receiver.

The following is a simple example. For more, check out the [examples](/_examples) directory.

```go
package main

import (
	"net/http"

	"github.com/hossein1376/grape"
)

type handler struct {
	// data/models
	// settings
	grape.Server
}

func main() {
	h := handler{Server: grape.New()} // grape.Server inside a struct
	r := grape.NewRouter()            // grape.Router for routing and starting the server 

	r.Use(h.LoggerMiddleware, h.RecoverMiddleware)
	r.Get("/{id}", h.paramHandler)

	if err := r.Serve(":3000"); err != nil {
		h.Error("failed to start server", "error", err)
		return
	}
}

func (h *handler) paramHandler(w http.ResponseWriter, r *http.Request) {
	h.Info("Param handler!")

	id := h.ParamInt(r, "id")
	if id == 0 {
		h.NotFoundResponse(w)
		return
	}

	h.OkResponse(w, grape.Map{"id": id})
}

```

It is possible customize Grape for different use-cases. You can view more inside the [examples](/_examples) directory.

## Composability

Grape consists of several components independent of each other. Giving developers **opt-in choice of features**.

### `grape.Server`

Providing methods for logging, interacting with json, common HTTP responses and some other useful utilities. 
It can be embedded inside a struct, placed as a regular field, instantiate as a global variable,
or even being passed around through the context.  
An instance of it is created by running `grape.New()` and its behaviour is customizable by passing `grape.Options` 
as an argument.

### `*grape.Router`

Enable routing via methods named after HTTP verbs, with route grouping and scope-specific middlewares.
Create a new instance by running `grape.NewRouter()`.  
All routes are registered on server's startup and the rest is handled by the standard library,
causing zero runtime overhead.

### `validator` package

Presenting wide range of functions for data validation. Start a new instance with `validator.New()`,
then `Check` each part of your data with as many `Case`s it's necessary.

## Why?

Go standard library is awesome. It's fast, easy to use, and has a great API.  
With the addition of log/slog in go 1.21 and improved HTTP router in go 1.22, in most cases there are not many reasons
to look any further.
Instead of breaking compatibility with net/http, Grape aims to add commonly used functions within the arm's reach of the
handlers.

## Note

Grape is still under development. My goal is to have a stable version before release of go 1.22.  
If you're interested, make sure to read [contributing](/CONTRIBUTING.md) document first. I appreciate all inputs :)
