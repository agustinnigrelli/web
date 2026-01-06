package router

import (
	"net/http"
	"strings"
)

type Router struct {
	mux    *http.ServeMux
	prefix string
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) handle(method, pattern string, handler http.HandlerFunc) {
	fullPattern := r.prefix + pattern
	r.mux.HandleFunc(fullPattern, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, req)
	})
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

func (r *Router) Group(prefix string) *Router {
	p := "/" + strings.Trim(prefix, "/")
	return &Router{
		mux:    r.mux,
		prefix: r.prefix + p,
	}
}
