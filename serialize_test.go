package grape

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_serialize_WriteJson(t *testing.T) {
	type response struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		w       http.ResponseWriter
		status  int
		data    any
		headers http.Header
		resp    resp
		err     error
	}{
		{
			name:   "Status code 204 with no body",
			w:      &httptest.ResponseRecorder{},
			status: 204,
			data:   nil,
			err:    nil,
		},
		{
			name:   "status code 201 with struct body",
			w:      &httptest.ResponseRecorder{},
			status: 204,
			data:   response{Name: "grape", Age: 2},
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := serialize{}.WriteJson(tt.w, tt.status, tt.data, tt.headers)
			code := tt.w.(*httptest.ResponseRecorder).Code

			if !errors.Is(err, tt.err) {
				t.Errorf("WriteJson() error = %v, wantErr %v", err, tt.err)
			}
			if code != tt.status {
				t.Errorf("WriteJson() status code = %v, expected %v", code, tt.status)
			}
		})
	}
}

func Test_serialize_ReadJson(t *testing.T) {
	type request struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name     string
		w        http.ResponseWriter
		r        *http.Request
		dst      *request
		response request
		err      error
	}{
		{
			name:     "Read request",
			w:        httptest.NewRecorder(),
			r:        httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name": "Grape"}`)),
			dst:      &request{},
			response: request{Name: "Grape"},
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := serialize{}.ReadJson(tt.w, tt.r, tt.dst)

			if !errors.Is(err, tt.err) {
				t.Errorf("ReadJson() error = %v, wantErr %v", err, tt.err)
			}
			if tt.dst.Name != tt.response.Name {
				t.Errorf("ReadJson() got %v, want %v", tt.dst.Name, tt.response.Name)
			}
		})
	}
}
