package grape

import (
	"net/http"
)

// LoggerMiddleware logs incoming request's method and URI in `info` level
func (s *Server) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Info("Received request", "method", r.Method, "uri", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// RecoverMiddleware will recover from panics. It'll display a log in `error` level with the error message.
func (s *Server) RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.Error("Panic recovered", "message", err)
				s.InternalServerErrorResponse(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
