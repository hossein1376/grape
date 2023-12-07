# Grape 🍇

Grape is a modern, zero-dependency HTTP library for Go.

It's a thin wrapper around the standard library, providing helper functions to facilitate faster and easier development.
Adding only a single dependency to your projects.

## Features

- Zero-dependency, 100% compatible with the standard library
- Structured logging with [log/slog](https://pkg.go.dev/log/slog)
- Using new, improved [net/http](https://pkg.go.dev/net/http) router
- Read and write json via the [encoding/json](https://pkg.go.dev/encoding/json)
- Customizable with different `slog` configurations; or to use 3rd-party serializer packages
- Helper functions for commonly used HTTP status code responses
- Featuring a built-in `validator` package for data validation

## Installation

You need Go version 1.22 or higher.

```shell
go get -u github.com/hossein1376/grape@latest
```

## Usage

Main philosophy is to embed Grape into the struct which handlers are a method to it, next to other fields like models,
settings, etc.  
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

	r.UseAll(h.LoggerMiddleware, h.RecoverMiddleware)
	r.Get("/{id}", h.pingHandler)

	if err := r.Serve(":3000"); err != nil {
		h.Error("failed to start server", "error", err)
		return
	}
}

func (h *handler) pingHandler(w http.ResponseWriter, r *http.Request) {
	h.Info("Ping handler!")

	id := h.ParamInt(r, "id")
	if id == 0 {
		h.NotFoundResponse(w)
		return
	}

	h.OkResponse(w, grape.Map{"id": id})
}

```

It is possible customize Grape for different use-cases. You can see more inside the [examples](/_examples) directory.

## Why?

Go standard library is awesome. It's fast, easy to use, and has a great API.  
With the addition of log/slog in go 1.21 and improved HTTP router in go 1.22, there are not many reasons to look any
further in most cases.
Instead of breaking compatibility with net/http, Grape aims to add commonly used functions within the arm's reach of the
handlers.

## Note

Grape is still under development. My goal is to have a stable version before release of go 1.22.  
If you're interested, make sure to read [contributing](/CONTRIBUTING.md) document first. I appreciate all inputs :)

### TODO

- [ ] Include grouping and scope middlewares
- [ ] Extend comments and documentation
- [ ] Add tests 
