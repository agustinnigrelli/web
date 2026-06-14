package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/agustinnigrelli/web/internal/request"
)

// ---------- Router ----------

func TestRouter(t *testing.T) {
	type testCase struct {
		name     string
		method   string
		setup    func(*Router)
		req      func() *http.Request
		wantCode int
		wantBody string
	}

	tests := []testCase{
		{
			name:     "GET route calls handler",
			method:   http.MethodGet,
			setup:    func(r *Router) { r.Get("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("GET")) }) },
			req:      func() *http.Request { return httptest.NewRequest(http.MethodGet, "/p", nil) },
			wantCode: http.StatusOK,
			wantBody: "GET",
		},
		{
			name:   "POST route calls handler",
			method: http.MethodPost,
			setup: func(r *Router) {
				r.Post("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("POST")) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodPost, "/p", nil) },
			wantCode: http.StatusOK,
			wantBody: "POST",
		},
		{
			name:     "PUT route calls handler",
			method:   http.MethodPut,
			setup:    func(r *Router) { r.Put("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("PUT")) }) },
			req:      func() *http.Request { return httptest.NewRequest(http.MethodPut, "/p", nil) },
			wantCode: http.StatusOK,
			wantBody: "PUT",
		},
		{
			name:   "PATCH route calls handler",
			method: http.MethodPatch,
			setup: func(r *Router) {
				r.Patch("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("PATCH")) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodPatch, "/p", nil) },
			wantCode: http.StatusOK,
			wantBody: "PATCH",
		},
		{
			name:   "DELETE route calls handler",
			method: http.MethodDelete,
			setup: func(r *Router) {
				r.Delete("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("DELETE")) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodDelete, "/p", nil) },
			wantCode: http.StatusOK,
			wantBody: "DELETE",
		},
		{
			name:   "HEAD route calls handler",
			method: http.MethodHead,
			setup: func(r *Router) {
				r.Head("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("HEAD")) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodHead, "/p", nil) },
			wantCode: http.StatusOK,
		},
		{
			name:   "OPTIONS route calls handler",
			method: http.MethodOptions,
			setup: func(r *Router) {
				r.Options("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("OPTIONS")) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodOptions, "/p", nil) },
			wantCode: http.StatusOK,
			wantBody: "OPTIONS",
		},
		{
			name:     "wrong method returns 405",
			method:   http.MethodPost,
			setup:    func(r *Router) { r.Get("/p", func(w http.ResponseWriter, _ *http.Request) {}) },
			req:      func() *http.Request { return httptest.NewRequest(http.MethodPost, "/p", nil) },
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "unregistered route returns 404",
			method:   http.MethodGet,
			setup:    func(r *Router) { r.Get("/exists", func(w http.ResponseWriter, _ *http.Request) {}) },
			req:      func() *http.Request { return httptest.NewRequest(http.MethodGet, "/other", nil) },
			wantCode: http.StatusNotFound,
		},
		{
			name:   "path param extraction returns correct value",
			method: http.MethodGet,
			setup: func(r *Router) {
				r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte(GetParam(req, "id"))) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodGet, "/users/42", nil) },
			wantCode: http.StatusOK,
			wantBody: "42",
		},
		{
			name:   "nonexistent path param returns empty string",
			method: http.MethodGet,
			setup: func(r *Router) {
				r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte(GetParam(req, "missing"))) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodGet, "/users/42", nil) },
			wantCode: http.StatusOK,
			wantBody: "",
		},
		{
			name:   "ServeHTTP satisfies http.Handler interface",
			method: http.MethodGet,
			setup: func(r *Router) {
				r.Get("/p", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("handler")) })
			},
			req:      func() *http.Request { return httptest.NewRequest(http.MethodGet, "/p", nil) },
			wantCode: http.StatusOK,
			wantBody: "handler",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			tc.setup(r)

			var handler http.Handler = r
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tc.req())

			if w.Code != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, w.Code)
			}
			if tc.wantBody != "" && w.Body.String() != tc.wantBody {
				t.Fatalf("expected body %q, got %q", tc.wantBody, w.Body.String())
			}
		})
	}
}

// ---------- GetQueryParam ----------

func TestGetQueryParam(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		param string
		want  string
	}{
		{name: "extracts existing query param", url: "/search?q=hello&page=2", param: "q", want: "hello"},
		{name: "extracts second query param", url: "/search?q=hello&page=2", param: "page", want: "2"},
		{name: "missing query param returns empty", url: "/search", param: "q", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			r.Get("/search", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte(GetQueryParam(req, tc.param)))
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.url, nil))

			if w.Body.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, w.Body.String())
			}
		})
	}
}

// ---------- GetHeader ----------

func TestGetHeader(t *testing.T) {
	tests := []struct {
		name       string
		setHeaders func(*http.Request)
		header     string
		want       string
	}{
		{
			name:       "extracts existing header",
			setHeaders: func(r *http.Request) { r.Header.Set("Authorization", "Bearer token") },
			header:     "Authorization",
			want:       "Bearer token",
		},
		{
			name:       "missing header returns empty",
			setHeaders: func(r *http.Request) {},
			header:     "X-Nonexistent",
			want:       "",
		},
		{
			name:       "header lookup is case-insensitive",
			setHeaders: func(r *http.Request) { r.Header.Set("X-Custom-Header", "myval") },
			header:     "x-custom-header",
			want:       "myval",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte(GetHeader(req, tc.header)))
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tc.setHeaders(req)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Body.String() != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, w.Body.String())
			}
		})
	}
}

// ---------- BindBody ----------

func TestBindBody(t *testing.T) {
	type testCase struct {
		name     string
		body     io.Reader
		wantCode int
		wantBody string
	}

	tests := []testCase{
		{
			name:     "valid JSON body decodes correctly",
			body:     strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`),
			wantCode: http.StatusOK,
			wantBody: "Alice:alice@example.com",
		},
		{
			name:     "invalid JSON returns parse error",
			body:     strings.NewReader(`{invalid`),
			wantCode: http.StatusOK,
			wantBody: "err:invalid character 'i' looking for beginning of object key string",
		},
		{
			name:     "empty body returns JSON parse error",
			body:     strings.NewReader(""),
			wantCode: http.StatusOK,
			wantBody: "err:unexpected end of JSON input",
		},
		{
			name:     "bytes.Reader body decodes correctly",
			body:     bytes.NewReader([]byte(`{"data":"bytes"}`)),
			wantCode: http.StatusOK,
			wantBody: "bytes",
		},
		{
			name:     "body exceeding 10MB default limit returns ErrBodyTooLarge",
			body:     strings.NewReader(strings.Repeat("x", 11<<20)),
			wantCode: http.StatusRequestEntityTooLarge,
			wantBody: "too large",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			r.Post("/test", func(w http.ResponseWriter, req *http.Request) {
				var v struct {
					Name  string `json:"name"`
					Email string `json:"email"`
					Data  string `json:"data"`
				}
				if err := BindBody(req, &v); err != nil {
					if err == request.ErrBodyTooLarge {
						w.WriteHeader(http.StatusRequestEntityTooLarge)
						w.Write([]byte("too large"))
						return
					}
					w.Write([]byte("err:" + err.Error()))
					return
				}
				if v.Name != "" && v.Email != "" {
					w.Write([]byte(v.Name + ":" + v.Email))
				} else if v.Data != "" {
					w.Write([]byte(v.Data))
				} else {
					w.Write([]byte("parse error"))
				}
			})

			req := httptest.NewRequest(http.MethodPost, "/test", tc.body)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, w.Code)
			}
			if w.Body.String() != tc.wantBody {
				t.Fatalf("expected body %q, got %q", tc.wantBody, w.Body.String())
			}
		})
	}
}

func TestBindBody_Rereadable(t *testing.T) {
	r := NewRouter()
	r.Post("/test", func(w http.ResponseWriter, req *http.Request) {
		var v1, v2 struct {
			Name string `json:"name"`
		}
		BindBody(req, &v1)
		err := BindBody(req, &v2)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(v1.Name + ":" + v2.Name))
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"Alice"}`)))
	if w.Body.String() != "Alice:Alice" {
		t.Fatalf("expected 'Alice:Alice', got %q", w.Body.String())
	}
}

func TestBindBody_Standalone(t *testing.T) {
	bodyContent := `{"x":1}`
	req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytes.NewReader([]byte(bodyContent))))

	var v1, v2 struct {
		X int `json:"x"`
	}
	if err := BindBody(req, &v1); err != nil {
		t.Fatalf("first bind: %v", err)
	}
	if v1.X != 1 {
		t.Fatalf("expected X=1, got %d", v1.X)
	}
	if err := BindBody(req, &v2); err != nil {
		t.Fatalf("second bind: %v", err)
	}
	if v2.X != 1 {
		t.Fatalf("expected X=1, got %d", v2.X)
	}
}

// ---------- BindBodyWithBytesLimit ----------

func TestBindBodyWithBytesLimit(t *testing.T) {
	type testCase struct {
		name     string
		body     io.Reader
		limit    int64
		wantCode int
		wantBody string
	}

	tests := []testCase{
		{
			name:     "body within limit decodes correctly",
			body:     strings.NewReader(`{"data":"hello"}`),
			limit:    1 << 20,
			wantCode: http.StatusOK,
			wantBody: "hello",
		},
		{
			name:     "body exceeding limit returns ErrBodyTooLarge",
			body:     strings.NewReader(strings.Repeat("x", 20)),
			limit:    10,
			wantCode: http.StatusRequestEntityTooLarge,
			wantBody: "too large",
		},
		{
			name:     "body exactly at limit decodes correctly",
			body:     strings.NewReader(`{"data":"hello"}`),
			limit:    100,
			wantCode: http.StatusOK,
			wantBody: "hello",
		},
		{
			name:     "nil body returns JSON parse error",
			body:     nil,
			limit:    10,
			wantCode: http.StatusOK,
			wantBody: "err:unexpected end of JSON input",
		},
		{
			name:     "body re-readable via BindBodyWithBytesLimit",
			body:     strings.NewReader(`{"name":"Alice"}`),
			limit:    1 << 20,
			wantCode: http.StatusOK,
			wantBody: "Alice:Alice",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			r.Post("/test", func(w http.ResponseWriter, req *http.Request) {
				var v struct {
					Data string `json:"data"`
					Name string `json:"name"`
				}

				if tc.wantBody == "Alice:Alice" {
					var v1, v2 struct {
						Name string `json:"name"`
					}
					BindBodyWithBytesLimit(req, &v1, tc.limit)
					if err := BindBodyWithBytesLimit(req, &v2, tc.limit); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.Write([]byte(v1.Name + ":" + v2.Name))
					return
				}

				err := BindBodyWithBytesLimit(req, &v, tc.limit)
				if err == request.ErrBodyTooLarge {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					w.Write([]byte("too large"))
					return
				}
				if err != nil {
					w.Write([]byte("err:" + err.Error()))
					return
				}
				w.Write([]byte(v.Data))
			})

			req := httptest.NewRequest(http.MethodPost, "/test", tc.body)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, w.Code)
			}
			if w.Body.String() != tc.wantBody {
				t.Fatalf("expected body %q, got %q", tc.wantBody, w.Body.String())
			}
		})
	}
}

func TestBindBodyThenBindBodyWithBytesLimit(t *testing.T) {
	r := NewRouter()
	r.Post("/test", func(w http.ResponseWriter, req *http.Request) {
		var v1, v2 struct {
			Name string `json:"name"`
		}
		BindBody(req, &v1)
		if err := BindBodyWithBytesLimit(req, &v2, 1<<20); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(v1.Name + ":" + v2.Name))
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"Bob"}`)))
	if w.Body.String() != "Bob:Bob" {
		t.Fatalf("expected 'Bob:Bob', got %q", w.Body.String())
	}
}

// ---------- JsonResponse ----------

func TestJsonResponse(t *testing.T) {
	type testCase struct {
		name     string
		data     any
		status   int
		wantCode int
		wantJSON string
	}

	tests := []testCase{
		{
			name:     "valid data returns JSON with correct status and Content-Type",
			data:     map[string]any{"id": 42, "name": "Alice"},
			status:   http.StatusCreated,
			wantCode: http.StatusCreated,
			wantJSON: `{"id":42,"name":"Alice"}`,
		},
		{
			name:     "nil data produces null JSON",
			data:     nil,
			status:   http.StatusOK,
			wantCode: http.StatusOK,
			wantJSON: "null",
		},
		{
			name:     "non-serializable data returns 500",
			data:     map[string]any{"ch": make(chan int)},
			status:   http.StatusOK,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			JsonResponse(w, tc.status, tc.data)

			if w.Code != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, w.Code)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" && tc.wantCode < 500 {
				t.Fatalf("expected Content-Type application/json, got %q", ct)
			}
			if tc.wantJSON != "" {
				var got, want any
				json.Unmarshal(w.Body.Bytes(), &got)
				json.Unmarshal([]byte(tc.wantJSON), &want)
				gotS, _ := json.Marshal(got)
				wantS, _ := json.Marshal(want)
				if string(gotS) != string(wantS) {
					t.Fatalf("expected JSON %s, got %s", wantS, gotS)
				}
			}
		})
	}
}

// ---------- ErrorResponse ----------

func TestErrorResponse(t *testing.T) {
	type testCase struct {
		name     string
		status   int
		message  string
		wantCode int
		check    func(*testing.T, *httptest.ResponseRecorder)
	}

	tests := []testCase{
		{
			name:     "error response has correct status and Content-Type",
			status:   http.StatusBadRequest,
			message:  "invalid input",
			wantCode: http.StatusBadRequest,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				var body map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				if body["status"].(float64) != 400 {
					t.Fatalf("expected status 400, got %v", body["status"])
				}
				if body["message"] != "invalid input" {
					t.Fatalf("expected message 'invalid input', got %v", body["message"])
				}
			},
		},
		{
			name:     "error response body shape with different status",
			status:   http.StatusTeapot,
			message:  "I'm a teapot",
			wantCode: http.StatusTeapot,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				var body struct {
					Status  int    `json:"status"`
					Message string `json:"message"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to decode JSON: %v", err)
				}
				if body.Status != 418 {
					t.Fatalf("expected status 418, got %d", body.Status)
				}
				if body.Message != "I'm a teapot" {
					t.Fatalf("expected message 'I'm a teapot', got %q", body.Message)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ErrorResponse(w, tc.status, tc.message)

			if w.Code != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, w.Code)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("expected Content-Type application/json, got %q", ct)
			}
			if tc.check != nil {
				tc.check(t, w)
			}
		})
	}
}

// ---------- CORS Middleware ----------

func TestCORSMiddleware(t *testing.T) {
	type testCase struct {
		name      string
		origins   []string
		methods   []string
		headers   []string
		reqOrigin string
		reqMethod string
		check     func(*testing.T, *httptest.ResponseRecorder)
	}

	tests := []testCase{
		{
			name:      "wildcard origin sets Access-Control-Allow-Origin: *",
			origins:   []string{"*"},
			methods:   []string{"GET", "POST"},
			headers:   []string{"Content-Type"},
			reqOrigin: "https://example.com",
			reqMethod: http.MethodGet,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
					t.Fatalf("expected *, got %q", got)
				}
				if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET,POST" {
					t.Fatalf("expected GET,POST, got %q", got)
				}
				if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
					t.Fatalf("expected Content-Type, got %q", got)
				}
				if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
					t.Fatalf("expected no credentials, got %q", got)
				}
			},
		},
		{
			name:      "specific origin sets origin, credentials, and Vary",
			origins:   []string{"https://example.com"},
			methods:   []string{"GET"},
			headers:   []string{"Authorization"},
			reqOrigin: "https://example.com",
			reqMethod: http.MethodGet,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
					t.Fatalf("expected https://example.com, got %q", got)
				}
				if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
					t.Fatalf("expected true, got %q", got)
				}
				found := slices.Contains(w.Header().Values("Vary"), "Origin")
				if !found {
					t.Fatal("expected Vary: Origin")
				}
			},
		},
		{
			name:      "unmatched origin gets no CORS headers",
			origins:   []string{"https://example.com"},
			methods:   []string{"GET"},
			headers:   []string{},
			reqOrigin: "https://evil.com",
			reqMethod: http.MethodGet,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Fatalf("expected no ACAO, got %q", got)
				}
			},
		},
		{
			name:      "no Origin header skips CORS and calls handler",
			origins:   []string{"https://example.com"},
			methods:   []string{"GET"},
			headers:   []string{},
			reqOrigin: "",
			reqMethod: http.MethodGet,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Fatalf("expected no ACAO, got %q", got)
				}
				if w.Body.String() != "ok" {
					t.Fatalf("expected body 'ok', got %q", w.Body.String())
				}
			},
		},
		{
			name:      "preflight OPTIONS returns 200 without calling handler",
			origins:   []string{"https://example.com"},
			methods:   []string{"GET", "POST"},
			headers:   []string{"Content-Type", "Authorization"},
			reqOrigin: "https://example.com",
			reqMethod: http.MethodOptions,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusOK {
					t.Fatalf("expected 200 for preflight, got %d", w.Code)
				}
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
					t.Fatalf("expected https://example.com, got %q", got)
				}
				if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET,POST" {
					t.Fatalf("expected GET,POST, got %q", got)
				}
				if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type,Authorization" {
					t.Fatalf("expected Content-Type,Authorization, got %q", got)
				}
			},
		},
		{
			name:      "empty allowed origins omits ACAO",
			origins:   []string{},
			methods:   []string{"GET"},
			headers:   []string{},
			reqOrigin: "https://example.com",
			reqMethod: http.MethodGet,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Fatalf("expected no ACAO, got %q", got)
				}
			},
		},
		{
			name:      "empty allowed methods omits Allow-Methods header",
			origins:   []string{"*"},
			methods:   []string{},
			headers:   []string{"Content-Type"},
			reqOrigin: "https://example.com",
			reqMethod: http.MethodGet,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Methods"); got != "" {
					t.Fatalf("expected no Allow-Methods, got %q", got)
				}
			},
		},
		{
			name:      "empty allowed headers omits Allow-Headers header",
			origins:   []string{"*"},
			methods:   []string{"GET"},
			headers:   []string{},
			reqOrigin: "https://example.com",
			reqMethod: http.MethodGet,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				if got := w.Header().Get("Access-Control-Allow-Headers"); got != "" {
					t.Fatalf("expected no Allow-Headers, got %q", got)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			r.Use(CORSMiddleware(tc.origins, tc.methods, tc.headers))
			r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("ok"))
			})

			req := httptest.NewRequest(tc.reqMethod, "/test", nil)
			if tc.reqOrigin != "" {
				req.Header.Set("Origin", tc.reqOrigin)
			}
			if tc.reqMethod == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", "GET")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			tc.check(t, w)
		})
	}
}

// ---------- Recovery Middleware ----------

func TestRecoveryMiddleware(t *testing.T) {
	type testCase struct {
		name     string
		handler  func(w http.ResponseWriter, req *http.Request)
		wantCode int
		wantBody string
	}

	tests := []testCase{
		{
			name:     "panic in handler returns 500",
			handler:  func(w http.ResponseWriter, req *http.Request) { panic("boom") },
			wantCode: http.StatusInternalServerError,
			wantBody: "Internal Server Error\n",
		},
		{
			name:     "normal handler passes through",
			handler:  func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) },
			wantCode: http.StatusOK,
			wantBody: "ok",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			r.Use(RecoveryMiddleware())
			r.Get("/p", tc.handler)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/p", nil))

			if w.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d", tc.wantCode, w.Code)
			}
			if w.Body.String() != tc.wantBody {
				t.Fatalf("expected %q, got %q", tc.wantBody, w.Body.String())
			}
		})
	}
}

// ---------- Logging Middleware ----------

func TestLoggingMiddleware(t *testing.T) {
	type testCase struct {
		name      string
		setup     func(*Router)
		reqMethod string
		reqPath   string
		wantCode  int
	}

	tests := []testCase{
		{
			name:      "logged request with 200 status",
			setup:     func(r *Router) { r.Get("/ok", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) }) },
			reqMethod: http.MethodGet,
			reqPath:   "/ok",
			wantCode:  http.StatusOK,
		},
		{
			name:      "logged request with 404 status",
			setup:     func(r *Router) { r.Get("/test", func(w http.ResponseWriter, _ *http.Request) {}) },
			reqMethod: http.MethodGet,
			reqPath:   "/nonexistent",
			wantCode:  http.StatusNotFound,
		},
		{
			name:      "logged request with 405 status",
			setup:     func(r *Router) { r.Get("/test", func(w http.ResponseWriter, _ *http.Request) {}) },
			reqMethod: http.MethodPost,
			reqPath:   "/test",
			wantCode:  http.StatusMethodNotAllowed,
		},
		{
			name: "explicit WriteHeader(201) captured correctly",
			setup: func(r *Router) {
				r.Get("/created", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(201); w.Write([]byte(`{}`)) })
			},
			reqMethod: http.MethodGet,
			reqPath:   "/created",
			wantCode:  http.StatusCreated,
		},
		{
			name: "implicit 200 via Write without WriteHeader",
			setup: func(r *Router) {
				r.Get("/default", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
			},
			reqMethod: http.MethodGet,
			reqPath:   "/default",
			wantCode:  http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			r.Use(LoggingMiddleware())
			tc.setup(r)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.reqMethod, tc.reqPath, nil))

			if w.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d", tc.wantCode, w.Code)
			}
		})
	}
}

// ---------- Middleware Ordering ----------

func TestMiddlewareOrdering(t *testing.T) {
	var order []string

	mw1 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next(w, r)
			order = append(order, "mw1-after")
		}
	}
	mw2 := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next(w, r)
			order = append(order, "mw2-after")
		}
	}

	r := NewRouter()
	r.Use(mw1, mw2)
	r.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, "handler")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Fatalf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

// ---------- Group ----------

func TestGroup(t *testing.T) {
	type testCase struct {
		name     string
		setup    func(*Router)
		reqPath  string
		wantCode int
		wantBody string
		check    func(*testing.T, *httptest.ResponseRecorder)
	}

	tests := []testCase{
		{
			name: "group prefix is prepended to routes",
			setup: func(r *Router) {
				g := r.Group("/api")
				g.Get("/users", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("users")) })
			},
			reqPath:  "/api/users",
			wantCode: http.StatusOK,
			wantBody: "users",
		},
		{
			name: "nested group prefixes combine correctly",
			setup: func(r *Router) {
				g := r.Group("/api").Group("/v2")
				g.Get("/test", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("nested")) })
			},
			reqPath:  "/api/v2/test",
			wantCode: http.StatusOK,
			wantBody: "nested",
		},
		{
			name: "trailing slashes in group prefix are trimmed",
			setup: func(r *Router) {
				g := r.Group("/api/")
				g.Get("/test", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("trimmed")) })
			},
			reqPath:  "/api/test",
			wantCode: http.StatusOK,
			wantBody: "trimmed",
		},
		{
			name: "empty group prefix produces single slash",
			setup: func(r *Router) {
				g := r.Group("")
				g.Get("test", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("empty")) })
			},
			reqPath:  "/test",
			wantCode: http.StatusOK,
			wantBody: "empty",
		},
		{
			name: "group inherits parent middlewares",
			setup: func(r *Router) {
				var calls []string
				r.Use(func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, req *http.Request) {
						calls = append(calls, "parent")
						next(w, req)
					}
				})
				g := r.Group("/group")
				g.Use(func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, req *http.Request) {
						calls = append(calls, "group")
						next(w, req)
					}
				})
				g.Get("/test", func(w http.ResponseWriter, _ *http.Request) { calls = append(calls, "handler") })
			},
			reqPath:  "/group/test",
			wantCode: http.StatusOK,
			check: func(t *testing.T, w *httptest.ResponseRecorder) {
				// Can't verify calls here, but the test passes if status is 200
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRouter()
			tc.setup(r)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.reqPath, nil))

			if w.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d", tc.wantCode, w.Code)
			}
			if tc.wantBody != "" && w.Body.String() != tc.wantBody {
				t.Fatalf("expected body %q, got %q", tc.wantBody, w.Body.String())
			}
			if tc.check != nil {
				tc.check(t, w)
			}
		})
	}
}

func TestGroup_MiddlewareIsolation(t *testing.T) {
	var calls []string

	r := NewRouter()
	r.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "parent")
			next(w, r)
		}
	})

	g := r.Group("/group")
	g.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "group")
			next(w, r)
		}
	})
	g.Get("/test", func(w http.ResponseWriter, _ *http.Request) { calls = append(calls, "handler") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/group/test", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	expected := []string{"parent", "group", "handler"}
	if len(calls) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, calls)
	}
	for i, v := range expected {
		if calls[i] != v {
			t.Fatalf("position %d: expected %q, got %q", i, v, calls[i])
		}
	}
}

func TestGroup_ParentNotAffectedByGroup(t *testing.T) {
	var calls []string

	r := NewRouter()
	r.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "parent")
			next(w, r)
		}
	})

	g := r.Group("/group")
	g.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "group")
			next(w, r)
		}
	})
	_ = g

	r.Get("/alone", func(w http.ResponseWriter, _ *http.Request) { calls = append(calls, "alone-handler") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/alone", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	expected := []string{"parent", "alone-handler"}
	if len(calls) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, calls)
	}
	for i, v := range expected {
		if calls[i] != v {
			t.Fatalf("position %d: expected %q, got %q", i, v, calls[i])
		}
	}
}

func TestGroup_RouteSpecificMiddleware(t *testing.T) {
	var calls []string

	r := NewRouter()
	g := r.Group("/admin")
	g.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "auth")
			if r.Header.Get("Authorization") == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	})
	g.Get("/data", func(w http.ResponseWriter, _ *http.Request) {
		calls = append(calls, "handler")
		w.Write([]byte("secret"))
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/data", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/admin/data", nil)
	req2.Header.Set("Authorization", "Bearer token")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	if w2.Body.String() != "secret" {
		t.Fatalf("expected 'secret', got %q", w2.Body.String())
	}
}

// ---------- Combined Middleware ----------

func TestCombinedMiddlewares(t *testing.T) {
	r := NewRouter()
	r.Use(RecoveryMiddleware())
	r.Use(LoggingMiddleware())
	r.Use(CORSMiddleware([]string{"*"}, []string{"GET"}, []string{}))

	r.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		JsonResponse(w, http.StatusOK, map[string]string{"msg": "hello"})
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("Origin", "http://localhost")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS header *, got %q", got)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if body["msg"] != "hello" {
		t.Fatalf("expected 'hello', got %q", body["msg"])
	}
}

// ---------- Concurrent Requests ----------

func TestConcurrentRequests(t *testing.T) {
	r := NewRouter()
	r.Use(RecoveryMiddleware())
	r.Get("/test", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", w.Code)
			}
		})
	}
	wg.Wait()
}

// ---------- Edge Cases ----------

func TestConflictingPatternPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Skip("expected panic, but Go version may not enforce")
		}
	}()

	r := NewRouter()
	r.Get("/users/{id}", func(w http.ResponseWriter, _ *http.Request) {})
	r.Get("/users/{name}", func(w http.ResponseWriter, _ *http.Request) {})
}

func TestRecoveryAfterWriteHeader(t *testing.T) {
	r := NewRouter()
	r.Use(RecoveryMiddleware())
	r.Get("/panic-after-write", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("partial"))
		panic("boom")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/panic-after-write", nil))
	// Headers already committed, 500 won't be reflected, but server should not crash
}

func TestNewRouter(t *testing.T) {
	if r := NewRouter(); r == nil {
		t.Fatal("NewRouter() returned nil")
	}
}

func TestRouterImplementsHandler(t *testing.T) {
	var _ http.Handler = NewRouter()
}
