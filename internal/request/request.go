package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrBodyTooLarge = errors.New("request body too large")

const DefaultMaxBytes int64 = 10 << 20

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
	lr := io.LimitReader(r.Body, DefaultMaxBytes+1)
	bodyBytes, err := io.ReadAll(lr)
	if err != nil {
		return err
	}
	if int64(len(bodyBytes)) > DefaultMaxBytes {
		return ErrBodyTooLarge
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
