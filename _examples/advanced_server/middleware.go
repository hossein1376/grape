package main

import (
	"log/slog"
	"net/http"
)

func checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check for token, cookie, etc ...
		slog.Info("checkAuth middleware")
		next.ServeHTTP(w, r)
	})
}

func usersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("user scope middleware")
		next.ServeHTTP(w, r)
	})
}
