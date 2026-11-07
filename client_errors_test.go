package grape

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/reqid"
	"github.com/hossein1376/grape/validator"
)

// jsonBody decodes a grape JSON response, failing if it isn't one.
func jsonBody(t *testing.T, rec *httptest.ResponseRecorder) Response {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected a JSON response, got content type %q: %q", ct, rec.Body.String())
	}
	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body %q: %v", rec.Body.String(), err)
	}
	return body
}

// assertResponse checks what ExtractFromErr sends for err.
func assertResponse(t *testing.T, err error, status int, msg string) Response {
	t.Helper()
	rec := httptest.NewRecorder()
	ExtractFromErr(context.Background(), rec, err)
	if rec.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, rec.Code, rec.Body.String())
	}
	body := jsonBody(t, rec)
	if body.Message != msg {
		t.Fatalf("expected message %q, got %q", msg, body.Message)
	}
	return body
}

func jsonRequest(contentType, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req
}

func TestReadJSON_ClientErrors(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		status      int
		msg         string
	}{
		{"missing content type", "", `{}`, 415, "content type must be application/json"},
		{"wrong content type", "text/plain", `{}`, 415, "content type must be application/json"},
		{"look-alike content type", "application/jsonx", `{}`, 415, "content type must be application/json"},
		{"too large", "application/json", `{"name":"` + strings.Repeat("x", 64) + `"}`, 413, "body must not be larger than 16 bytes"},
		{"empty", "application/json", ``, 400, "body must not be empty"},
		{"badly-formed", "application/json", `{"name":`, 400, "body contains badly-formed JSON"},
		{"syntax error", "application/json", `{"name" "x"}`, 400, "body contains badly-formed JSON (at character 9)"},
		{"wrong type", "application/json", `{"age":"x"}`, 400, `body contains incorrect JSON type for field "age"`},
		{"unknown field", "application/json", `{"nope":1}`, 400, `body contains unknown key "nope"`},
		{"two values", "application/json", `{}{}`, 400, "body must only contain a single JSON value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadJSON[personPayload](
				httptest.NewRecorder(),
				jsonRequest(tt.contentType, tt.body),
				WithMaxBodySize(16),
			)
			assertResponse(t, err, tt.status, tt.msg)
		})
	}
}

func TestReadJSON_KeepsUnderlyingError(t *testing.T) {
	_, err := ReadJSON[personPayload](
		httptest.NewRecorder(), jsonRequest("application/json", `{"name" "x"}`),
	)
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("expected a *json.SyntaxError in the chain, got %v", err)
	}
}

func TestReadJSON_AcceptsJSONBasedContentTypes(t *testing.T) {
	for _, ct := range []string{
		"application/json",
		"Application/JSON; charset=utf-8",
		"application/merge-patch+json",
	} {
		got, err := ReadJSON[personPayload](
			httptest.NewRecorder(), jsonRequest(ct, `{"name":"a"}`),
		)
		if err != nil || got.Name != "a" {
			t.Fatalf("%q: got %v, %v", ct, got, err)
		}
	}
}

type conflictValidated struct{}

func (conflictValidated) Validate() error { return errs.Conflict(errs.WithMsg("taken")) }

type fieldsValidated struct{ Name string }

func (f fieldsValidated) Validate() error {
	v := validator.New()
	v.String("name", f.Name).Required()
	return v.Err()
}

func TestReadValidateJSON_ClientErrors(t *testing.T) {
	_, err := ReadValidateJSON[payloadWithValidate](
		httptest.NewRecorder(), jsonRequest("application/json", `{"name":""}`),
	)
	assertResponse(t, err, 400, "name required")

	_, err = ReadValidateJSON[*ptrValidated](
		httptest.NewRecorder(), jsonRequest("application/json", `null`),
	)
	assertResponse(t, err, 400, "body must not be null")

	_, err = ReadValidateJSON[conflictValidated](
		httptest.NewRecorder(), jsonRequest("application/json", `{}`),
	)
	assertResponse(t, err, 409, "taken")

	_, err = ReadValidateJSON[fieldsValidated](
		httptest.NewRecorder(), jsonRequest("application/json", `{"Name":""}`),
	)
	body := assertResponse(t, err, 400, "Bad Request")
	if fields, _ := body.Data.(map[string]any); fields["name"] == nil {
		t.Fatalf("expected field errors in data, got %v", body.Data)
	}
}

// grape.Param and grape.Query errors, sent through a router as a handler
// would.
func TestParamAndQuery_ClientErrors(t *testing.T) {
	r := NewRouter()
	handle := func(err error, w http.ResponseWriter, req *http.Request) {
		if err != nil {
			ExtractFromErr(req.Context(), w, err)
		}
	}
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
		_, err := Param(req, "id", strconv.Atoi)
		handle(err, w, req)
	})
	r.Get("/files/{path...}", func(w http.ResponseWriter, req *http.Request) {
		_, err := Param(req, "path", func(s string) (string, error) { return s, nil })
		handle(err, w, req)
	})
	r.Get("/typo/{id}", func(w http.ResponseWriter, req *http.Request) {
		_, err := Param(req, "ID", strconv.Atoi) // no such wildcard
		handle(err, w, req)
	})
	r.Get("/items/{sku}", func(w http.ResponseWriter, req *http.Request) {
		_, err := Param(req, "sku", func(string) (int, error) {
			return 0, errs.NotFound(errs.WithMsg("no such item"))
		})
		handle(err, w, req)
	})
	r.Get("/list", func(w http.ResponseWriter, req *http.Request) {
		_, err := Query(req.URL.Query(), "page", ParseInt[int]())
		handle(err, w, req)
	})

	tests := []struct {
		path   string
		status int
		msg    string
	}{
		{"/users/abc", 400, `invalid path parameter "id"`},
		{"/files/", 400, `path parameter "path" is required`},
		{"/typo/1", 500, "Internal Server Error"},
		{"/items/x", 404, "no such item"},
		{"/list", 400, `query parameter "page" is required`},
		{"/list?page=", 400, `query parameter "page" is required`},
		{"/list?page=abc", 400, `invalid query parameter "page"`},
		{"/list?page=2", 200, ""},
	}
	for _, tt := range tests {
		rec := serve(r, http.MethodGet, tt.path)
		if rec.Code != tt.status {
			t.Fatalf("%s: expected %d, got %d: %s", tt.path, tt.status, rec.Code, rec.Body.String())
		}
		if tt.msg != "" {
			if got := jsonBody(t, rec).Message; got != tt.msg {
				t.Fatalf("%s: expected message %q, got %q", tt.path, tt.msg, got)
			}
		}
	}
}

func TestQuery_KeepsSentinelAndParseError(t *testing.T) {
	q := map[string][]string{"n": {"abc"}}
	if _, err := Query(q, "n", ParseInt[int]()); !errors.Is(err, strconv.ErrSyntax) {
		t.Fatalf("expected strconv.ErrSyntax in the chain, got %v", err)
	}
	if _, err := Query(q, "m", ParseInt[int]()); !errors.Is(err, ErrMissingQuery) {
		t.Fatalf("expected ErrMissingQuery in the chain, got %v", err)
	}
}

func TestRouter_JSONNotFoundAndMethodNotAllowed(t *testing.T) {
	r := NewRouter()
	r.Get("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {})
	r.Get("/gone", func(w http.ResponseWriter, req *http.Request) { http.NotFound(w, req) })
	r.Get("/dir/", func(w http.ResponseWriter, _ *http.Request) {})

	rec := serve(r, http.MethodGet, "/nope")
	if rec.Code != 404 || jsonBody(t, rec).Message != "Not Found" {
		t.Fatalf("404: got %d %q", rec.Code, rec.Body.String())
	}

	rec = serve(r, http.MethodPost, "/users/1")
	if rec.Code != 405 || jsonBody(t, rec).Message != "Method Not Allowed" {
		t.Fatalf("405: got %d %q", rec.Code, rec.Body.String())
	}
	if allow := rec.Header().Get("Allow"); allow != "GET, HEAD" {
		t.Fatalf("405: expected Allow header to be kept, got %q", allow)
	}

	// A handler's own 404 is left alone.
	rec = serve(r, http.MethodGet, "/gone")
	if rec.Code != 404 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "text/plain") {
		t.Fatalf("handler 404: got %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}

	// So are mux's redirects.
	rec = serve(r, http.MethodGet, "/dir")
	if rec.Code != http.StatusTemporaryRedirect || rec.Header().Get("Location") != "/dir/" {
		t.Fatalf("redirect: got %d %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestRecoverMiddleware_JSONBodyWithRequestID(t *testing.T) {
	captureLogs(t)
	var id string
	h := RequestIDMiddleware(RecoverMiddleware(http.HandlerFunc(
		func(_ http.ResponseWriter, r *http.Request) {
			id, _ = reqid.RequestID(r.Context())
			panic("boom")
		},
	)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != 500 {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	body := jsonBody(t, rec)
	if body.Message != "Internal Server Error" || body.Data != id || id == "" {
		t.Fatalf("unexpected body %+v, request ID %q", body, id)
	}
}

// Once the response has started, a 500 can't be sent, so the connection is
// aborted instead.
func TestRecoverMiddleware_AbortsStartedResponse(t *testing.T) {
	captureLogs(t)
	h := RecoverMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("partial"))
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	defer func() {
		if err, _ := recover().(error); !errors.Is(err, http.ErrAbortHandler) {
			t.Fatalf("expected ErrAbortHandler, got %v", err)
		}
		if rec.Code != 200 || rec.Body.String() != "partial" {
			t.Fatalf("expected the response to be left as is, got %d %q", rec.Code, rec.Body.String())
		}
	}()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
}
