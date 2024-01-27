package main

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

// serializer implements grape.Serializer.
// I have intentionally used an external dependency just to demonstrate you can use anything you would like.
type serializer struct{}

func newSerializer() serializer {
	return serializer{}
}

func (serializer) WriteJson(w http.ResponseWriter, status int, data any, _ http.Header) error {
	js, err := jsoniter.Marshal(data)
	if err != nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

func (serializer) ReadJson(_ http.ResponseWriter, r *http.Request, dst any) error {
	dec := jsoniter.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		return err
	}

	return nil
}
