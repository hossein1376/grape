package grape

import (
	"net/http"
	"slices"
	"strings"
	"time"
)

// Router provides methods such as Get, Post and Use (among others) for routing.
type Router struct {
	scope       string
	routes      map[string]http.Handler
	middlewares []func(http.Handler) http.Handler
	base        *base
}

type base struct {
	global []func(http.Handler) http.Handler
	routes map[string]*Router
}

// NewRouter will initialize a new router of type Router.
// This function is expected to be called only once. Subsequent sub-path Router instance are created by the Group method.
func NewRouter() *Router {
	rt := &Router{
		routes: make(map[string]http.Handler),
		base: &base{
			global: make([]func(http.Handler) http.Handler, 0),
			routes: make(map[string]*Router),
		},
	}
	rt.base.routes[""] = rt
	return rt
}

// Group creates a new Router instance from the current one, inheriting scope and middlewares.
func (r *Router) Group(prefix string) *Router {
	newScope := r.scope + prefix
	newRouter := &Router{
		scope:       newScope,
		routes:      make(map[string]http.Handler),
		middlewares: r.middlewares,
		base:        r.base,
	}

	r.base.routes[newScope] = newRouter
	return newRouter
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
// As a bonus, if the provided route has a trailing slash, it disables net/http's default catch-all behaviour.
func (r *Router) Method(method, route string, handler http.HandlerFunc) {
	if strings.HasSuffix(route, "/") {
		route += "{$}"
	}

	rt := r.base.routes[r.scope]
	rt.routes[method+" "+r.scope+route] = r.withMiddlewares(handler)
}

// Use add middlewares to the routes that are defined after it.
// Note that declared middlewares won't be applied to the previous routes or the default handlers such as NotFound or MethodNotAllowed.
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	// To set correct middlewares order, and in regard to `withMiddlewares`
	// method, the order will be in reverse of the middlewares slice.
	// Meaning, the first middleware to run must be the last one to apply.
	// To achieve that, the last defined middleware should be the first one
	// in the slice.
	//
	// Another approach was to reverse middlewares before applying them.

	slices.Reverse(middlewares)
	r.middlewares = slices.Concat(middlewares, r.middlewares)
}

// UseAll will add declared middleware to all the handlers.
// No matter defined before or after it; as well as the default handlers such as NotFound or MethodNotAllowed.
//
// These middlewares will take precedence over all other middlewares on the same scope and path.
func (r *Router) UseAll(middlewares ...func(http.Handler) http.Handler) {
	// Refer to `Use` method for documentation.
	slices.Reverse(middlewares)
	r.base.global = slices.Concat(middlewares, r.base.global)
}

// Serve will start the server on the provided address. It makes no difference on which instance of Router this method
// is called from.
//
// It takes an optional argument to modify the http.Server's configurations.
// Note that two fields Addr and Handler are populated by the function and will be ignored if provided.
func (r *Router) Serve(addr string, server ...*http.Server) error {
	handler := &http.ServeMux{}
	for _, rt := range r.base.routes {
		for path, handle := range rt.routes {
			handler.Handle(path, handle)
		}
	}

	var h http.Handler = handler
	for _, middleware := range r.base.global {
		h = middleware(h)
	}

	srv := &http.Server{
		// A good value between Apache's and Nginx's defaults
		// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout
		// https://httpd.apache.org/docs/2.4/mod/directive-dict.html#Default
		ReadHeaderTimeout: time.Second * 45,
	}
	if len(server) != 0 {
		srv = server[0]
	}
	srv.Addr, srv.Handler = addr, h

	return srv.ListenAndServe()
}

func (r *Router) withMiddlewares(handler http.Handler) http.Handler {
	for _, middleware := range r.middlewares {
		handler = middleware(handler)
	}
	return handler
}
