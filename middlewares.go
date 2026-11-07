package grape

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/hossein1376/grape/reqid"
	"github.com/hossein1376/grape/slogger"
)

// respWriter records the response status code for logging.
type respWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool // a final (non-informational) status has been written
	hijacked    bool
}

func (w *respWriter) Write(b []byte) (int, error) {
	w.markOK()
	return w.ResponseWriter.Write(b)
}

func (w *respWriter) WriteHeader(statusCode int) {
	// Only the first final status reaches the client; later calls are
	// superfluous. Informational 1xx responses, other than 101, may be
	// followed by the final one.
	if !w.wroteHeader {
		w.statusCode = statusCode
		w.wroteHeader = statusCode >= 200 ||
			statusCode == http.StatusSwitchingProtocols
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// markOK records the implicit 200 status net/http sends when the body is
// written or flushed before WriteHeader.
func (w *respWriter) markOK() {
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}
}

func (w *respWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err == nil {
		w.hijacked = true
	}
	return conn, rw, err
}

func (w *respWriter) FlushError() error {
	w.markOK()
	return http.NewResponseController(w.ResponseWriter).Flush()
}

func (w *respWriter) Flush() {
	_ = w.FlushError()
}

// Unwrap lets [http.ResponseController] reach the underlying writer, e.g. for
// setting read and write deadlines.
func (w *respWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// status returns the status code the client received. Hijacked connections
// that wrote no status are logged with 0.
func (w *respWriter) status(completed bool) int {
	switch {
	case w.wroteHeader || w.hijacked:
		return w.statusCode
	case !completed:
		// The handler panicked before writing anything.
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}

// LoggerMiddleware logs every request in Info level. The client_ip field is
// the connection's remote address. X-Forwarded-For and X-Real-Ip headers are
// set by clients or proxies and can't be trusted on their own, so they're
// logged separately, as forwarded_for and real_ip, when present.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery
		rw := &respWriter{ResponseWriter: w}
		completed := false

		reqAttrs := []any{
			slog.String("client_ip", r.RemoteAddr),
			slog.String("method", r.Method),
		}
		if v := r.Header.Values("X-Forwarded-For"); len(v) != 0 {
			reqAttrs = append(
				reqAttrs, slog.String("forwarded_for", strings.Join(v, ", ")),
			)
		}
		if v := r.Header.Get("X-Real-Ip"); v != "" {
			reqAttrs = append(reqAttrs, slog.String("real_ip", v))
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
					append(reqAttrs, slog.String("request_path", path))...,
				),
				slog.Group(
					"resp",
					slog.Int("status", rw.status(completed)),
					slog.String("elapsed", time.Since(start).String()),
				),
			)
		}()
		next.ServeHTTP(rw, r)
		completed = true
	})
}

// RecoverMiddleware will recover from panics. It will display a log in error
// level, with the error message. [http.ErrAbortHandler] panics are passed on,
// so net/http aborts the response as intended.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			msg := recover()
			if msg == nil {
				return
			}
			if err, ok := msg.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(msg)
			}
			slogger.Error(
				r.Context(),
				"Panic recovered",
				slog.Any("message", msg),
				slog.String("stack", string(debug.Stack())),
			)
			Respond(r.Context(), w, http.StatusInternalServerError, nil)
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

// CORSMiddleware allows cross-origin requests from any origin, and answers
// preflight requests with 204 No Content.
//
// Register it with [Router.UseAll]. Preflight requests use the OPTIONS method,
// which doesn't match routes registered with [Router.Get], [Router.Post], and
// the like, so with [Router.Use] they're answered 405 before reaching it.
//
// Since it allows any origin, it doesn't allow credentials: browsers reject
// credentialed requests when Access-Control-Allow-Origin is "*".
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set(
			"Access-Control-Allow-Methods",
			"POST, GET, OPTIONS, PUT, PATCH, DELETE",
		)
		h.Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
		)

		if r.Method == http.MethodOptions &&
			r.Header.Get("Access-Control-Request-Method") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
