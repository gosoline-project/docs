---
sidebar:
  order: 1
title: Build an HTTP service
---

In this guide, you will build an HTTP service from scratch. You will add routes, typed handlers, dependency injection, request binding, and negotiated responses.

## What is httpserver?

The httpserver package is built on [Gin](https://gin-gonic.com/), an HTTP framework for Go. You can use Gin routing and middleware, and you can register raw `*gin.Context` handlers when needed.

httpserver adds a typed way to build HTTP services. The **`With` pattern** gives each handler group a constructor that receives `ctx`, `config`, and `logger`. The constructor initializes dependencies that related routes share. The **`Bind` function** uses struct tags to bind and validate request data. Typed handlers return Go values. httpserver selects a response format from the request's `Accept` header and writes the response.

Together with built-in middleware for logging, metrics, compression, and graceful shutdown — all configured through YAML — httpserver lets you focus on your application logic rather than HTTP plumbing.

## Getting started

Install the package:

```bash
go get github.com/gosoline-project/httpserver@v0.6.4
```

The entry point for any HTTP server is a `RouterFactory` — a function that receives a `*Router` and registers routes on it:

```go
httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    router.GET("/hello", httpserver.BindN(func(ctx context.Context) (map[string]string, error) {
        return map[string]string{"message": "Hello, World!"}, nil
    }))
    return nil
})
```

`RunDefaultServer` creates a server named `"default"` reading from the `httpserver.default` config key. All you need is a `config.dist.yml`:

```yaml
httpserver:
  default:
    port: 8088
```

You can use standard [Gin](https://gin-gonic.com/) handlers directly. The `Router` supports all HTTP methods:

```go
router.GET("/users", listUsers)
router.POST("/users", createUser)
router.PUT("/users/:id", updateUser)
router.DELETE("/users/:id", deleteUser)
```

Group related routes with `router.Group()`:

```go
api := router.Group("/api")
api.GET("/ping", httpserver.BindN(func(ctx context.Context) (map[string]string, error) {
    return map[string]string{"message": "pong"}, nil
}))
```

Raw Gin handlers remain useful when you need direct access to Gin APIs. Use typed handlers when you want request binding and negotiated output.

## The With pattern

`httpserver.With` is the recommended way to define routes. It's a generic function that:

1. Calls a **handler factory** to create a handler instance (with access to `ctx`, `config`, and `logger`)
2. Calls a **registration function** that wires handler methods to routes

```go
router.Group("/api/users").HandleWith(httpserver.With(NewHandler, func(r *httpserver.Router, h *Handler) {
    r.GET("", httpserver.Bind(h.ListUsers))
    r.POST("", httpserver.Bind(h.CreateUser))
    r.GET("/:id", httpserver.Bind(h.GetUser))
    r.DELETE("/:id", httpserver.Bind(h.DeleteUser))
}))
```

### Handler factory

The handler factory must have this signature:

```go
func NewHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*Handler, error)
```

Use it to initialize dependencies — database clients, external services, caches:

```go
type Handler struct {
    db *sql.DB
}

func NewHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*Handler, error) {
    db, err := sql.Open("postgres", config.GetString("db_url"))
    if err != nil {
        return nil, err
    }
    return &Handler{db: db}, nil
}
```

### Bind, BindR, BindN, and BindNR

The standard binding helpers differ by two concerns: whether the handler receives a bound input struct, and whether it receives the raw `*http.Request`.

The suffixes follow a simple pattern:

- **`R`** means the raw `*http.Request` is passed to the handler.
- **`N`** means there is no input struct to bind.

| Helper | Handler shape | Use when |
|---|---|---|
| `Bind[I, O any]` | `func(ctx context.Context, input *I) (O, error)` | The endpoint needs request data bound into an input struct. This is the default choice for most endpoints. |
| `BindR[I, O any]` | `func(ctx context.Context, req *http.Request, input *I) (O, error)` | The endpoint needs both bound input and raw request access, for example headers, method, body metadata, or client IP resolution. |
| `BindN[O any]` | `func(ctx context.Context) (O, error)` | The endpoint does not need request input, for example health checks or static status endpoints. |
| `BindNR[O any]` | `func(ctx context.Context, req *http.Request) (O, error)` | The endpoint does not need a bound input struct, but still needs raw request access. |

`Bind` and `BindR` can bind request data from JSON bodies, query parameters, form data, URI parameters, headers, and other supported sources. `BindN` and `BindNR` skip request binding entirely because there is no input struct.

```go
// With input
func (h *Handler) GetUser(ctx context.Context, input *UserIdInput) (*User, error) {
    user, ok := h.users[input.Id]
    if !ok {
        return nil, httpserver.NewErrorWithStatus(http.StatusNotFound, errors.New("user not found"))
    }
    return user, nil
}

// With input and raw request access
func (h *Handler) Upload(ctx context.Context, req *http.Request, input *UploadInput) (map[string]string, error) {
    contentType := req.Header.Get("Content-Type")
    // input is populated from the request
    return map[string]string{"contentType": contentType}, nil
}

// Without input
func (h *Handler) Health(ctx context.Context) (map[string]string, error) {
    // no request data needed
    return map[string]string{"status": "ok"}, nil
}

// Without input, but with raw request access
func (h *Handler) Ping(ctx context.Context, req *http.Request) (map[string]string, error) {
    userAgent := req.Header.Get("User-Agent")
    // no input struct is bound
    return map[string]string{"userAgent": userAgent}, nil
}
```

You can register multiple handler groups on the same router, each with its own constructor:

```go
router.Group("/api/users").HandleWith(httpserver.With(NewUserHandler, func(r *httpserver.Router, h *UserHandler) {
    r.GET("", httpserver.Bind(h.ListUsers))
    r.POST("", httpserver.Bind(h.CreateUser))
}))

router.Group("/api/orders").HandleWith(httpserver.With(NewOrderHandler, func(r *httpserver.Router, h *OrderHandler) {
    r.GET("", httpserver.Bind(h.ListOrders))
}))
```

## Binding request data

`Bind` automatically populates input structs from the incoming request using struct tags.

### URI parameters

Use the `uri` tag to bind path parameters. Define them in your route with `:name` syntax:

```go
type UserIdInput struct {
    Id int `uri:"id" binding:"required"`
}

router.GET("/users/:id", httpserver.Bind(h.GetUser))
```

### Query string and form parameters

Use the `form` tag for query string parameters (GET) or form-encoded bodies (POST):

```go
type ListUsersInput struct {
    Role   string `form:"role"`
    Limit  int    `form:"limit"`
    Offset int    `form:"offset"`
}
```

### JSON request body

Use the `json` tag to bind from a JSON body. The content type `application/json` is auto-detected:

```go
type CreateUserInput struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Role  string `json:"role" binding:"omitempty,oneof=admin user guest"`
}
```

### HTTP headers

Use the `header` tag to bind from request headers:

```go
type AuthenticatedInput struct {
    RequestId string `header:"X-Request-Id"`
    UserAgent string `header:"User-Agent"`
}
```

### Combining multiple sources

Combine tags to pull from multiple sources in a single input struct:

```go
type ListFilesInput struct {
    Database   string            `uri:"database"`
    Table      string            `uri:"table"`
    Partitions map[string]string `json:"partitions" form:"partitions"`
    RequestId  string            `header:"X-Request-Id"`
}
```

URI parameters and headers are always bound when their tags are present. Body/query binding depends on the request's Content-Type header.

### Auto-detection logic

`Bind` examines the input type's struct tags and the request Content-Type to determine how to bind:

The `uri` and `header` tags are inferred from the input type and are not selected by Content-Type.

| Content-Type | Tag used |
|---|---|
| `application/json` | `json` |
| `application/xml` | `xml` |
| `application/x-yaml` | `yaml` |
| `application/x-protobuf` | `protobuf` |
| `application/x-msgpack` | `msgpack` |
| `application/x-www-form-urlencoded` | `form` |
| `multipart/form-data` | `form` |
| `text/plain` | `plain` |

### Validation

The `binding` tag uses [go-playground/validator](https://pkg.go.dev/github.com/go-playground/validator/v10). If binding or validation fails, the request is rejected with `400 Bad Request` and a descriptive client error. Validation bind failures are also logged at warning level, not error level, with field, tag, invalid value, and parameter details:

| Tag | Description |
|---|---|
| `required` | Field must be present and non-zero |
| `email` | Must be a valid email |
| `oneof=a b c` | Must be one of the listed values |
| `gte=0` | Must be greater than or equal to 0 |
| `omitempty` | Skip validation if field is empty |

### BindR — access the raw request

If you need the raw `*http.Request` alongside your input struct:

```go
func (h *Handler) Upload(ctx context.Context, req *http.Request, input *UploadInput) (map[string]string, error) {
    contentType := req.Header.Get("Content-Type")
    return map[string]string{"contentType": contentType}, nil
}
```

### Resolve the client IP

When your handler needs the caller IP address, use `ResolveClientIP` with the raw request from `BindR` or `BindNR`:

```go
func (h *Handler) GetProfile(ctx context.Context, req *http.Request, input *GetProfileInput) (map[string]string, error) {
    clientIP, err := httpserver.ResolveClientIP(req)
    if err != nil {
        return nil, err
    }

    return map[string]string{"clientIP": clientIP}, nil
}
```

`ResolveClientIP` checks `X-Forwarded-For` first, then `X-Real-IP`, and falls back to `req.RemoteAddr`. Use it when your service runs behind a trusted proxy or load balancer that sets these headers.

## Sending responses

Typed handlers return Go values. The default response negotiator encodes these values as JSON. If the request has no `Accept` header, httpserver selects JSON. The default negotiator supports only JSON. If `Accept` does not allow JSON, httpserver returns `406 Not Acceptable`.

### JSON by default

Return a Go value from a handler. httpserver encodes it as JSON:

```go
func (h *Handler) Health(ctx context.Context) (map[string]string, error) {
    return map[string]string{"status": "ok"}, nil
}
```

Send a compatible `Accept` header to get JSON:

```bash
curl -i -H "Accept: application/json" http://localhost:8088/api/users/health
# HTTP/1.1 200 OK
# {"status":"ok"}
```

The default negotiator does not support plain text. This request returns `406 Not Acceptable`:

```bash
curl -i -H "Accept: text/plain" http://localhost:8088/api/users/health
# HTTP/1.1 406 Not Acceptable
```

To support XML or another format, create a negotiator with the representations that the server can produce:

```go
negotiator, err := httpserver.NewContentNegotiator(
    httpserver.ContentTypeApplicationJson,
    httpserver.JSONRepresentation(),
    httpserver.XMLRepresentation(),
)
if err != nil {
    return err
}
```

Configure the negotiator for the whole server with `WithResponseNegotiator`. Pass this server option to `NewServer` or `NewServerWithSettings`. When using the application helper, put it in `ServerDefinition.Options` passed to `RunServers`:

```go
httpserver.RunServers(map[string]httpserver.ServerDefinition{
    "default": {
        RouterFactory: func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
            type HelloResponse struct {
                Message string `json:"message" xml:"message"`
            }
            router.GET("/hello", httpserver.BindN(func(ctx context.Context) (HelloResponse, error) {
                return HelloResponse{Message: "Hello, World!"}, nil
            }))
            return nil
        },
        Options: []httpserver.ServerOption{
            httpserver.WithResponseNegotiator(negotiator),
        },
    },
})
```

In this example, JSON is the default and XML is available. XML encoding requires output types that Go's `encoding/xml` package supports.

Use `ResponseNegotiationMiddleware` only when a router or group needs to override the server negotiator. For example, this group uses JSON only, while other routes retain the server's JSON and XML representations:

```go
group := router.Group("/json-only")
group.Use(httpserver.ResponseNegotiationMiddleware(
    httpserver.NewDefaultResponseNegotiator(),
))
```

Register the group's routes after installing the middleware. You do not need this middleware to configure the server-wide negotiator.

### Customizing responses

A typed output can implement `StatusCode() int` to set its HTTP status. The default status is `200 OK`. The complete example uses this for `201 Created` from `CreateUser`; `DeleteUser` uses an explicit bodyless response, described below.

```go
type CreatedUserOutput struct {
    *User
}

func (CreatedUserOutput) StatusCode() int {
    return http.StatusCreated
}
```

To add a header to `CreatedUserOutput`, implement `Header() http.Header`:

```go
func (CreatedUserOutput) Header() http.Header {
    header := make(http.Header)
    header.Set("X-User-Created", "true")
    return header
}
```

httpserver negotiates the typed output before it applies these methods. The selected representation controls `Content-Type`. A `Header` method cannot override `Content-Type`. A `ContentType() string` method by itself does not change the selected representation.

A full `httpserver.Response` bypasses negotiation. An explicit `Response` implements `ContentType() string`, `Body() ([]byte, error)`, `Header() http.Header`, and `StatusCode() int`. The server does not compare its content type with the request's `Accept` header.

For a bodyless response such as `204 No Content`, use `NewStatusResponse(http.StatusNoContent)`. A typed output with that status is negotiated first and can return `406 Not Acceptable` if `Accept` excludes supported representations; the explicit response preserves `204` regardless of `Accept`.

```go
return httpserver.NewStatusResponse(http.StatusNoContent), nil
```

This is specific to explicit responses; ordinary typed outputs, including those with custom status codes, remain subject to negotiation.

```go
func (h *Handler) PlainText(ctx context.Context) (httpserver.Response, error) {
    return httpserver.NewTextResponse("plain text"), nil
}
```

Use `NewJsonResponse`, `NewTextResponse`, or `NewStatusResponse` to create a full response. Each constructor bypasses negotiation. `NewTextResponse` creates a text response. `NewStatusResponse(http.StatusNoContent)` creates a no-content response.

Use response options to set a status, headers, or body:

```go
func explicitUserResponse(user *User) httpserver.Response {
    return httpserver.NewJsonResponse(
        user,
        httpserver.WithStatusCode(http.StatusCreated),
        httpserver.WithHeader("X-Custom-Header", "my-value"),
    )
}
```

| Option | Description |
|---|---|
| `WithStatusCode(code int)` | Set the HTTP status code. |
| `WithHeader(key, value string)` | Add one header. |
| `WithHeaders(headers http.Header)` | Merge multiple headers. |
| `WithBody(body []byte)` | Set the response body. |

### Error handling

Return a zero output and an error for unexpected failures. The error middleware returns `500 Internal Server Error` with `{"err":"internal server error"}`. This prevents internal details from reaching clients.

```go
func (h *Handler) Health(ctx context.Context) (map[string]string, error) {
    if len(h.users) > 10000 {
        return nil, fmt.Errorf("too many users")
    }
    return map[string]string{"status": "ok"}, nil
}
```

For a client error with a specific status code, return a zero output and an error wrapped with `NewErrorWithStatus`:

```go
func (h *Handler) GetUser(ctx context.Context, input *UserIdInput) (*User, error) {
    user, ok := h.users[input.Id]
    if !ok {
        return nil, httpserver.NewErrorWithStatus(http.StatusNotFound, errors.New("user not found"))
    }
    return user, nil
}
```

The error middleware maps both types of errors to HTTP responses. Use `NewErrorWithStatus` for expected client errors and return unwrapped errors for unexpected failures.

If middleware or lower-level code attaches an error to the Gin context and needs a specific status code, wrap it with `NewErrorWithStatus`:

```go
ginCtx.Error(httpserver.NewErrorWithStatus(http.StatusBadRequest, err))
```

## Complete example


<details>
<summary>main.go</summary>

```go title="main.go" lineNumbers
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gosoline-project/httpserver"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

func main() {
	httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
		router.Group("/api/users").HandleWith(httpserver.With(NewHandler, func(r *httpserver.Router, h *Handler) {
			r.GET("", httpserver.Bind(h.ListUsers))
			r.POST("", httpserver.Bind(h.CreateUser))
			r.GET("/:id", httpserver.Bind(h.GetUser))
			r.DELETE("/:id", httpserver.Bind(h.DeleteUser))
			r.GET("/health", httpserver.BindN(h.Health))
		}))

		return nil
	})
}

type ListUsersInput struct {
	Role   string `form:"role"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type CreateUserInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"omitempty,oneof=admin user guest"`
}

type UserIdInput struct {
	Id int `uri:"id" binding:"required"`
}

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CreatedUserOutput struct {
	*User
}

func (CreatedUserOutput) StatusCode() int {
	return http.StatusCreated
}

type Handler struct {
	users map[int]*User
	next  int
}

func NewHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*Handler, error) {
	return &Handler{
		users: map[int]*User{},
		next:  1,
	}, nil
}

func (h *Handler) ListUsers(ctx context.Context, input *ListUsersInput) ([]*User, error) {
	var result []*User
	for _, u := range h.users {
		if input.Role != "" && u.Role != input.Role {
			continue
		}
		result = append(result, u)
		if input.Limit > 0 && len(result) >= input.Limit {
			break
		}
	}
	return result, nil
}

func (h *Handler) CreateUser(ctx context.Context, input *CreateUserInput) (CreatedUserOutput, error) {
	user := &User{
		Id:    h.next,
		Name:  input.Name,
		Email: input.Email,
		Role:  input.Role,
	}
	h.users[user.Id] = user
	h.next++

	return CreatedUserOutput{User: user}, nil
}

func (h *Handler) GetUser(ctx context.Context, input *UserIdInput) (*User, error) {
	user, ok := h.users[input.Id]
	if !ok {
		return nil, httpserver.NewErrorWithStatus(http.StatusNotFound, errors.New("user not found"))
	}
	return user, nil
}

func (h *Handler) DeleteUser(ctx context.Context, input *UserIdInput) (httpserver.Response, error) {
	delete(h.users, input.Id)
	return httpserver.NewStatusResponse(http.StatusNoContent), nil
}

func (h *Handler) Health(ctx context.Context) (map[string]string, error) {
	if len(h.users) > 10000 {
		return nil, fmt.Errorf("too many users")
	}
	return map[string]string{"status": "ok"}, nil
}
```

</details>

<details>
<summary>config.dist.yml</summary>

```yaml title="config.dist.yml" lineNumbers
app:
  env: dev
  name: build-service

httpserver:
  default:
    port: 8088
```

</details>

Test it:

```bash
# Create a user
curl -X POST http://localhost:8088/api/users/ \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","role":"admin"}'
# {"id":1,"name":"Alice","email":"alice@example.com","role":"admin"}

# List users
curl http://localhost:8088/api/users/
# [{"id":1,"name":"Alice","email":"alice@example.com","role":"admin"}]

# Get a user by ID
curl http://localhost:8088/api/users/1
# {"id":1,"name":"Alice","email":"alice@example.com","role":"admin"}

# Get a missing user
curl http://localhost:8088/api/users/99
# {"err":"user not found"}

# Delete a user; the explicit bodyless response preserves 204 for this unsupported representation
curl -X DELETE http://localhost:8088/api/users/1 -H "Accept: text/plain" -v
# HTTP/1.1 204 No Content

# The user was deleted
curl -i http://localhost:8088/api/users/1
# HTTP/1.1 404 Not Found
# {"err":"user not found"}

# Health check
curl http://localhost:8088/api/users/health
# {"status":"ok"}
```

## What's next?

- [Stream with Server-Sent Events](/docs/pr-6/how-to/http-server/stream-with-sse/) — push real-time updates to clients
- [Serve a frontend](/docs/pr-6/how-to/http-server/serve-a-frontend/) — embed and serve a SPA from your Go binary
- [Configure your server](/docs/pr-6/how-to/http-server/configure-your-server/) — timeouts, compression, multiple servers
- [Add middleware](/docs/pr-6/how-to/http-server/add-middleware/) — custom middleware, CORS
- [Real-world example](/docs/pr-6/how-to/http-server/real-world-example/) — patterns from production applications
