package grape

import (
	"net/http"
)

type Router struct {
	*http.ServeMux
}

func NewRouter() *Router {
	return &Router{ServeMux: http.NewServeMux()}
}

func (r *Router) Get(route string, handler http.HandlerFunc) {
	//r.Handle(fmt.Sprintf("%s %s", http.MethodGet, route), handler)
	r.Handle(route, checkMethod(http.MethodGet, handler))
}

func (r *Router) Post(route string, handler http.HandlerFunc) {
	//r.Handle(fmt.Sprintf("%s %s", http.MethodPost, route), handler)
	r.Handle(route, checkMethod(http.MethodPost, handler))
}

func (r *Router) Put(route string, handler http.HandlerFunc) {
	//r.Handle(fmt.Sprintf("%s %s", http.MethodPut, route), handler)
	r.Handle(route, checkMethod(http.MethodPut, handler))
}

func (r *Router) Patch(route string, handler http.HandlerFunc) {
	//r.Handle(fmt.Sprintf("%s %s", http.MethodPatch, route), handler)
	r.Handle(route, checkMethod(http.MethodPatch, handler))
}

func (r *Router) Delete(route string, handler http.HandlerFunc) {
	//r.Handle(fmt.Sprintf("%s %s", http.MethodDelete, route), handler)
	http.Handle(route, checkMethod(http.MethodDelete, handler))
}

func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	for _, middleware := range middlewares {
		middleware(*r)
	}
}

func (r *Router) Serve(addr string) error {
	return http.ListenAndServe(addr, r)
}

func checkMethod(method string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
			return
		}
		next.ServeHTTP(w, r)
	})
}
