package grape

import (
	"net/http"
)

type Router struct {
	local  []func(http.Handler) http.Handler
	global []func(http.Handler) http.Handler
	*http.ServeMux
}

// NewRouter will initialize a router of type Router which embeds type ServeMux,
// with added helper functions such ad Get, Post, Use and ...
func NewRouter() *Router {
	return &Router{ServeMux: http.NewServeMux()}
}

func (r *Router) Get(route string, handler http.HandlerFunc) {
	r.handle(http.MethodGet, route, handler)
}

func (r *Router) Post(route string, handler http.HandlerFunc) {
	r.handle(http.MethodPost, route, handler)
}

func (r *Router) Put(route string, handler http.HandlerFunc) {
	r.handle(http.MethodPut, route, handler)
}

func (r *Router) Patch(route string, handler http.HandlerFunc) {
	r.handle(http.MethodPatch, route, handler)
}

func (r *Router) Delete(route string, handler http.HandlerFunc) {
	r.handle(http.MethodDelete, route, handler)
}

func (r *Router) handle(method, route string, handler http.HandlerFunc) {
	r.Handle(method+" "+route, r.withMiddlewares(handler))
}

// Use add middlewares to the routes that are defined after it.
// Note that declared middlewares won't be applied to the previous routes or the default handlers such as NotFound or MethodNotAllowed.
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.local = append(r.local, middlewares...)
}

// UseAll will add declared middleware to all the handlers.
// No matter defined before or after it; as well as the default handlers such as NotFound or MethodNotAllowed.
func (r *Router) UseAll(middlewares ...func(http.Handler) http.Handler) {
	r.global = append(r.global, middlewares...)
}

// Serve will start the server on the provided address. It takes an optional argument to modify the server's configurations.
func (r *Router) Serve(addr string, server ...*http.Server) error {
	var h http.Handler = r
	for _, middleware := range r.global {
		h = middleware(h)
	}

	srv := &http.Server{}
	if len(server) != 0 {
		srv = server[0]
	}
	srv.Addr, srv.Handler = addr, h

	return srv.ListenAndServe()
}

func (r *Router) withMiddlewares(handler http.Handler) http.Handler {
	for _, middleware := range r.local {
		handler = middleware(handler)
	}
	return handler
}
