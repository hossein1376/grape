package grape

import (
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// Router provides methods such as [Router.Get], [Router.Post], and [Router.Use]
// (among others) for routing.
type Router struct {
	scope       string
	middlewares []func(http.Handler) http.Handler
	root        *root
}

// root is shared by a router and all of its groups.
type root struct {
	// mux holds every registered route. Routes are registered on it directly,
	// so routes added after the server has started are served as well.
	mux *http.ServeMux

	mu     sync.Mutex // guards global and rebuilding handler
	global []func(http.Handler) http.Handler
	// handler is mux wrapped with the global middlewares. It's built on the
	// first request, and reset whenever global middlewares change.
	handler atomic.Pointer[http.Handler]
}

// NewRouter will initialize and returns a new router. This function is expected
// to be called only once. Subsequent sub-path [Router] instances must be created
// via the [Router.Group] method.
func NewRouter() *Router {
	return &Router{root: &root{mux: http.NewServeMux()}}
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.root.serveHandler().ServeHTTP(writer, request)
}

func (rt *root) serveHandler() http.Handler {
	if h := rt.handler.Load(); h != nil {
		return *h
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if h := rt.handler.Load(); h != nil {
		return *h
	}
	var h http.Handler = rt.mux
	for _, middleware := range rt.global {
		h = middleware(h)
	}
	rt.handler.Store(&h)
	return h
}

// Group creates a new Router instance from the current one, inheriting scope
// and middlewares. Middlewares added to the current router afterward don't
// apply to the group, and the ones added to the group don't apply to the
// current router.
func (r *Router) Group(prefix string) *Router {
	return &Router{
		scope:       r.scope + prefix,
		middlewares: slices.Clone(r.middlewares),
		root:        r.root,
	}
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

// Method accepts an http method, a single route, and one handler. Like
// [http.ServeMux.Handle], it panics if the route is invalid or conflicts with
// an already registered one.
func (r *Router) Method(method, route string, handler http.HandlerFunc) {
	r.root.mux.Handle(method+" "+r.scope+route, r.withMiddlewares(handler))
}

// Use adds middlewares to the routes that are defined **after** it.
// Provided middlewares won't be applied for the previous routes, or the default
// handlers such as NotFound or MethodNotAllowed.
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = prependReversed(middlewares, r.middlewares)
}

// UseAll will add provided middleware to **all** the handlers.
// It does not matter that the route is defined before or after it. This
// included the default handlers such as NotFound or MethodNotAllowed as well.
//
// These middlewares will take precedence over all other middlewares on the same
// scope and path.
func (r *Router) UseAll(middlewares ...func(http.Handler) http.Handler) {
	r.root.mu.Lock()
	defer r.root.mu.Unlock()
	r.root.global = prependReversed(middlewares, r.root.global)
	r.root.handler.Store(nil)
}

// prependReversed returns a new slice of the reversed middlewares followed by
// existing ones. It doesn't modify its arguments.
//
// To set correct middlewares order, and in regard to the helper
// `withMiddlewares` method, the order will be in reverse of the middlewares
// slice. Meaning, the first middleware to run must be the last one to apply.
// To achieve that, the last defined middleware should be the first one in the
// slice.
func prependReversed(
	middlewares, existing []func(http.Handler) http.Handler,
) []func(http.Handler) http.Handler {
	reversed := slices.Clone(middlewares)
	slices.Reverse(reversed)
	return slices.Concat(reversed, existing)
}

// Serve will start the server on the provided address. It makes no difference
// on which instance of Router this method is called from.
// A nil value for server is valid. The two fields [Addr] and [Handler] of
// [http.Server] are populated by the function itself.
func (r *Router) Serve(addr string, server *http.Server) error {
	if server == nil {
		server = &http.Server{
			// A good value between Apache's and Nginx's defaults
			// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout
			// https://httpd.apache.org/docs/2.4/mod/directive-dict.html#Default
			ReadHeaderTimeout: time.Second * 45,
		}
	}
	server.Addr = addr
	server.Handler = r
	return server.ListenAndServe()
}

func (r *Router) withMiddlewares(handler http.Handler) http.Handler {
	for _, middleware := range r.middlewares {
		handler = middleware(handler)
	}
	return handler
}
