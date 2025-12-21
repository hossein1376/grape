package grape

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- WriteJSON tests ---

func TestWriteJSON_NilData_SetsStatusAndHeaders(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(context.TODO(), rec, WithStatus(http.StatusAccepted))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d got %d", http.StatusAccepted, rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf(
			"expected empty body when data is nil, got %q", rec.Body.String(),
		)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", ct)
	}
}

func TestWriteJSON_WithData_Marshals(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(
		context.TODO(), rec, WithStatus(http.StatusOK), WithData(Map{"x": "y"}),
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", ct)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["x"] != "y" {
		t.Fatalf("unexpected body: %v", got)
	}
}

func TestWriteJSON_MarshalErrorReturns500(t *testing.T) {
	rec := httptest.NewRecorder()
	// channels cannot be marshaled to JSON
	WriteJSON(
		context.TODO(), rec, WithStatus(http.StatusOK), WithData(make(chan int)),
	)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when marshal fails, got %d", rec.Code)
	}
}

func TestWriteJSON_WithHeaders_Merged(t *testing.T) {
	rec := httptest.NewRecorder()
	hdrs := make(http.Header)
	hdrs.Set("X-Custom", "v")
	WriteJSON(
		context.TODO(),
		rec,
		WithStatus(http.StatusTeapot),
		WithData(Map{"a": 1}),
		WithHeaders(hdrs),
	)

	if rec.Header().Get("X-Custom") != "v" {
		t.Fatalf(
			"expected custom header merged, got %q", rec.Header().Get("X-Custom"),
		)
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected status %d got %d", http.StatusTeapot, rec.Code)
	}
}

// --- ReadJSON tests ---

type personPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type payloadWithValidate struct {
	Name string `json:"name"`
}

func (p payloadWithValidate) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("name required")
	}
	return nil
}

func TestReadJSON_Success(t *testing.T) {
	jsonBody := `{"name":"alice","age":30}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost, "/", bytes.NewBufferString(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")

	got, err := ReadJSON[personPayload](rec, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "alice" || got.Age != 30 {
		t.Fatalf("unexpected decoded value: %#v", got)
	}
}

func TestReadJSON_ValidateFails(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost, "/", bytes.NewBufferString(`{"name": ""}`),
	)
	req.Header.Set("Content-Type", "application/json")

	_, err := ReadJSON[payloadWithValidate](rec, req)
	if err == nil || !strings.Contains(err.Error(), "name required") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestReadJSON_ContentTypeMismatch(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "text/plain")

	_, err := ReadJSON[personPayload](rec, req)
	if err == nil || err.Error() != "content type is not application/json" {
		t.Fatalf("expected content type error, got %v", err)
	}
}

func TestReadJSON_EmptyBody(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")

	_, err := ReadJSON[personPayload](rec, req)
	if err == nil || !strings.Contains(err.Error(), "body must not be empty") {
		t.Fatalf("expected empty body error, got %v", err)
	}
}

func TestReadJSON_BadlyFormedJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost, "/", bytes.NewBufferString(`{"x":`),
	)
	req.Header.Set("Content-Type", "application/json")

	_, err := ReadJSON[personPayload](rec, req)
	if err == nil || !strings.Contains(err.Error(), "badly-formed") {
		t.Fatalf("expected badly-formed JSON error, got %v", err)
	}
}

func TestReadJSON_UnknownField(t *testing.T) {
	rec := httptest.NewRecorder()
	// 'extra' is not part of personPayload
	req := httptest.NewRequest(
		http.MethodPost, "/", bytes.NewBufferString(`{"name":"bob","extra":1}`),
	)
	req.Header.Set("Content-Type", "application/json")

	_, err := ReadJSON[personPayload](rec, req)
	if err == nil || !strings.Contains(err.Error(), "body contains unknown key") {
		t.Fatalf("expected unknown key error, got %v", err)
	}
}

func TestReadJSON_TooLargeBody(t *testing.T) {
	rec := httptest.NewRecorder()
	large := strings.Repeat("x", 1024)
	req := httptest.NewRequest(
		http.MethodPost, "/", bytes.NewBufferString(`{"name":"`+large+`"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	// provide a correctly-typed option for the generic ReadJSON
	opt := func(o *readOptions[personPayload]) { o.maxBodySize = 16 }
	_, err := ReadJSON[personPayload](rec, req, opt)
	if err == nil || !strings.Contains(err.Error(), "body must not be larger than") {
		t.Fatalf("expected body too large error, got %v", err)
	}
}

func TestReadJSON_MultipleTopLevelValues(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost, "/", bytes.NewBufferString(`{"name":"a"}{"name":"b"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	_, err := ReadJSON[personPayload](rec, req)
	if err == nil || !strings.Contains(err.Error(), "single JSON value") {
		t.Fatalf("expected single JSON value error, got %v", err)
	}
}

// Additional coverage: a stream that returns an unexpected read error
type brokenReader int

func (brokenReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestReadJSON_UnexpectedReadError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", brokenReader(0))
	req.Header.Set("Content-Type", "application/json")

	_, err := ReadJSON[personPayload](rec, req)
	if err == nil {
		t.Fatalf("expected error for broken reader, got nil")
	}
}
