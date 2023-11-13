package grape

import (
	"net/http"
)

type Router struct {
	local  []func(http.Handler) http.Handler
	global []func(http.Handler) http.Handler
	*http.ServeMux
}

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

func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.local = append(r.local, middlewares...)
}

func (r *Router) UseAll(middlewares ...func(http.Handler) http.Handler) {
	r.global = append(r.global, middlewares...)
}

func (r *Router) Serve(addr string) error {
	var h http.Handler = r
	for _, middleware := range r.global {
		h = middleware(h)
	}
	srv := http.Server{Addr: addr, Handler: h}
	return srv.ListenAndServe()
}

func (r *Router) withMiddlewares(handler http.Handler) http.Handler {
	for _, middleware := range r.local {
		handler = middleware(handler)
	}
	return handler
}
