package grape

import (
	"net/http"
)

// Router embeds *http.ServeMux, adding helper functions such as Get, Post, Use and others.
type Router struct {
	local  []func(http.Handler) http.Handler
	global []func(http.Handler) http.Handler
	*http.ServeMux
}

// NewRouter will initialize a new router of type Router.
func NewRouter() *Router {
	return &Router{ServeMux: http.NewServeMux()}
}

// Get calls Method with http.MethodGet.
func (r *Router) Get(route string, handler http.HandlerFunc) {
	r.Method(http.MethodGet, route, handler)
}

// Post calls Method with http.MethodPost.
func (r *Router) Post(route string, handler http.HandlerFunc) {
	r.Method(http.MethodPost, route, handler)
}

// Put calls Method with http.MethodPut.
func (r *Router) Put(route string, handler http.HandlerFunc) {
	r.Method(http.MethodPut, route, handler)
}

// Patch calls Method with http.MethodPatch.
func (r *Router) Patch(route string, handler http.HandlerFunc) {
	r.Method(http.MethodPatch, route, handler)
}

// Delete calls Method with http.MethodDelete.
func (r *Router) Delete(route string, handler http.HandlerFunc) {
	r.Method(http.MethodDelete, route, handler)
}

// Method accept a http method, route and one handler.
func (r *Router) Method(method, route string, handler http.HandlerFunc) {
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

// Serve will start the server on the provided address.
// It takes an optional argument to modify the http.Server's configurations.
// Note that two fields Addr and Handler are populated by the function and should not be provided.
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
