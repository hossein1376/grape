package grape

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hossein1376/grape/slogger"
)

// captureLogs points the default logger at a buffer for the test's duration.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slogger.New(slogger.WithDestination(&buf)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

type servedLog struct {
	Msg string `json:"msg"`
	Req struct {
		ClientIP     string `json:"client_ip"`
		ForwardedFor string `json:"forwarded_for"`
		RealIP       string `json:"real_ip"`
	} `json:"req"`
	Resp struct {
		Status int `json:"status"`
	} `json:"resp"`
}

// servedLogLine returns the "request served" log line.
func servedLogLine(t *testing.T, buf *bytes.Buffer) servedLog {
	t.Helper()
	for line := range bytes.Lines(buf.Bytes()) {
		var l servedLog
		if err := json.Unmarshal(line, &l); err != nil {
			t.Fatalf("unmarshal log line %q: %v", line, err)
		}
		if l.Msg == "request served" {
			return l
		}
	}
	t.Fatalf("no request served log in %q", buf.String())
	return servedLog{}
}

func TestLoggerMiddleware_Status(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    int
	}{
		{"nothing written", func(http.ResponseWriter, *http.Request) {}, http.StatusOK},
		{"body only", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("x")) }, http.StatusOK},
		{"flush only", func(w http.ResponseWriter, _ *http.Request) { w.(http.Flusher).Flush() }, http.StatusOK},
		{"superfluous WriteHeader", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusTeapot)
		}, http.StatusCreated},
		{"informational then final", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusEarlyHints)
			w.Write([]byte("x"))
		}, http.StatusOK},
		{"panic before writing", func(http.ResponseWriter, *http.Request) { panic("boom") }, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := captureLogs(t)
			h := RecoverMiddleware(LoggerMiddleware(tt.handler))
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
			if got := servedLogLine(t, buf).Resp.Status; got != tt.want {
				t.Fatalf("logged status %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoggerMiddleware_ClientIPIsRemoteAddr(t *testing.T) {
	buf := captureLogs(t)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	req.Header.Set("X-Real-Ip", "203.0.113.8")
	LoggerMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), req)

	l := servedLogLine(t, buf).Req
	if l.ClientIP != "192.0.2.1:1234" || l.ForwardedFor != "203.0.113.9" ||
		l.RealIP != "203.0.113.8" {
		t.Fatalf("unexpected request log fields: %+v", l)
	}
}

func TestLoggerMiddleware_ResponseControllerReachesConn(t *testing.T) {
	captureLogs(t)
	errc := make(chan error, 1)
	srv := httptest.NewServer(LoggerMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			errc <- http.NewResponseController(w).
				SetWriteDeadline(time.Now().Add(time.Second))
		},
	)))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if err := <-errc; err != nil {
		t.Fatalf("SetWriteDeadline through LoggerMiddleware: %v", err)
	}
}

func TestRecoverMiddleware_RepanicsErrAbortHandler(t *testing.T) {
	captureLogs(t)
	h := RecoverMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	defer func() {
		if err, _ := recover().(error); !errors.Is(err, http.ErrAbortHandler) {
			t.Fatalf("expected ErrAbortHandler to propagate, got %v", err)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}

func TestCORSMiddleware(t *testing.T) {
	r := NewRouter()
	r.UseAll(CORSMiddleware)
	r.Post("/x", func(w http.ResponseWriter, _ *http.Request) {})

	preflight := httptest.NewRequest(http.MethodOptions, "/x", nil)
	preflight.Header.Set("Origin", "https://example.com")
	preflight.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, preflight)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight: expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("preflight: unexpected Allow-Origin %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Allow-Credentials must not be sent with a wildcard origin, got %q", got)
	}

	// A plain OPTIONS request isn't a preflight and goes to the router.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/x", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("plain OPTIONS: expected 405 from the router, got %d", rec.Code)
	}
}
