package router

import (
	"net/http"
	"strings"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type Router struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []Middleware
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Middlewares are applied in the order they are passed to Use.
func (r *Router) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Router) handle(method, pattern string, handler http.HandlerFunc) {
	fullPattern := r.prefix + pattern

	routeHandler := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, req)
	}

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		routeHandler = r.middlewares[i](routeHandler)
	}

	r.mux.HandleFunc(fullPattern, routeHandler)
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
		mux:         r.mux,
		prefix:      r.prefix + p,
		middlewares: append([]Middleware(nil), r.middlewares...),
	}
}
