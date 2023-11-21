package main

import (
	"net/http"
)

// Note: you can optionally have this function as a receiver to `handler` as well
func checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check for token, cookie, etc ...
		next.ServeHTTP(w, r)
	})
}
