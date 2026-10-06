---
sidebar:
  order: 2
title: HTTP Server Package
---

This guide explains how to migrate from the old `github.com/justtrackio/gosoline/pkg/httpserver` package directly to the standalone `github.com/gosoline-project/httpserver` module at `v0.6.4`.

The new package keeps the Gin-based server lifecycle and middleware model, but it changes the application-facing API. Instead of building and returning a `*httpserver.Definitions` tree, you receive a `*httpserver.Router` and register routes on it directly. Instead of implementing handler interfaces and manually unpacking an `httpserver.Request`, handlers are plain functions or methods registered through `Bind`, `BindN`, `BindR`, or SSE binding helpers.

## 1. Update the Dependency

Add the standalone module:

```bash
go get github.com/gosoline-project/httpserver@v0.6.4
```

Update imports:

```go
// Old
import "github.com/justtrackio/gosoline/pkg/httpserver"

// New
import "github.com/gosoline-project/httpserver"
```

Keep gosoline imports for application, config, logging, and other framework packages:

```go
import (
    "github.com/gosoline-project/httpserver"
    "github.com/justtrackio/gosoline/pkg/application"
    "github.com/justtrackio/gosoline/pkg/cfg"
    "github.com/justtrackio/gosoline/pkg/log"
)
```

## 2. Migrate Server Startup

### Old

The old package commonly used a `Definer` returning route definitions:

```go
func main() {
    application.RunHttpDefaultServer(DefineRouter)
}

func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger) (*httpserver.Definitions, error) {
    definitions := &httpserver.Definitions{}
    definitions.GET("/api/users", httpserver.CreateHandler(handler))

    return definitions, nil
}
```

### New

The new package uses a `RouterFactory`. The router is passed in and your function mutates it:

```go
func main() {
    httpserver.RunDefaultServer(DefineRouter)
}

func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    router.GET("/api/users", listUsers)

    return nil
}
```

If your application already uses `application.Run` with explicit modules, register the HTTP server as a module factory:

```go
func main() {
    application.Run(
        application.WithModuleFactory("http", httpserver.NewServer("default", DefineRouter)),
    )
}
```

## 3. Migrate Route Definitions

### Old

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger) (*httpserver.Definitions, error) {
    definitions := &httpserver.Definitions{}
    api := definitions.Group("/api")

    api.Use(authMiddleware)
    api.GET("/users", httpserver.CreateHandler(listUsersHandler))
    api.GET("/users/:id", httpserver.CreateHandler(getUserHandler))
    api.POST("/users", httpserver.CreateJsonHandler(createUserHandler))

    return definitions, nil
}
```

### New

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    api := router.Group("/api")

    api.Use(authMiddleware)
    api.GET("/users", httpserver.BindN(listUsers))
    api.GET("/users/:id", httpserver.Bind(getUser))
    api.POST("/users", httpserver.Bind(createUser))

    return nil
}
```

Route groups, HTTP methods, and Gin middleware still work the same way conceptually. The important change is that the root router is provided by the server, so you no longer allocate and return `&httpserver.Definitions{}`.

## 4. Use the `With` Pattern for Handler Dependencies

For handlers that need dependencies, use `httpserver.With`. The handler factory receives `ctx`, `config`, and `logger`, and the registration function wires methods to routes.

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    router.Group("/api/users").HandleWith(httpserver.With(NewUserHandler, func(r *httpserver.Router, h *UserHandler) {
        r.GET("", httpserver.BindN(h.ListUsers))
        r.GET("/:id", httpserver.Bind(h.GetUser))
        r.POST("", httpserver.Bind(h.CreateUser))
        r.DELETE("/:id", httpserver.Bind(h.DeleteUser))
    }))

    return nil
}

type UserHandler struct {
    store UserStore
}

func NewUserHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*UserHandler, error) {
    store, err := NewUserStore(ctx, config, logger)
    if err != nil {
        return nil, err
    }

    return &UserHandler{store: store}, nil
}
```

This replaces the old pattern of constructing handler interface implementations manually before adding routes to `Definitions`.

### From Route-Specific Handlers to Route-Group Handlers

With the old package, applications commonly used one handler instance per route. Each route-specific handler owned its own `Handle` method and was constructed before route registration:

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger) (*httpserver.Definitions, error) {
    definitions := &httpserver.Definitions{}
    definitions.Use(authMiddleware)

    listUsersHandler, err := NewListUsersHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create list users handler: %w", err)
    }

    getUserHandler, err := NewGetUserHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create get user handler: %w", err)
    }

    definitions.GET("/api/users", httpserver.CreateHandler(listUsersHandler))
    definitions.GET("/api/users/:id", httpserver.CreateUriHandler(getUserHandler))

    return definitions, nil
}
```

The new package makes it natural to group related routes behind one handler type. The handler is constructed once, shared dependencies are initialized once, and each route maps to a method:

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    router.Group("/api/users").HandleWith(httpserver.With(NewUserHandler, func(r *httpserver.Router, h *UserHandler) {
        r.GET("", httpserver.BindN(h.ListUsers))
        r.GET("/:id", httpserver.Bind(h.GetUser))
    }))

    return nil
}
```

This consolidation is not required for every migration, but it is usually the cleaner target. Keep separate handlers when routes truly have unrelated dependencies or lifecycle needs. Otherwise, prefer one handler per domain or route group, with one method per endpoint.

## 5. Migrate Handler Signatures

### Handler without Request Input

Old handlers often implemented `HandlerWithoutInput`:

```go
type listUsersHandler struct {
    store UserStore
}

func (h *listUsersHandler) Handle(ctx context.Context, request *httpserver.Request) (*httpserver.Response, error) {
    users, err := h.store.List(ctx)
    if err != nil {
        return nil, err
    }

    return httpserver.NewJsonResponse(users), nil
}
```

New handlers without request input use `BindN` and return a typed result:

```go
func (h *UserHandler) ListUsers(ctx context.Context) ([]User, error) {
    users, err := h.store.List(ctx)
    if err != nil {
        return nil, err
    }

    return users, nil
}
```

Register it with:

```go
r.GET("", httpserver.BindN(h.ListUsers))
```

### Handler with Input

Old input binding required `GetInput` and one of the `Create*Handler` helpers:

```go
type getUserHandler struct {
    store UserStore
}

type GetUserInput struct {
    Id uint `uri:"id" binding:"required"`
}

func (h *getUserHandler) GetInput() any {
    return &GetUserInput{}
}

func (h *getUserHandler) Handle(ctx context.Context, request *httpserver.Request) (*httpserver.Response, error) {
    input := request.Body.(*GetUserInput)

    user, err := h.store.Get(ctx, input.Id)
    if err != nil {
        return nil, err
    }

    return httpserver.NewJsonResponse(user), nil
}
```

New handlers receive the typed input directly:

```go
type GetUserInput struct {
    Id uint `uri:"id" binding:"required"`
}

func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (User, error) {
    user, err := h.store.Get(ctx, input.Id)
    if err != nil {
        return User{}, err
    }

    return user, nil
}
```

Register it with:

```go
r.GET("/:id", httpserver.Bind(h.GetUser))
```

## 6. Migrate Request Binding

The old package selected a binding helper explicitly:

Bound handlers registered with `Bind`, `BindR`, `BindN`, or `BindNR` accept a typed result `O`. Return ordinary results directly. See the response migration section for explicit response cases.

| Old helper | Typical new migration |
|---|---|
| `CreateHandler` | `BindN` for no input, or raw Gin handler if you need `*gin.Context` |
| `CreateJsonHandler` | `Bind` with `json` tags |
| `CreateQueryHandler` | `Bind` with `form` tags |
| `CreateUriHandler` | `Bind` with `uri` tags |
| `CreateMultipleBindingsHandler` | `Bind` with multiple tags, or explicit binders if needed |
| `CreateRawHandler` | `BindNR` or `BindR` and read `req.Body` |
| `CreateReaderHandler` | `BindNR` or `BindR` and use `req.Body` |
| `CreateSseHandler` | `BindSse` or `BindSseN` |

The new `Bind` function inspects struct tags and request content type:

```go
type CreateUserInput struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type CreatedUser struct {
    User
}

func (CreatedUser) StatusCode() int {
    return http.StatusCreated
}

func (h *UserHandler) CreateUser(ctx context.Context, input *CreateUserInput) (CreatedUser, error) {
    user, err := h.store.Create(ctx, input.Name, input.Email)
    if err != nil {
        return CreatedUser{}, err
    }

    return CreatedUser{User: user}, nil
}
```

The typed output preserves the old `201 Created` status without forcing a JSON response.

You can combine URI, query, and body fields in one input struct:

```go
type UpdateUserInput struct {
    Id    uint   `uri:"id" binding:"required"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Force bool   `form:"force"`
}
```

Use `form` tags for query string parameters and form-encoded bodies.

## 7. Replace Request Parameter Helpers

The old package exposed helpers such as `GetStringFromRequest` and `GetUintFromRequest` for path parameters stored in `httpserver.Request.Params`.

### Old

```go
func (h *getUserHandler) Handle(ctx context.Context, request *httpserver.Request) (*httpserver.Response, error) {
    id, ok := httpserver.GetUintFromRequest(request, "id")
    if !ok {
        return httpserver.NewStatusResponse(http.StatusBadRequest), nil
    }

    user, err := h.store.Get(ctx, *id)
    if err != nil {
        return nil, err
    }

    return httpserver.NewJsonResponse(user), nil
}
```

### New

Prefer typed input structs with `uri` tags:

```go
type GetUserInput struct {
    Id uint `uri:"id" binding:"required"`
}

func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (User, error) {
    user, err := h.store.Get(ctx, input.Id)
    if err != nil {
        return User{}, err
    }

    return user, nil
}
```

If you need direct Gin access, register a regular Gin handler:

```go
router.GET("/api/users/:id", func(ginCtx *gin.Context) {
    id := ginCtx.Param("id")
    ginCtx.JSON(http.StatusOK, gin.H{"id": id})
})
```

## 8. Access Raw Requests and Bodies

The old `CreateRawHandler` read the body into `request.Body` as a string. The new package keeps this explicit by passing the raw `*http.Request` to your handler.

```go
func (h *UserHandler) ImportUsers(ctx context.Context, req *http.Request) (httpserver.Response, error) {
    body, err := io.ReadAll(req.Body)
    if err != nil {
        return nil, fmt.Errorf("could not read request body: %w", err)
    }

    if err := h.store.Import(ctx, body); err != nil {
        return nil, err
    }

    return httpserver.NewStatusResponse(http.StatusAccepted), nil
}
```

Register it with `BindNR` because it has no typed input but needs the request:

```go
r.POST("/import", httpserver.BindNR(h.ImportUsers))
```

For typed input plus raw request access, use `BindR`:

```go
func (h *UserHandler) UploadAvatar(ctx context.Context, req *http.Request, input *UploadAvatarInput) (httpserver.Response, error) {
    contentType := req.Header.Get("Content-Type")
    _ = contentType

    return httpserver.NewStatusResponse(http.StatusNoContent), nil
}
```

### Client IP

The old `httpserver.Request` exposed `request.ClientIp`. In the standalone package, use raw request access and resolve the client IP explicitly:

```go
func (h *UserHandler) GetUser(ctx context.Context, req *http.Request, input *GetUserInput) (User, error) {
    clientIP, err := httpserver.ResolveClientIP(req)
    if err != nil {
        return User{}, err
    }

    _ = clientIP

    user, err := h.store.Get(ctx, input.Id)
    if err != nil {
        return User{}, err
    }

    return user, nil
}
```

Register handlers that need both typed input and the raw request with `BindR`. Use `BindNR` when the handler needs the raw request but no typed input.

## 9. Migrate Responses

Bound handlers in `v0.6.4` accept typed result values. Return ordinary results directly, and the server negotiates their response format. The default negotiator supports JSON only. Register other representations explicitly. A missing `Accept` header selects JSON, while an unsupported `Accept` value returns `406 Not Acceptable`.

For example, a JSON handler returns the domain value:

```go
return user, nil
```

See [Customizing responses](/docs/pr-6/how-to/http-server/build-an-http-service#customizing-responses) for typed status or header methods and explicit `Response` overrides.

### Fixed response formats and bodyless responses

Keep an explicit response when the route must return a fixed media type or no body. Typed outputs are negotiated before their `StatusCode()` method is applied, so even a typed output with `204 No Content` can return `406 Not Acceptable` for an unsupported `Accept` value. An explicit `httpserver.Response` bypasses negotiation; use `NewStatusResponse(http.StatusNoContent)` to preserve `204` regardless of `Accept`. Ordinary typed outputs, including outputs with custom statuses, remain subject to negotiation.

Old low-level text response:

```go
return httpserver.NewResponse("ok", httpserver.ContentTypeText, http.StatusOK, make(http.Header)), nil
```

New fixed text response:

```go
return httpserver.NewTextResponse("ok"), nil
```

For a bodyless status such as `204 No Content`, use:

```go
return httpserver.NewStatusResponse(http.StatusNoContent), nil
```

For raw bytes with a required media type, keep `NewResponse` with options:

```go
return httpserver.NewResponse(
    httpserver.WithBody([]byte("accepted")),
    httpserver.WithHeader("Content-Type", "text/plain; charset=utf-8"),
    httpserver.WithStatusCode(http.StatusAccepted),
), nil
```

## 10. Migrate Error Handling

For unexpected errors, return the zero value for the typed result and the error:

```go
func (h *UserHandler) ListUsers(ctx context.Context) ([]User, error) {
    users, err := h.store.List(ctx)
    if err != nil {
        return nil, fmt.Errorf("could not list users: %w", err)
    }

    return users, nil
}
```

For an expected client error with a specific status, return the zero result and `NewErrorWithStatus`:

```go
func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (User, error) {
    user, err := h.store.Get(ctx, input.Id)
    if errors.Is(err, ErrUserNotFound) {
        return User{}, httpserver.NewErrorWithStatus(http.StatusNotFound, err)
    }
    if err != nil {
        return User{}, fmt.Errorf("could not get user: %w", err)
    }

    return user, nil
}
```

Configure `WithErrorHandler` as a server option. Pass it to `NewServer` or `NewServerWithSettings`, or put it in `ServerDefinition.Options` for `RunServers`.

The callback type is `func(statusCode int, err error) any`. Error middleware calls it after it selects the HTTP status from the last recorded error.

Status selection follows this order:

1. An `ErrorWithStatus`, such as `NewErrorWithStatus`, sets the status.
2. Otherwise, the first matching `WithErrorMapper` sets the status.
3. Otherwise, validation errors use `400 Bad Request`.
4. All other errors use `500 Internal Server Error`.

The endpoint still returns its error through the normal `(output, error)` result. This callback returns the body to send to the client, not another Go error. If it returns a body value such as the `ErrorBody` struct below, the framework keeps the selected HTTP status. With private 5xx privacy, middleware passes `internal server error` to the callback.

Return a struct, map, or other serializable Go value that does not implement `httpserver.Response` to let the server choose the response format from the request's `Accept` header. The configured negotiator serializes that value and sets `Content-Type`. JSON is the default. XML requires a negotiator that includes `XMLRepresentation()` and a value that `encoding/xml` can encode, such as this struct:

```go
type ErrorBody struct {
    Message string `json:"message" xml:"message"`
}

httpserver.WithErrorHandler(func(statusCode int, err error) any {
    return ErrorBody{Message: err.Error()}
})
```

In `v0.6.4`, error middleware falls back to JSON with the selected error status if it cannot negotiate or encode that body value. To always send a fixed format instead, return a `httpserver.Response`. It bypasses `Accept` negotiation and controls its own status and `Content-Type`. This example always sends JSON; `WithStatusCode(statusCode)` preserves the selected error status:

```go
httpserver.WithErrorHandler(func(statusCode int, err error) any {
    return httpserver.NewJsonResponse(
        ErrorBody{Message: err.Error()},
        httpserver.WithStatusCode(statusCode),
    )
})
```

In `v0.6.4`, the default error handler exposes 4xx error messages but sanitizes 5xx responses as `{"err":"internal server error"}`. Binding and validation failures from `Bind` and `BindSse` are client errors and return `400 Bad Request`.

The 5xx behavior is controlled by `httpserver.<name>.errors.privacy`. The default is `private`, which hides internal error details. Set it to `public` only when clients should receive the original internal error message:

```yaml
httpserver:
  default:
    errors:
      privacy: public
```

If middleware attaches an error to the Gin context and needs a non-500 status code, wrap it with `NewErrorWithStatus`:

```go
ginCtx.Error(httpserver.NewErrorWithStatus(http.StatusBadRequest, err))
```

## 11. Migrate Middleware

Gin middleware remains compatible:

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    router.Use(requestIdMiddleware)

    api := router.Group("/api")
    api.Use(authMiddleware)
    api.GET("/users", httpserver.BindN(listUsers))

    return nil
}
```

For middleware that needs configuration, a logger, or server settings, use `UseFactory`:

```go
router.UseFactory(func(ctx context.Context, config cfg.Config, logger log.Logger, settings *httpserver.Settings) (gin.HandlerFunc, error) {
    token := config.GetString("api_token")

    return func(ginCtx *gin.Context) {
        if ginCtx.GetHeader("Authorization") != "Bearer "+token {
            ginCtx.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        ginCtx.Next()
    }, nil
})
```

The `settings` argument contains the resolved server settings. `settings.Name` is the server name, such as `"default"` for `RunDefaultServer` or `"admin"` for `NewServer("admin", ...)`.

### CORS

Move the old global CORS settings below the named HTTP server:

| Old gosoline key | New standalone key for server `default` |
|---|---|
| `api_cors_allowed_origin_pattern` | `httpserver.default.cors.allowed_origin_pattern` |
| `api_cors_allowed_headers` | `httpserver.default.cors.allowed_headers` |
| `api_cors_allowed_methods` | `httpserver.default.cors.allowed_methods` |

```yaml
httpserver:
  default:
    cors:
      allowed_origin_pattern: ".*"
      allowed_headers:
        - Content-Type
        - Authorization
      allowed_methods:
        - GET
        - POST
        - PUT
        - DELETE
```

Prefer the settings-aware factory so the middleware reads the current server name automatically:

```go
router.UseFactory(httpserver.CorsFactory)
```

If you construct the middleware manually, pass the server name explicitly:

```go
corsMiddleware, err := httpserver.Cors(config, "default")
if err != nil {
    return err
}
router.Use(corsMiddleware)
```

The origin pattern is matched against the full `Origin` value. For example, `https://example\\.com` allows `https://example.com`, but not `https://example.com.evil.com`.

### Authentication

The standalone module includes auth helpers in `github.com/gosoline-project/httpserver/auth`. Use them when they match the old embedded package behavior.

Import the standalone auth package separately:

```go
import (
    "github.com/gosoline-project/httpserver"
    "github.com/gosoline-project/httpserver/auth"
)
```

Auth settings are now scoped to the named HTTP server. `httpserver.RunDefaultServer` uses the server name `default`, so auth config belongs below `httpserver.default.auth`. If you register `httpserver.NewServer("admin", DefineRouter)`, the auth settings belong below `httpserver.admin.auth`, and you pass `"admin"` to auth constructors.

Common config key migrations:

| Old gosoline key | New standalone key for server `default` |
|---|---|
| `api_auth_keys` | `httpserver.default.auth.keys` |
| `api_auth_basic_users` | `httpserver.default.auth.basic.users` |
| `api_auth_bearer_id_header` | `httpserver.default.auth.bearer.id_header` |
| `api_auth_bearer_token_header` | `httpserver.default.auth.bearer.token_header` |
| `httpserver.<name>.auth.jwt.*` | `httpserver.<name>.auth.jwt.*` |

### API Key Auth

Old application-owned middleware often looked like this:

```go
func ApiKeyMiddleware(expected string) gin.HandlerFunc {
    return func(ginCtx *gin.Context) {
        if ginCtx.Query("api_key") != expected {
            ginCtx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"err": "invalid authorization"})
            return
        }

        ginCtx.Next()
    }
}
```

Migrate to `auth.NewConfigKeyHandler` when the valid keys should come from config:

```yaml
httpserver:
  default:
    auth:
      keys:
        - ${env:API_KEY}
```

```go
apiKeyAuth, err := auth.NewConfigKeyHandler(config, logger, "default", auth.ProvideValueFromHeader(auth.HeaderApiKey))
if err != nil {
    return err
}

api := router.Group("/api")
api.Use(apiKeyAuth)
```

Use `auth.ProvideValueFromQueryParam("api_key")` or `auth.ProvideValueFromUriPath("apiKey")` if your old middleware read the key from a query parameter or path parameter instead of a header.

For reusable middleware registration that should use the current server name automatically, register the settings-aware factory:

```go
api := router.Group("/api")
api.UseFactory(auth.ConfigKeyHandlerFactory(auth.ProvideValueFromHeader(auth.HeaderApiKey)))
```

### Basic Auth

Move Basic Auth users below the named server:

```yaml
httpserver:
  default:
    auth:
      basic:
        users:
          - admin:${env:BASIC_AUTH_ADMIN_PASSWORD}
```

Register the middleware with the same server name:

```go
basicAuth, err := auth.NewBasicAuthHandler(config, logger, "default")
if err != nil {
    return err
}

router.Group("/admin").Use(basicAuth)
```

Or use the settings-aware factory:

```go
router.Group("/admin").UseFactory(auth.BasicAuthHandlerFactory)
```

Unauthorized Basic Auth responses use the configured gosoline app identity name as the Basic realm.

### Bearer Token Auth

Move bearer header configuration below the named server:

```yaml
httpserver:
  default:
    auth:
      bearer:
        id_header: X-BEARER-ID
        token_header: X-BEARER-TOKEN
```

Then pass the server name and your bearer provider:

```go
bearerAuth, err := auth.NewTokenBearerHandler(config, logger, "default", provider)
if err != nil {
    return err
}

router.Group("/api").Use(bearerAuth)
```

Or use the settings-aware factory:

```go
router.Group("/api").UseFactory(auth.TokenBearerHandlerFactory(provider))
```

### JWT Auth

JWT settings live below `httpserver.<name>.auth.jwt`:

```yaml
httpserver:
  default:
    auth:
      jwt:
        signingSecret: ${env:JWT_SIGNING_SECRET}
        issuer: my-service
        expireDuration: 15m
```

```go
jwtAuth, err := auth.NewJwtAuthHandler(config, "default")
if err != nil {
    return err
}

router.Group("/api").Use(jwtAuth)
```

Or use the settings-aware factory:

```go
router.Group("/api").UseFactory(auth.JwtAuthHandlerFactory)
```

The standalone JWT helper validates `Authorization: Bearer <token>` headers using HS256 and requires an `email` claim for the authenticated subject.

### Auth Subjects and Chains

Successful authenticators attach an `auth.Subject` to the request context. In bound handlers, retrieve it with `auth.GetSubject(ctx)`:

```go
func (h *UserHandler) Me(ctx context.Context) (*auth.Subject, error) {
    subject := auth.GetSubject(ctx)

    return subject, nil
}
```

If a route accepts multiple auth methods, build authenticators and combine them with `auth.NewChainHandler`. You can optionally restrict enabled methods per server with `httpserver.<name>.auth.allowedAuthenticators`.

```yaml
httpserver:
  default:
    auth:
      allowedAuthenticators:
        - apiKey
        - jwtAuth
```

```go
authenticators := map[string]auth.Authenticator{
    auth.ByApiKey: apiKeyAuth,
    auth.ByJWT:    jwtAuth,
}

authenticators, err = auth.OnlyConfiguredAuthenticators(config, "default", authenticators)
if err != nil {
    return err
}

router.Group("/api").Use(auth.NewChainHandler(authenticators))
```

The old Google auth helper was not ported to the standalone module. If you used `NewConfigGoogleHandler` or `NewConfigGoogleAuthenticator`, keep that logic in application-owned middleware or an application-owned auth package.

See [Authenticate requests](/docs/pr-6/how-to/http-server/authentication/) for the full standalone auth guide.

## 12. Review Configuration

The main server configuration still lives under `httpserver.<name>`:

```yaml
httpserver:
  default:
    port: 8080
    mode: release
    timeout:
      read: 60s
      write: 60s
      idle: 60s
      drain: 0s
      shutdown: 60s
    compression:
      level: default
      decompression: true
    cors:
      allowed_origin_pattern: ".*"
      allowed_headers:
        - Content-Type
        - Authorization
      allowed_methods:
        - GET
        - POST
        - PUT
        - DELETE
    errors:
      privacy: private
    max_body_bytes: 10485760
```

Keep these points in mind while migrating:

| Area | Notes |
|---|---|
| Server name | `RunDefaultServer` reads `httpserver.default`. `NewServer("admin", ...)` reads `httpserver.admin`. |
| Routes | `/health` is still registered by the server. Unhealthy modules are reported as `"unhealthy"`; underlying error strings are logged but not exposed in the response. |
| Compression | Exclude SSE endpoints from compression. |
| Request body limit | `max_body_bytes` defaults to `10485760` bytes (10 MiB). Set it to `0` to disable the limit or raise it for large upload endpoints. The limit is applied after request decompression. |
| Binding errors | `Bind` and `BindSse` binding or validation failures return `400 Bad Request`. |
| Error responses | The default handler exposes 4xx error messages and returns `{"err":"internal server error"}` for 5xx responses. Set `httpserver.<name>.errors.privacy` to `public` only when 5xx error details should be exposed. |
| CORS | Move old `api_cors_*` keys to `httpserver.<name>.cors.*`. `allowed_origin_pattern` is matched against the full `Origin` value, so partial regex matches do not allow origins. |
| Profiling | Profiling is still configured under `profiling` and binds to `127.0.0.1:<port>`. |
| Auth | Auth config is scoped below `httpserver.<name>.auth`; old global auth keys must be moved below the server name. |

## 13. Migrate Tests

Old test suites usually implemented definition-based setup:

```go
func (s *UserApiSuite) SetupApiDefinitions() httpserver.Definer {
    return DefineRouter
}
```

New httpserver test cases use router factories:

```go
func (s *UserApiSuite) SetupHttpServerRouter() httpserver.RouterFactory {
    return DefineRouter
}
```

Assert typed output values and domain behavior through handler methods. Test negotiation, headers, and status codes through HTTP requests.


For binding and routing tests, use the package test helpers or an application test case with `SetupHttpServerRouter`.

## 14. Complete Before and After Example

### Old

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger) (*httpserver.Definitions, error) {
    listUsersHandler, err := NewListUsersHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create list users handler: %w", err)
    }

    getUserHandler, err := NewGetUserHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create get user handler: %w", err)
    }

    createUserHandler, err := NewCreateUserHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create create user handler: %w", err)
    }

    deleteUserHandler, err := NewDeleteUserHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create delete user handler: %w", err)
    }

    definitions := &httpserver.Definitions{}
    users := definitions.Group("/api/users")

    users.GET("", httpserver.CreateHandler(listUsersHandler))
    users.GET("/:id", httpserver.CreateUriHandler(getUserHandler))
    users.POST("", httpserver.CreateJsonHandler(createUserHandler))
    users.DELETE("/:id", httpserver.CreateUriHandler(deleteUserHandler))

    return definitions, nil
}
```

In this style, each route has its own handler instance, even though all four routes belong to the same users API.

### New

```go
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    router.Group("/api/users").HandleWith(httpserver.With(NewUserHandler, func(r *httpserver.Router, h *UserHandler) {
        r.GET("", httpserver.BindN(h.ListUsers))
        r.GET("/:id", httpserver.Bind(h.GetUser))
        r.POST("", httpserver.Bind(h.CreateUser))
        r.DELETE("/:id", httpserver.Bind(h.DeleteUser))
    }))

    return nil
}

type GetUserInput struct {
    Id uint `uri:"id" binding:"required"`
}

type CreateUserInput struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

func (h *UserHandler) ListUsers(ctx context.Context) ([]User, error) {
    users, err := h.store.List(ctx)
    if err != nil {
        return nil, err
    }

    return users, nil
}

func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (User, error) {
    user, err := h.store.Get(ctx, input.Id)
    if errors.Is(err, ErrUserNotFound) {
        return User{}, httpserver.NewErrorWithStatus(http.StatusNotFound, err)
    }
    if err != nil {
        return User{}, err
    }

    return user, nil
}

type CreatedUser struct {
    User
}

func (CreatedUser) StatusCode() int {
    return http.StatusCreated
}

func (h *UserHandler) CreateUser(ctx context.Context, input *CreateUserInput) (CreatedUser, error) {
    user, err := h.store.Create(ctx, input.Name, input.Email)
    if err != nil {
        return CreatedUser{}, err
    }

    return CreatedUser{User: user}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, input *GetUserInput) (httpserver.Response, error) {
    if err := h.store.Delete(ctx, input.Id); err != nil {
        return nil, err
    }

    return httpserver.NewStatusResponse(http.StatusNoContent), nil
}
```

## Migration Checklist

- Replace `github.com/justtrackio/gosoline/pkg/httpserver` imports with `github.com/gosoline-project/httpserver`.
- Replace `Definer` functions returning `*Definitions` with `RouterFactory` functions accepting `*Router`.
- Replace `application.RunHttpDefaultServer` with `httpserver.RunDefaultServer`, or register `httpserver.NewServer` as a module factory.
- Replace `Definitions` route registration with direct `Router` registration.
- Replace old handler interfaces with plain methods registered through `Bind`, `BindN`, `BindR`, or `BindNR`.
- Consolidate route-specific handlers into route-group or domain handlers where the routes share dependencies.
- Replace `GetInput` and `request.Body.(*Input)` with typed input arguments.
- Replace path parameter helpers with `uri` tags on input structs.
- Replace `CreateRawHandler` and `CreateReaderHandler` with `BindR` or `BindNR` and explicit `req.Body` handling.
- Replace `request.ClientIp` with `ResolveClientIP(req)` in `BindR` or `BindNR` handlers.
- Return typed values directly from ordinary bound handlers. Keep explicit responses for fixed media types, raw bodies, and bodyless status responses. See [Customizing responses](/docs/pr-6/how-to/http-server/build-an-http-service#customizing-responses).
- Review `max_body_bytes`; keep the v0.6.4 default 10 MiB limit, raise it, or set it to `0` intentionally.
- Update tests for v0.6.4 behavior: binding errors return 400, default/private 5xx bodies are sanitized, `errors.privacy: public` exposes 5xx messages, CORS patterns match full origins, health responses hide module error details, and profiling binds to loopback.
- Update test suites from definition setup to router factory setup.
- Run your HTTP tests and exercise validation errors, client errors, and middleware behavior after migration.

## AI Agent Migration Instructions

If you want an AI agent to perform this migration, copy the prepared migration instructions and provide them as the task prompt. The instructions are intentionally operational and include discovery steps, migration order, API mappings, validation checks, and stop conditions.

[Download the migration instructions](/docs/pr-6/downloads/httpserver-migration-agent-instructions.txt). Use the code block’s copy button to copy the full prompt.

```text title="AI migration instructions"
Migrate the application from the old embedded package github.com/justtrackio/gosoline/pkg/httpserver directly to the standalone package github.com/gosoline-project/httpserver at v0.6.4, while preserving intended HTTP behavior, routes, middleware, response codes, request binding, and tests.

Scope discovery:

- Find all imports of github.com/justtrackio/gosoline/pkg/httpserver.
- Find all imports of github.com/justtrackio/gosoline/pkg/httpserver/auth.
- Find all route definer functions with this signature: func(context.Context, cfg.Config, log.Logger) (*httpserver.Definitions, error).
- Find all application.RunHttpDefaultServer(...) calls.
- Find all httpserver.New(...) or httpserver.NewWithSettings(...) module registrations from the old package.
- Find all &httpserver.Definitions{} allocations and Definitions.Group, Definitions.Use, GET, POST, PUT, PATCH, DELETE, OPTIONS, or Handle calls.
- Find all old handler helpers: CreateHandler, CreateJsonHandler, CreateQueryHandler, CreateUriHandler, CreateMultipleBindingsHandler, CreateRawHandler, CreateReaderHandler, CreateSseHandler, CreateStreamHandler, and CreateDownloadHandler.
- Find all old handler interfaces: HandlerWithoutInput, HandlerWithInput, HandlerWithMultipleBindings, and HandlerWithStream.
- Find all uses of httpserver.Request, request.Body, request.Params, request.Url, request.Header, request.Cookies, request.Method, and request.ClientIp.
- Find all uses of GetStringFromRequest and GetUintFromRequest.
- Find all uses of old response constructors and fields, especially NewResponse(body, contentType, statusCode, header), ContentTypeText, ContentTypeJson, ContentTypeHtml, ContentTypeProtobuf, Response.AddHeader, Response.WithBody, and Response.WithContentType.
- Find all HTTP server test setup helpers that return an old httpserver.Definer or implement SetupApiDefinitions.
- Find endpoints that accept large request bodies and check whether they need a v0.6.4 max_body_bytes override.
- Find tests or clients that assert internal error response bodies, binding error status codes, CORS partial-origin matching, health-check error details, or profiling listen addresses.
- Find all router.UseFactory calls and custom httpserver.MiddlewareFactory implementations, because v0.6.4 middleware factories receive server settings as a fourth argument.
- Find any environment-specific need to expose internal 5xx error details, which now requires httpserver.<name>.errors.privacy: public.
- Find old global CORS config keys: api_cors_allowed_origin_pattern, api_cors_allowed_headers, and api_cors_allowed_methods.

Do not start with a global import replacement. First classify each usage, because the old package also contained auth and helper APIs that do not map one-to-one to the standalone package.

Migration order:

1. Add the new dependency with: go get github.com/gosoline-project/httpserver@v0.6.4
2. Migrate server startup and router factories.
3. Migrate route definitions from returned Definitions to mutating *Router.
4. Migrate middleware registration.
5. Migrate auth middleware and auth configuration.
6. Migrate handlers route group by route group.
7. Consolidate old route-specific handlers into domain or route-group handlers where they share dependencies.
8. Migrate request binding and raw request access.
9. Migrate responses and handled errors.
10. Review v0.6.4 runtime behavior: request body limits, binding errors, configurable error privacy, CORS origin matching, health responses, and profiling bind address.
11. Migrate tests.
12. Run formatting and tests.

Keep each step small enough that the codebase remains understandable after every edit.

Server startup transformation:

- Replace application.RunHttpDefaultServer(DefineRouter, options...) with httpserver.RunDefaultServer(DefineRouter, options...).
- Replace old definer signatures from func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger) (*httpserver.Definitions, error) to func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error.
- Replace application.WithModuleFactory("http", httpserver.New("default", DefineRouter)) with application.WithModuleFactory("http", httpserver.NewServer("default", DefineRouter)).
- If the application wraps the old httpserver.Definer in helper functions, update those helper types to accept or return httpserver.RouterFactory.

Route definition transformation:

- Replace allocations of &httpserver.Definitions{} with use of the router argument.
- Preserve all route paths, methods, group prefixes, and middleware order.
- Do not normalize or redesign paths during the migration.
- Convert definitions.Use(...) to router.Use(...).
- Convert definitions.Group("/api") to router.Group("/api").
- Convert returned definitions to return nil.
- Convert error returns from return nil, err to return err in router factories.
- Convert custom UseFactory callbacks from func(ctx, config, logger) (gin.HandlerFunc, error) to func(ctx, config, logger, settings *httpserver.Settings) (gin.HandlerFunc, error).
- Use settings.Name inside middleware factories when middleware needs the current server name.
- For package CORS middleware, prefer router.UseFactory(httpserver.CorsFactory). If constructing it manually, call httpserver.Cors(config, "<serverName>").

Handler consolidation instructions:

- Old code commonly creates one handler per route. When those handlers share dependencies or belong to the same domain, consolidate them into one route-group handler.
- Prefer one handler type per cohesive route group, with one method per endpoint.
- Register route-group handlers with router.Group(...).HandleWith(httpserver.With(NewHandler, func(r *httpserver.Router, h *Handler) { ... })).
- Only keep separate handler types if their dependencies, lifecycle, or domain responsibilities are genuinely different.

Handler signature transformations:

- CreateHandler(handler) where handler.Handle(ctx, request) ignores input becomes BindN(h.Method) with func(ctx context.Context) (O, error).
- CreateJsonHandler(handler) becomes Bind(h.Method) with func(ctx context.Context, input *Input) (O, error) and json tags on the input struct.
- CreateQueryHandler(handler) becomes Bind(h.Method) with func(ctx context.Context, input *Input) (O, error) and form tags on the input struct.
- CreateUriHandler(handler) becomes Bind(h.Method) with func(ctx context.Context, input *Input) (O, error) and uri tags on the input struct.
- CreateMultipleBindingsHandler(handler) becomes Bind(h.Method) with func(ctx context.Context, input *Input) (O, error) and combined tags, or explicit Gin binders if necessary.
- CreateRawHandler(handler) becomes BindNR(h.Method) with func(ctx context.Context, req *http.Request) (O, error) and explicit req.Body reading.
- CreateReaderHandler(handler) becomes BindNR(h.Method) with func(ctx context.Context, req *http.Request) (O, error) and direct req.Body use.
- CreateSseHandler(handler) becomes BindSse or BindSseN.
The ordinary result type O can be any Go value. Return it directly. See the response migration rules and the dedicated response customization section below.

Replace old GetInput methods with input struct arguments. Old code that returns &Input{} from GetInput and then reads request.Body.(*Input) should become a method that directly accepts input *Input.

Request data migration rules:

- request.Body.(*Input) becomes handler input argument input *Input.
- GetStringFromRequest(request, "id") becomes a string field tagged `uri:"id" binding:"required"`.
- GetUintFromRequest(request, "id") becomes a uint field tagged `uri:"id" binding:"required"`.
- request.Url.RawQuery becomes req.URL.RawQuery via BindR or BindNR.
- request.Url.Path becomes req.URL.Path via BindR or BindNR.
- request.Method becomes req.Method via BindR or BindNR.
- request.Header becomes req.Header via BindR or BindNR.
- request.Cookies becomes req.Cookies() via BindR or BindNR.
- request.ClientIp becomes httpserver.ResolveClientIP(req) via BindR or BindNR.

Prefer typed binding over manual request inspection. Use BindR or BindNR only when the handler truly needs the raw *http.Request, for example to read the body directly or resolve the client IP.

In v0.6.4, Bind and BindSse binding or validation failures are client errors and return 400 Bad Request through the default error middleware. Update tests that expected 500 for malformed input.

Response migration rules:

- Bind[I, O] uses func(ctx context.Context, input *I) (O, error). BindR[I, O] uses func(ctx context.Context, req *http.Request, input *I) (O, error). BindN[O] uses func(ctx context.Context) (O, error). BindNR[O] uses func(ctx context.Context, req *http.Request) (O, error).
- A typed result can implement StatusCode() int or Header() http.Header. See /how-to/http-server/build-an-http-service#customizing-responses for status and header rules.
- NewResponse("text", ContentTypeText, code, header) becomes NewTextResponse(text, WithStatusCode(code), WithHeaders(header)).
- NewResponse([]byte(...), contentType, code, header) becomes NewResponse(WithBody(...), WithHeader("Content-Type", contentType), WithStatusCode(code), WithHeaders(header)).
- A typed result that is a httpserver.Response bypasses response negotiation. Use it only when the handler must control the response directly.
- Keep NewTextResponse for a fixed text representation, NewStatusResponse for bodyless responses that must bypass negotiation (for example, to preserve 204 for an unsupported Accept value), and NewResponse options for raw bytes or a required media type.
- Do not use Header() to replace the negotiated Content-Type. The response negotiator sets that header.
- The Response interface exposes ContentType(), Body(), Header(), and StatusCode(). Do not access fields such as resp.StatusCode or resp.Header in tests.
- For errors, return the zero value for O and the error. Use NewErrorWithStatus(status, err) when the error must set an explicit status.
- Custom error-body callbacks use func(statusCode int, err error) any. Return ordinary values for negotiation or a full Response to bypass it.
- The v0.6.4 default error handler returns the actual error message for 4xx responses and returns {"err":"internal server error"} for 5xx responses.
- The 5xx behavior is controlled by httpserver.<name>.errors.privacy. The default private value hides internal details. Set privacy to public only when exposing internal 5xx messages is intentional.

Auth migration rules:

- Replace old auth imports from github.com/justtrackio/gosoline/pkg/httpserver/auth with github.com/gosoline-project/httpserver/auth when the old usage maps to a standalone auth helper.
- Import the standalone auth package separately from github.com/gosoline-project/httpserver.
- Auth config is scoped to the HTTP server name under httpserver.<name>.auth. RunDefaultServer uses the name default. NewServer("admin", ...) uses the name admin.
- Pass the same server name to auth constructors that read config, for example auth.NewConfigKeyHandler(config, logger, "default", provider), auth.NewBasicAuthHandler(config, logger, "default"), auth.NewTokenBearerHandler(config, logger, "default", provider), and auth.NewJwtAuthHandler(config, "default").
- Migrate old api_auth_keys to httpserver.<name>.auth.keys.
- Migrate old api_auth_basic_users to httpserver.<name>.auth.basic.users.
- Migrate old api_auth_bearer_id_header to httpserver.<name>.auth.bearer.id_header.
- Migrate old api_auth_bearer_token_header to httpserver.<name>.auth.bearer.token_header.
- Keep JWT settings under httpserver.<name>.auth.jwt, and ensure signingSecret, issuer, and expireDuration match the application behavior.
- Migrate configured API-key auth to auth.NewConfigKeyHandler or auth.NewConfigKeyAuthenticator with an appropriate provider: ProvideValueFromHeader, ProvideValueFromQueryParam, or ProvideValueFromUriPath.
- Migrate Basic Auth to auth.NewBasicAuthHandler or auth.NewBasicAuthAuthenticator.
- Migrate bearer-token auth to auth.NewTokenBearerHandler or auth.NewTokenBearerAuthenticator and preserve the provider's token lookup behavior.
- Migrate JWT auth to auth.NewJwtAuthHandler or auth.NewJWTAuthAuthenticator. The standalone JWT authenticator validates Authorization: Bearer <token> and requires an email claim for the auth subject.
- When using router.UseFactory, prefer the settings-aware auth middleware factories so the current server name is used automatically: auth.ConfigKeyHandlerFactory(provider), auth.BasicAuthHandlerFactory, auth.TokenBearerHandlerFactory(provider), and auth.JwtAuthHandlerFactory.
- Use auth.NewChainHandler for routes that accept multiple auth methods. Use auth.OnlyConfiguredAuthenticators with httpserver.<name>.auth.allowedAuthenticators if the enabled methods are configurable.
- Successful standalone authenticators attach an auth.Subject to the request context. If handlers read old auth data from context.Context, update them to use auth.GetSubject(ctx), or preserve the old context values in application-owned middleware if downstream code requires them.
- Preserve accepted credentials, rejection status codes, response bodies, and context values used by downstream handlers.
- Basic Auth unauthorized responses use the configured gosoline app identity name as the Basic realm.
- The old Google auth helper was not ported to the standalone module. If the application uses NewConfigGoogleHandler or NewConfigGoogleAuthenticator, keep that logic in application-owned Gin middleware or an application-owned auth package.
- Stop and ask for guidance if the old auth logic depends on package internals that are not obvious from the application code or if Google auth behavior cannot be preserved safely.

CORS migration rules:

- Move api_cors_allowed_origin_pattern to httpserver.<name>.cors.allowed_origin_pattern.
- Move api_cors_allowed_headers to httpserver.<name>.cors.allowed_headers.
- Move api_cors_allowed_methods to httpserver.<name>.cors.allowed_methods.
- RunDefaultServer uses the server name default. NewServer("admin", ...) uses the server name admin.
- Prefer router.UseFactory(httpserver.CorsFactory) so the current server name is used automatically.
- If constructing the middleware manually, pass the server name explicitly: httpserver.Cors(config, "default").
- CORS allowed_origin_pattern is matched against the full Origin value. Do not rely on partial regex matches; use an explicit pattern that matches the complete allowed origin set.

Test migration rules:

- Replace SetupApiDefinitions() httpserver.Definer with SetupHttpServerRouter() httpserver.RouterFactory.
- Instantiate the new route-group handler directly in handler unit tests.
- Call handler methods directly for business logic tests.
- Assert response.StatusCode(), response.Header(), and response.Body() instead of old response fields.
- Keep integration or routing tests for binding, middleware, and path behavior.

Configuration and runtime behavior rules for v0.6.4:

- The main server configuration lives under httpserver.<name>. RunDefaultServer uses the name default.
- The server adds /health automatically. Unhealthy modules are reported as "unhealthy"; underlying error messages are logged but not returned in the response body.
- Incoming request bodies are limited by httpserver.<name>.max_body_bytes. The default is 10485760 bytes (10 MiB). Set it to 0 to disable the limit, or raise it when the old service intentionally accepted larger bodies.
- The body limit is applied after decompression, so compressed uploads are limited by their decompressed size.
- Internal 5xx error details are hidden by default and with httpserver.<name>.errors.privacy: private. Use privacy: public only when clients should see internal 5xx messages.
- CORS settings live under httpserver.<name>.cors. The allowed_origin_pattern is matched against the full Origin value. Do not rely on partial regex matches; use an explicit pattern that matches the complete allowed origin set.
- Profiling is configured under profiling and, when enabled, listens on 127.0.0.1:<profiling.api.port>.
- Keep compression exclusions for SSE endpoints; gzip buffering breaks real-time SSE streams.

Validation checklist:

- No application code still imports github.com/justtrackio/gosoline/pkg/httpserver, unless intentionally left for a separate migration step.
- No application code imports github.com/justtrackio/gosoline/pkg/httpserver/auth, unless Google auth or custom auth behavior was explicitly deferred.
- No route definer returns *httpserver.Definitions.
- No &httpserver.Definitions{} allocations remain.
- No old Create*Handler calls remain.
- No old handler interfaces remain in application code.
- No code reads old httpserver.Response fields directly.
- All routes from the old router are present in the new router with the same HTTP method and path.
- All middleware is registered in the same effective order.
- Auth config has been moved below httpserver.<name>.auth and auth constructors receive the matching server name.
- CORS config has been moved from api_cors_* keys to httpserver.<name>.cors.* and package CORS registration uses CorsFactory or Cors(config, name).
- All request binding tags still match the route parameters, query parameters, and body fields expected by clients.
- All old request.ClientIp usages are replaced with ResolveClientIP(req) in BindR or BindNR handlers.
- All handled client errors still return the intended HTTP status code.
- Large-body endpoints have an explicit max_body_bytes setting if the v0.6.4 default 10 MiB limit is too small.
- Tests expecting binding failures now assert 400 Bad Request.
- Tests expecting internal error details in 5xx JSON responses now assert the sanitized default/private body, configure httpserver.<name>.errors.privacy: public, or install a custom error handler intentionally.
- CORS tests use full-origin matches and do not depend on partial regex matching.
- Health-check tests do not expect underlying module error strings in the HTTP response.
- Profiling tests or deployment checks expect the profiling API on loopback only.
- Tests compile and cover at least one route using URI binding, one route using JSON binding, and one middleware-protected route if applicable.

Commands to run:

- Run the repository's normal formatting and test commands.
- For a typical Go application, run: gofumpt -w . && go test ./...
- If the repository uses gosoline's fixture or integration build tags, use the repository-specific commands instead, for example: go test -tags fixtures,integration ./...
- If tests fail, fix compile errors first, then binding and response behavior regressions.
- Do not change public route paths, request field names, or status codes merely to make tests pass unless the migration explicitly requires it.

Stop conditions:

- Stop and ask for human guidance if route behavior is ambiguous and there are no tests or callers to confirm it.
- Stop and ask for human guidance if old auth code depends on shared package internals, including Google auth behavior, that cannot be reproduced safely.
- Stop and ask for human guidance if multiple old route-specific handlers have conflicting initialization logic and it is unclear whether they should be consolidated.
- Stop and ask for human guidance if a route depends on Gin-specific behavior and there is no clear equivalent without changing behavior.
- Stop and ask for human guidance if the migration would require unrelated application refactoring.
```
