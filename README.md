# Web Framework

A lightweight, minimal HTTP routing and middleware framework for Go 1.22+.

## Features

- **Simple routing** - Clean API for defining routes with support for HTTP methods
- **Route grouping** - Organize routes with prefixes
- **Middleware support** - Composable middleware chain (CORS, Recovery, Logging)
- **Path parameters** - Extract dynamic values from route patterns (e.g., `/users/{id}`)
- **Query parameters** - Helper functions for extracting query strings
- **JSON helpers** - Built-in response helpers for JSON APIs
- **Request binding** - Parse JSON request bodies

## Installation

```bash
go get github.com/agustinnigrelli/web
```

## Quick Start

```go
package main

import (
	"net/http"
	"github.com/agustinnigrelli/web"
)

func main() {
	r := web.NewRouter()

	r.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		web.JsonResponse(w, http.StatusOK, map[string]string{
			"message": "Hello, World!",
		})
	})

	http.ListenAndServe(":8080", r)
}
```

## Routing

### HTTP Methods

```go
r.Get("/path", handler)
r.Post("/path", handler)
r.Put("/path", handler)
r.Patch("/path", handler)
r.Delete("/path", handler)
r.Head("/path", handler)
```

### Path Parameters

```go
r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
	id := web.GetParam(req, "id")
	web.JsonResponse(w, http.StatusOK, map[string]string{
		"id": id,
	})
})
```

### Route Grouping

```go
api := r.Group("/api")
api.Get("/users", getUsersHandler)
api.Post("/users", createUserHandler)

v2 := api.Group("/v2")
v2.Get("/users", getUsersV2Handler)
```

## Middleware

Middleware is applied in the order registered with `Use()`.

### Built-in Middleware

#### CORS Middleware

```go
r.Use(web.CORSMiddleware(
	[]string{"https://example.com", "http://localhost:3000"},
	[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	[]string{"Content-Type", "Authorization"},
))
```

**CORS behavior:**
- When `allowedOrigins` contains `"*"`, the middleware sets `Access-Control-Allow-Origin: *`
- Otherwise, it echoes the request `Origin` header if it matches an entry in `allowedOrigins`
- `Access-Control-Allow-Credentials: true` is only set for specific origins (not with wildcard)
- The `Vary: Origin` header is added to ensure proper caching

#### Recovery Middleware

Catches panics and returns HTTP 500 instead of crashing:

```go
r.Use(web.RecoveryMiddleware())
```

#### Logging Middleware

Logs all HTTP requests with method, path, status code, and response time:

```go
r.Use(web.LoggingMiddleware())
```

### Custom Middleware

```go
customMW := func(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Before handler
		println("Request:", r.Method, r.URL.Path)
		
		next(w, r)
		
		// After handler
		println("Response sent")
	}
}

r.Use(customMW)
```

## Request Helpers

### Get Query Parameters

```go
func handler(w http.ResponseWriter, req *http.Request) {
	page := web.GetQueryParam(req, "page")
	limit := web.GetQueryParam(req, "limit")
}
```

### Get Headers

```go
func handler(w http.ResponseWriter, req *http.Request) {
	token := web.GetHeader(req, "Authorization")
	contentType := web.GetHeader(req, "Content-Type")
}
```

### Bind JSON Body

```go
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func handler(w http.ResponseWriter, req *http.Request) {
	var user User
	if err := web.BindBody(req, &user); err != nil {
		web.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// Use user...
}
```

### Bind JSON Body with Size Limit

`BindBody` enforces a 10MB default limit. For custom limits:

```go
func handler(w http.ResponseWriter, req *http.Request) {
	var user User
	maxBytes := int64(5 << 20) // 5MB
	if err := web.BindBodyWithBytesLimit(req, &user, maxBytes); err != nil {
		if err == web.ErrBodyTooLarge {
			web.ErrorResponse(w, http.StatusRequestEntityTooLarge, "Body too large")
		} else {
			web.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		}
		return
	}
	// Use user...
}
```

## Response Helpers

### JSON Response

```go
web.JsonResponse(w, http.StatusOK, map[string]any{
	"id":   123,
	"name": "Alice",
})
```

### Error Response

```go
web.ErrorResponse(w, http.StatusBadRequest, "Invalid input")
```

## Complete Example

```go
package main

import (
	"net/http"
	"github.com/agustinnigrelli/web"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	r := web.NewRouter()

	// Middleware
	r.Use(web.LoggingMiddleware())
	r.Use(web.RecoveryMiddleware())
	r.Use(web.CORSMiddleware(
		[]string{"*"},
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		[]string{"Content-Type"},
	))

	// Routes
	r.Get("/users", listUsers)
	r.Get("/users/{id}", getUser)
	r.Post("/users", createUser)
	r.Put("/users/{id}", updateUser)
	r.Delete("/users/{id}", deleteUser)

	http.ListenAndServe(":8080", r)
}

func listUsers(w http.ResponseWriter, req *http.Request) {
	web.JsonResponse(w, http.StatusOK, map[string]any{
		"users": []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
		},
	})
}

func getUser(w http.ResponseWriter, req *http.Request) {
	id := web.GetParam(req, "id")
	web.JsonResponse(w, http.StatusOK, map[string]any{
		"id":    id,
		"name":  "Alice",
		"email": "alice@example.com",
	})
}

func createUser(w http.ResponseWriter, req *http.Request) {
	var user User
	if err := web.BindBody(req, &user); err != nil {
		web.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user.ID = 3
	web.JsonResponse(w, http.StatusCreated, user)
}

func updateUser(w http.ResponseWriter, req *http.Request) {
	id := web.GetParam(req, "id")
	var user User
	if err := web.BindBody(req, &user); err != nil {
		web.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	web.JsonResponse(w, http.StatusOK, map[string]any{
		"message": "User updated",
		"id":      id,
	})
}

func deleteUser(w http.ResponseWriter, req *http.Request) {
	id := web.GetParam(req, "id")
	web.JsonResponse(w, http.StatusOK, map[string]any{
		"message": "User deleted",
		"id":      id,
	})
}
```

## Project Structure

```
.
├── web.go                      # Public API
├── go.mod                       # Module definition
├── internal/
│   ├── router/
│   │   ├── router.go           # Router implementation
│   │   └── middlewares.go      # Built-in middlewares
│   ├── request/
│   │   └── request.go          # Request helpers
│   └── response/
│       └── response.go         # Response helpers
└── examples/
    └── main.go                 # Complete example
```

## Considerations

- **Go 1.22+** required for pattern-based routing support with path parameters
- **Minimal dependencies** - Only uses Go standard library
- **Unopinionated** - No database, template, or auth integrations included
- **Thread-safe** - Underlying `http.ServeMux` is safe for concurrent use

## License

MIT
