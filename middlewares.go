package grape

import (
	"net/http"
)

func (s *Server) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Info("Received request", "method", r.Method, "uri", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.Info("Panic recovered", "message", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
