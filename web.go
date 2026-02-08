package web

import (
	"net/http"

	"github.com/agustinnigrelli/web/internal/request"
	"github.com/agustinnigrelli/web/internal/response"
	"github.com/agustinnigrelli/web/internal/router"
)

type Router = router.Router
type Middleware = router.Middleware

func NewRouter() *Router {
	return router.NewRouter()
}

// Response helpers
func JsonResponse(w http.ResponseWriter, status int, data any) {
	response.JsonResponse(w, status, data)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	response.ErrorResponse(w, status, message)
}

// Request helpers
func GetParam(r *http.Request, name string) string {
	return request.GetParam(r, name)
}

func GetQueryParam(r *http.Request, name string) string {
	return request.GetQueryParam(r, name)
}

func GetHeader(r *http.Request, name string) string {
	return request.GetHeader(r, name)
}

func BindBody(r *http.Request, v any) error {
	return request.BindBody(r, v)
}

func BindBodyWithBytesLimit(r *http.Request, v any, maxBytes int64) error {
	return request.BindBodyWithBytesLimit(r, v, maxBytes)
}

// Middleware functions
func CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders []string) Middleware {
	return router.CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders)
}

func RecoveryMiddleware() Middleware {
	return router.RecoveryMiddleware()
}

func LoggingMiddleware() Middleware {
	return router.LoggingMiddleware()
}
