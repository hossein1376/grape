package grape

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/hossein1376/grape/reqid"
	"github.com/hossein1376/grape/slogger"
)

type respWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *respWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery
		rw := &respWriter{ResponseWriter: w}

		ip := r.Header.Get("X-Real-Ip")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}

		defer func() {
			if raw != "" {
				path = path + "?" + raw
			}
			slogger.Info(
				r.Context(),
				"request served",
				slog.Group(
					"req",
					slog.String("client_ip", ip),
					slog.String("method", r.Method),
					slog.String("request_path", path),
				),
				slog.Group(
					"resp",
					slog.Int("status", rw.statusCode),
					slog.String("elapsed", time.Since(start).String()),
				),
			)
		}()
		next.ServeHTTP(rw, r)
	})
}

// RecoverMiddleware will recover from panics. It will display a log in error
// level, with the error message.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if msg := recover(); msg != nil {
				slog.Error("Panic recovered", slog.Any("message", msg))
				Respond(nil, w, http.StatusInternalServerError, nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := reqid.NewRequestID()
		ctx := context.WithValue(r.Context(), reqid.RequestIDKey, id)
		ctx = slogger.WithAttrs(ctx, slog.String("request_id", string(id)))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add(
			"Access-Control-Allow-Methods",
			"POST, GET, OPTIONS, PUT, DELETE",
		)
		w.Header().Add(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
		)

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
