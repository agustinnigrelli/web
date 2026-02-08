package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrBodyTooLarge = errors.New("request body too large")

func GetParam(r *http.Request, name string) string {
	return r.PathValue(name)
}

func GetQueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func GetHeader(r *http.Request, name string) string {
	return r.Header.Get(name)
}

func BindBody(r *http.Request, v any) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return json.Unmarshal(bodyBytes, v)
}

func BindBodyWithBytesLimit(r *http.Request, v any, maxBytes int64) error {
	lr := io.LimitReader(r.Body, maxBytes+1)
	bodyBytes, err := io.ReadAll(lr)
	if err != nil {
		return err
	}
	if int64(len(bodyBytes)) > maxBytes {
		return ErrBodyTooLarge
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return json.Unmarshal(bodyBytes, v)
}
