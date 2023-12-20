package main

import (
	"net/http"
)

func (h *handler) checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check for token, cookie, etc ...
		h.Info("checkAuth middleware")
		next.ServeHTTP(w, r)
	})
}

func (h *handler) usersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.Info("user scope middleware")
		next.ServeHTTP(w, r)
	})
}
