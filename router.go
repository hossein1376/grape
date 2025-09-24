package grape

import (
	"net/http"
	"slices"
	"time"
)

// Router provides methods such as [Router.Get], [Router.Post], and [Router.Use]
// (among others) for routing.
type Router struct {
	scope       string
	routes      map[string]http.Handler
	middlewares []func(http.Handler) http.Handler
	root        *root
}

type root struct {
	global  []func(http.Handler) http.Handler
	routes  map[string]*Router
	handler http.Handler
}

// NewRouter will initialize and returns a new router. This function is expected
// to be called only once. Subsequent sub-path [Router] instances must be created
// via the [Router.Group] method.
func NewRouter() *Router {
	rt := &Router{
		routes: make(map[string]http.Handler),
		root: &root{
			global: make([]func(http.Handler) http.Handler, 0),
			routes: make(map[string]*Router),
		},
	}
	rt.root.routes[""] = rt
	return rt
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if r.root.handler == nil {
		r.root.handler = r.newHandler()
	}
	r.root.handler.ServeHTTP(writer, request)
}

// Group creates a new Router instance from the current one, inheriting scope
// and middlewares.
func (r *Router) Group(prefix string) *Router {
	newScope := r.scope + prefix
	existing, ok := r.root.routes[newScope]
	if ok {
		return existing
	}
	newRouter := &Router{
		scope:       newScope,
		routes:      make(map[string]http.Handler),
		middlewares: r.middlewares,
		root:        r.root,
	}

	r.root.routes[newScope] = newRouter
	return newRouter
}

// Get calls [Method] with the [http.MethodGet] method.
func (r *Router) Get(route string, handler http.HandlerFunc) {
	r.Method(http.MethodGet, route, handler)
}

// Post calls [Method] with the [http.MethodPost] method.
func (r *Router) Post(route string, handler http.HandlerFunc) {
	r.Method(http.MethodPost, route, handler)
}

// Put calls [Method] with the [http.MethodPut] method.
func (r *Router) Put(route string, handler http.HandlerFunc) {
	r.Method(http.MethodPut, route, handler)
}

// Patch calls [Method] with the [http.MethodPatch] method.
func (r *Router) Patch(route string, handler http.HandlerFunc) {
	r.Method(http.MethodPatch, route, handler)
}

// Delete calls [Method] with the [http.MethodDelete] method.
func (r *Router) Delete(route string, handler http.HandlerFunc) {
	r.Method(http.MethodDelete, route, handler)
}

// Method accepts an http method, a single route, and one handler.
func (r *Router) Method(method, route string, handler http.HandlerFunc) {
	rt := r.root.routes[r.scope]
	rt.routes[method+" "+r.scope+route] = r.withMiddlewares(handler)
}

// Use adds middlewares to the routes that are defined **after** it.
// Provided middlewares won't be applied for the previous routes, or the default
// handlers such as NotFound or MethodNotAllowed.
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	// To set correct middlewares order, and in regard to the helper
	// `withMiddlewares` method, the order will be in reverse of the
	// middlewares slice. Meaning, the first middleware to run must be
	// the last one to apply. To achieve that, the last defined
	// middleware should be the first one in the slice.
	//
	// Another approach is to reverse middlewares before applying them.

	slices.Reverse(middlewares)
	r.middlewares = slices.Concat(middlewares, r.middlewares)
}

// UseAll will add provided middleware to **all** the handlers.
// It does not matter that the route is defined before or after it. This
// included the default handlers such as NotFound or MethodNotAllowed as well.
//
// These middlewares will take precedence over all other middlewares on the same
// scope and path.
func (r *Router) UseAll(middlewares ...func(http.Handler) http.Handler) {
	// Refer to [Use] method for documentation.
	slices.Reverse(middlewares)
	r.root.global = slices.Concat(middlewares, r.root.global)
}

// Serve will start the server on the provided address. It makes no difference
// on which instance of Router this method is called from.
// A nil value for server is valid. The two fields [Addr] and [Handler] of
// [http.Server] are populated by the function itself.
func (r *Router) Serve(addr string, server *http.Server) error {
	h := r.root.handler
	if h == nil {
		h = r.newHandler()
	}
	if server == nil {
		server = &http.Server{
			// A good value between Apache's and Nginx's defaults
			// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout
			// https://httpd.apache.org/docs/2.4/mod/directive-dict.html#Default
			ReadHeaderTimeout: time.Second * 45,
		}
	}
	server.Addr = addr
	server.Handler = h
	return server.ListenAndServe()
}

func (r *Router) newHandler() http.Handler {
	mux := http.NewServeMux()
	for _, rt := range r.root.routes {
		for path, handle := range rt.routes {
			mux.Handle(path, handle)
		}
	}

	var h http.Handler = mux
	for _, middleware := range r.root.global {
		h = middleware(h)
	}
	return h
}

func (r *Router) withMiddlewares(handler http.Handler) http.Handler {
	for _, middleware := range r.middlewares {
		handler = middleware(handler)
	}
	return handler
}
