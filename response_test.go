package grape

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hossein1376/grape/errs"
	"github.com/hossein1376/grape/reqid"
)

// TestRespondWritesJSON verifies Respond writes JSON body and headers when data
// is provided.
func TestRespondWritesJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := context.Background()

	Respond(ctx, rec, http.StatusCreated, Map{"ok": true, "n": 5})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d got %d", http.StatusCreated, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", ct)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if body["ok"] != true {
		t.Fatalf("unexpected body content: %v", body)
	}
}

// TestRespondNoContentWhenDataNil verifies Respond returns NoContent and an
// empty body when data is nil.
func TestRespondNoContentWhenDataNil(t *testing.T) {
	rec := httptest.NewRecorder()
	Respond(context.Background(), rec, http.StatusNoContent, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d got %d", http.StatusNoContent, rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf(
			"expected empty body for no-content response, got %q",
			rec.Body.String(),
		)
	}
}

// TestExtractFromErrNil verifies ExtractFromErr produces NoContent when error
// is nil.
func TestExtractFromErrNil(t *testing.T) {
	rec := httptest.NewRecorder()
	ExtractFromErr(context.Background(), rec, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d got %d", http.StatusNoContent, rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf(
			"expected empty body when error is nil, got %q", rec.Body.String(),
		)
	}
}

// TestExtractFromErrErrsErrorWithCustomMessage verifies ExtractFromErr extracts
// message and status from errs.Error.
func TestExtractFromErrErrsErrorWithCustomMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := context.Background()

	e := errs.BadRequest(errs.WithMsg("bad input"))
	ExtractFromErr(ctx, rec, e)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d got %d", http.StatusBadRequest, rec.Code)
	}

	var r Response
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if r.Message == nil || r.Message != "bad input" {
		t.Fatalf("expected message 'bad input', got %v", r.Message)
	}
}

// TestExtractFromErrErrsErrorNoMessage verifies ExtractFromErr uses HTTP status
// text when errs.Error has no Message.
func TestExtractFromErrErrsErrorNoMessage(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := context.Background()

	e := errs.NotFound() // no custom message provided
	ExtractFromErr(ctx, rec, e)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d got %d", http.StatusNotFound, rec.Code)
	}

	var r Response
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if r.Message != http.StatusText(http.StatusNotFound) {
		t.Fatalf(
			"expected message %q got %v",
			http.StatusText(http.StatusNotFound),
			r.Message,
		)
	}
}

// TestExtractFromErrGenericErrorIncludesReqID verifies that for generic errors
// ExtractFromErr returns 500 and includes request id from context.
func TestExtractFromErrGenericErrorIncludesReqID(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := context.WithValue(
		context.Background(), reqid.RequestIDKey, reqid.ReqID("my-id-123"),
	)

	ExtractFromErr(ctx, rec, errors.New("error"))

	rec = httptest.NewRecorder()
	ExtractFromErr(ctx, rec, http.ErrHandlerTimeout) // generic error

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
	var r Response
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if r.Message != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf(
			"expected message %q got %v",
			http.StatusText(http.StatusInternalServerError),
			r.Message,
		)
	}
	if r.Data != "my-id-123" {
		t.Fatalf("expected Data to equal request id, got %v", r.Data)
	}
}
