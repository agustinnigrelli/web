package router

import (
	"net/http"
	"strings"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Router struct {
	mux             *http.ServeMux
	prefix          string
	middlewares     []Middleware
	optionsPatterns map[string]bool
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux(), optionsPatterns: map[string]bool{}}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Middlewares are applied in the order they are passed to Use.
func (r *Router) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Router) handle(method, pattern string, handler http.HandlerFunc) {
	fullPattern := method + " " + r.prefix + pattern

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	r.mux.HandleFunc(fullPattern, handler)

	if method != http.MethodOptions {
		optionsPattern := http.MethodOptions + " " + r.prefix + pattern
		if !r.optionsPatterns[optionsPattern] {
			r.optionsPatterns[optionsPattern] = true
			optionsHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}
			for i := len(r.middlewares) - 1; i >= 0; i-- {
				optionsHandler = r.middlewares[i](optionsHandler)
			}
			r.mux.HandleFunc(optionsPattern, optionsHandler)
		}
	}
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodGet, pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodPost, pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodPut, pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodDelete, pattern, handler)
}

func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodPatch, pattern, handler)
}

func (r *Router) Head(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodHead, pattern, handler)
}

func (r *Router) Group(prefix string) *Router {
	p := "/" + strings.Trim(prefix, "/")
	return &Router{
		mux:             r.mux,
		prefix:          r.prefix + p,
		middlewares:     append([]Middleware(nil), r.middlewares...),
		optionsPatterns: r.optionsPatterns,
	}
}

func (r *Router) Options(pattern string, handler http.HandlerFunc) {
	r.handle(http.MethodOptions, pattern, handler)
}
