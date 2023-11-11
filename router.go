package grape

import (
	"fmt"
	"net/http"
)

type Router struct {
	*http.ServeMux
}

func NewRouter() *Router {
	return &Router{ServeMux: http.NewServeMux()}
}

func (r *Router) Get(route string, handler http.HandlerFunc) {
	r.Handle(fmt.Sprintf("%s %s", http.MethodGet, route), handler)
}

func (r *Router) Post(route string, handler http.HandlerFunc) {
	r.Handle(fmt.Sprintf("%s %s", http.MethodPost, route), handler)
}

func (r *Router) Put(route string, handler http.HandlerFunc) {
	r.Handle(fmt.Sprintf("%s %s", http.MethodPut, route), handler)
}

func (r *Router) Patch(route string, handler http.HandlerFunc) {
	r.Handle(fmt.Sprintf("%s %s", http.MethodPatch, route), handler)
}

func (r *Router) Delete(route string, handler http.HandlerFunc) {
	r.Handle(fmt.Sprintf("%s %s", http.MethodDelete, route), handler)
}

func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	for _, middleware := range middlewares {
		middleware(*r)
	}
}

func (r *Router) Serve(addr string) error {
	return http.ListenAndServe(addr, r)
}
