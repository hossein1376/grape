package grape

import (
	"net/http"
)

type Router struct {
	middlewares []func(http.Handler) http.Handler
	*http.ServeMux
}

func NewRouter() *Router {
	return &Router{ServeMux: http.NewServeMux()}
}

func (r *Router) Get(route string, handler http.HandlerFunc) {
	r.Handle(http.MethodGet+" "+route, r.withMiddlewares(handler))
}

func (r *Router) Post(route string, handler http.HandlerFunc) {
	r.Handle(http.MethodPost+" "+route, r.withMiddlewares(handler))
}

func (r *Router) Put(route string, handler http.HandlerFunc) {
	r.Handle(http.MethodPut+" "+route, r.withMiddlewares(handler))
}

func (r *Router) Patch(route string, handler http.HandlerFunc) {
	r.Handle(http.MethodPatch+" "+route, r.withMiddlewares(handler))
}

func (r *Router) Delete(route string, handler http.HandlerFunc) {
	r.Handle(http.MethodDelete+" "+route, r.withMiddlewares(handler))
}

func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) Serve(addr string) error {
	srv := http.Server{Addr: addr, Handler: r}
	return srv.ListenAndServe()
}

func (r *Router) withMiddlewares(handler http.Handler) http.Handler {
	h := handler
	for _, middleware := range r.middlewares {
		h = middleware(h)
	}
	return h
}
