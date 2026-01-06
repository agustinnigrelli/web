package web

import (
	"net/http"

	"github.com/agustinnigrelli/web/internal/response"
	"github.com/agustinnigrelli/web/internal/router"
)

type Router = router.Router

func NewRouter() *Router {
	return router.NewRouter()
}

func JsonResponse(w http.ResponseWriter, status int, data any) {
	response.JsonResponse(w, status, data)
}
