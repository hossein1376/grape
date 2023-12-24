package grape

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Serializer interface consists of two methods, one for reading json inputs and one for writing json outputs.
type Serializer interface {
	WriteJson(w http.ResponseWriter, status int, data any, headers http.Header) error
	ReadJson(w http.ResponseWriter, r *http.Request, dst any) error
}

// serializer implements Serializer interface.
type serializer struct {
	maxBytes int64
}

// Default request's body maximum size, if not overridden by grape.Options.
var maxBodySize int64 = 1_048_576

// WriteJson will write back data in json format with the provided status code and headers.
// It automatically sets content-type and date headers. To override, provide them as headers.
func (serializer) WriteJson(w http.ResponseWriter, status int, data any, headers http.Header) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Date", time.Now().Format(http.TimeFormat))

	for key, value := range headers {
		w.Header()[key] = value
	}

	if data == nil {
		w.WriteHeader(status)
		return nil
	}

	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

// ReadJson will decode incoming json requests. It will return a human-readable error in case of failure.
func (s serializer) ReadJson(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, s.maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &invalidUnmarshalError):
			return err

		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", s.maxBytes)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
