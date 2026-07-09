# Migrating to the Standalone HTTP Server Package

This guide explains how to migrate from the old `github.com/justtrackio/gosoline/pkg/httpserver` package directly to the standalone `github.com/gosoline-project/httpserver` module at `v0.5.6`.

The new package keeps the Gin-based server lifecycle and middleware model, but it changes the application-facing API. Instead of building and returning a `*httpserver.Definitions` tree, you receive a `*httpserver.Router` and register routes on it directly. Instead of implementing handler interfaces and manually unpacking an `httpserver.Request`, handlers are plain functions or methods registered through `Bind`, `BindN`, `BindR`, or SSE binding helpers.

## 1. Update the Dependency[​](#1-update-the-dependency "Direct link to 1. Update the Dependency")

Add the standalone module:

```
go get github.com/gosoline-project/httpserver@v0.5.6
```

Update imports:

```
// Old

import "github.com/justtrackio/gosoline/pkg/httpserver"



// New

import "github.com/gosoline-project/httpserver"
```

Keep gosoline imports for application, config, logging, and other framework packages:

```
import (

    "github.com/gosoline-project/httpserver"

    "github.com/justtrackio/gosoline/pkg/application"

    "github.com/justtrackio/gosoline/pkg/cfg"

    "github.com/justtrackio/gosoline/pkg/log"

)
```

## 2. Migrate Server Startup[​](#2-migrate-server-startup "Direct link to 2. Migrate Server Startup")

### Old[​](#old "Direct link to Old")

The old package commonly used a `Definer` returning route definitions:

```
func main() {

    application.RunHttpDefaultServer(DefineRouter)

}



func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger) (*httpserver.Definitions, error) {

    definitions := &httpserver.Definitions{}

    definitions.GET("/api/users", httpserver.CreateHandler(handler))



    return definitions, nil

}
```

### New[​](#new "Direct link to New")

The new package uses a `RouterFactory`. The router is passed in and your function mutates it:

```
func main() {

    httpserver.RunDefaultServer(DefineRouter)

}



func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

    router.GET("/api/users", listUsers)



    return nil

}
```

If your application already uses `application.Run` with explicit modules, register the HTTP server as a module factory:

```
func main() {

    application.Run(

        application.WithModuleFactory("http", httpserver.NewServer("default", DefineRouter)),

    )

}
```

## 3. Migrate Route Definitions[​](#3-migrate-route-definitions "Direct link to 3. Migrate Route Definitions")

### Old[​](#old-1 "Direct link to Old")

```
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

### New[​](#new-1 "Direct link to New")

```
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

## 4. Use the `With` Pattern for Handler Dependencies[​](#4-use-the-with-pattern-for-handler-dependencies "Direct link to 4-use-the-with-pattern-for-handler-dependencies")

For handlers that need dependencies, use `httpserver.With`. The handler factory receives `ctx`, `config`, and `logger`, and the registration function wires methods to routes.

```
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

### From Route-Specific Handlers to Route-Group Handlers[​](#from-route-specific-handlers-to-route-group-handlers "Direct link to From Route-Specific Handlers to Route-Group Handlers")

With the old package, applications commonly used one handler instance per route. Each route-specific handler owned its own `Handle` method and was constructed before route registration:

```
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

```
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

    router.Group("/api/users").HandleWith(httpserver.With(NewUserHandler, func(r *httpserver.Router, h *UserHandler) {

        r.GET("", httpserver.BindN(h.ListUsers))

        r.GET("/:id", httpserver.Bind(h.GetUser))

    }))



    return nil

}
```

This consolidation is not required for every migration, but it is usually the cleaner target. Keep separate handlers when routes truly have unrelated dependencies or lifecycle needs. Otherwise, prefer one handler per domain or route group, with one method per endpoint.

## 5. Migrate Handler Signatures[​](#5-migrate-handler-signatures "Direct link to 5. Migrate Handler Signatures")

### Handler without Request Input[​](#handler-without-request-input "Direct link to Handler without Request Input")

Old handlers often implemented `HandlerWithoutInput`:

```
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

New handlers without request input use `BindN` and return `httpserver.Response`:

```
func (h *UserHandler) ListUsers(ctx context.Context) (httpserver.Response, error) {

    users, err := h.store.List(ctx)

    if err != nil {

        return nil, err

    }



    return httpserver.NewJsonResponse(users), nil

}
```

Register it with:

```
r.GET("", httpserver.BindN(h.ListUsers))
```

### Handler with Input[​](#handler-with-input "Direct link to Handler with Input")

Old input binding required `GetInput` and one of the `Create*Handler` helpers:

```
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

```
type GetUserInput struct {

    Id uint `uri:"id" binding:"required"`

}



func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (httpserver.Response, error) {

    user, err := h.store.Get(ctx, input.Id)

    if err != nil {

        return nil, err

    }



    return httpserver.NewJsonResponse(user), nil

}
```

Register it with:

```
r.GET("/:id", httpserver.Bind(h.GetUser))
```

## 6. Migrate Request Binding[​](#6-migrate-request-binding "Direct link to 6. Migrate Request Binding")

The old package selected a binding helper explicitly:

| Old helper                      | Typical new migration                                               |
| ------------------------------- | ------------------------------------------------------------------- |
| `CreateHandler`                 | `BindN` for no input, or raw Gin handler if you need `*gin.Context` |
| `CreateJsonHandler`             | `Bind` with `json` tags                                             |
| `CreateQueryHandler`            | `Bind` with `form` tags                                             |
| `CreateUriHandler`              | `Bind` with `uri` tags                                              |
| `CreateMultipleBindingsHandler` | `Bind` with multiple tags, or explicit binders if needed            |
| `CreateRawHandler`              | `BindNR` or `BindR` and read `req.Body`                             |
| `CreateReaderHandler`           | `BindNR` or `BindR` and use `req.Body`                              |
| `CreateSseHandler`              | `BindSse` or `BindSseN`                                             |

The new `Bind` function inspects struct tags and request content type:

```
type CreateUserInput struct {

    Name  string `json:"name" binding:"required"`

    Email string `json:"email" binding:"required,email"`

}



func (h *UserHandler) CreateUser(ctx context.Context, input *CreateUserInput) (httpserver.Response, error) {

    user, err := h.store.Create(ctx, input.Name, input.Email)

    if err != nil {

        return nil, err

    }



    return httpserver.NewJsonResponse(user, httpserver.WithStatusCode(http.StatusCreated)), nil

}
```

You can combine URI, query, and body fields in one input struct:

```
type UpdateUserInput struct {

    Id    uint   `uri:"id" binding:"required"`

    Name  string `json:"name"`

    Email string `json:"email"`

    Force bool   `form:"force"`

}
```

Use `form` tags for query string parameters and form-encoded bodies.

## 7. Replace Request Parameter Helpers[​](#7-replace-request-parameter-helpers "Direct link to 7. Replace Request Parameter Helpers")

The old package exposed helpers such as `GetStringFromRequest` and `GetUintFromRequest` for path parameters stored in `httpserver.Request.Params`.

### Old[​](#old-2 "Direct link to Old")

```
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

### New[​](#new-2 "Direct link to New")

Prefer typed input structs with `uri` tags:

```
type GetUserInput struct {

    Id uint `uri:"id" binding:"required"`

}



func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (httpserver.Response, error) {

    user, err := h.store.Get(ctx, input.Id)

    if err != nil {

        return nil, err

    }



    return httpserver.NewJsonResponse(user), nil

}
```

If you need direct Gin access, register a regular Gin handler:

```
router.GET("/api/users/:id", func(ginCtx *gin.Context) {

    id := ginCtx.Param("id")

    ginCtx.JSON(http.StatusOK, gin.H{"id": id})

})
```

## 8. Access Raw Requests and Bodies[​](#8-access-raw-requests-and-bodies "Direct link to 8. Access Raw Requests and Bodies")

The old `CreateRawHandler` read the body into `request.Body` as a string. The new package keeps this explicit by passing the raw `*http.Request` to your handler.

```
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

```
r.POST("/import", httpserver.BindNR(h.ImportUsers))
```

For typed input plus raw request access, use `BindR`:

```
func (h *UserHandler) UploadAvatar(ctx context.Context, req *http.Request, input *UploadAvatarInput) (httpserver.Response, error) {

    contentType := req.Header.Get("Content-Type")

    _ = contentType



    return httpserver.NewStatusResponse(http.StatusNoContent), nil

}
```

### Client IP[​](#client-ip "Direct link to Client IP")

The old `httpserver.Request` exposed `request.ClientIp`. In the standalone package, use raw request access and resolve the client IP explicitly:

```
func (h *UserHandler) GetUser(ctx context.Context, req *http.Request, input *GetUserInput) (httpserver.Response, error) {

    clientIP, err := httpserver.ResolveClientIP(req)

    if err != nil {

        return nil, err

    }



    _ = clientIP



    // ...

}
```

Register handlers that need both typed input and the raw request with `BindR`. Use `BindNR` when the handler needs the raw request but no typed input.

## 9. Migrate Responses[​](#9-migrate-responses "Direct link to 9. Migrate Responses")

The new package returns a `Response` interface instead of the old concrete `*Response` struct.

### JSON Responses[​](#json-responses "Direct link to JSON Responses")

```
return httpserver.NewJsonResponse(user), nil
```

### Text Responses[​](#text-responses "Direct link to Text Responses")

Old low-level response:

```
return httpserver.NewResponse("ok", httpserver.ContentTypeText, http.StatusOK, make(http.Header)), nil
```

New response:

```
return httpserver.NewTextResponse("ok"), nil
```

### Status Responses[​](#status-responses "Direct link to Status Responses")

```
return httpserver.NewStatusResponse(http.StatusNoContent), nil
```

### Custom Status Codes and Headers[​](#custom-status-codes-and-headers "Direct link to Custom Status Codes and Headers")

Old response options were limited and the constructor accepted status and headers directly. New responses use options consistently:

```
return httpserver.NewJsonResponse(

    user,

    httpserver.WithStatusCode(http.StatusCreated),

    httpserver.WithHeader("X-User-Id", strconv.FormatUint(uint64(user.Id), 10)),

), nil
```

For raw bodies, use `NewResponse` with options:

```
return httpserver.NewResponse(

    httpserver.WithBody([]byte("accepted")),

    httpserver.WithHeader("Content-Type", "text/plain; charset=utf-8"),

    httpserver.WithStatusCode(http.StatusAccepted),

), nil
```

## 10. Migrate Error Handling[​](#10-migrate-error-handling "Direct link to 10. Migrate Error Handling")

For unexpected errors, keep returning `nil, err`:

```
func (h *UserHandler) ListUsers(ctx context.Context) (httpserver.Response, error) {

    users, err := h.store.List(ctx)

    if err != nil {

        return nil, fmt.Errorf("could not list users: %w", err)

    }



    return httpserver.NewJsonResponse(users), nil

}
```

For handled client errors, return an error response and a nil Go error:

```
func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (httpserver.Response, error) {

    user, err := h.store.Get(ctx, input.Id)

    if errors.Is(err, ErrUserNotFound) {

        return httpserver.GetErrorHandler()(http.StatusNotFound, err), nil

    }

    if err != nil {

        return nil, fmt.Errorf("could not get user: %w", err)

    }



    return httpserver.NewJsonResponse(user), nil

}
```

In `v0.5.6`, the default error handler exposes 4xx error messages but sanitizes 5xx responses as `{"err":"internal server error"}`. Binding and validation failures from `Bind` and `BindSse` are client errors and return `400 Bad Request`.

The 5xx behavior is controlled by `httpserver.<name>.errors.privacy`. The default is `private`, which hides internal error details. Set it to `public` only when clients should receive the original internal error message:

```
httpserver:

  default:

    errors:

      privacy: public
```

If middleware attaches an error to the Gin context and needs a non-500 status code, wrap it with `NewErrorWithStatus`:

```
ginCtx.Error(httpserver.NewErrorWithStatus(http.StatusBadRequest, err))
```

## 11. Migrate Middleware[​](#11-migrate-middleware "Direct link to 11. Migrate Middleware")

Gin middleware remains compatible:

```
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

    router.Use(requestIdMiddleware)



    api := router.Group("/api")

    api.Use(authMiddleware)

    api.GET("/users", httpserver.BindN(listUsers))



    return nil

}
```

For middleware that needs configuration, a logger, or server settings, use `UseFactory`:

```
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

### CORS[​](#cors "Direct link to CORS")

Move the old global CORS settings below the named HTTP server:

| Old gosoline key                  | New standalone key for server `default`          |
| --------------------------------- | ------------------------------------------------ |
| `api_cors_allowed_origin_pattern` | `httpserver.default.cors.allowed_origin_pattern` |
| `api_cors_allowed_headers`        | `httpserver.default.cors.allowed_headers`        |
| `api_cors_allowed_methods`        | `httpserver.default.cors.allowed_methods`        |

```
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

```
router.UseFactory(httpserver.CorsFactory)
```

If you construct the middleware manually, pass the server name explicitly:

```
corsMiddleware, err := httpserver.Cors(config, "default")

if err != nil {

    return err

}

router.Use(corsMiddleware)
```

The origin pattern is matched against the full `Origin` value. For example, `https://example\\.com` allows `https://example.com`, but not `https://example.com.evil.com`.

### Authentication[​](#authentication "Direct link to Authentication")

The standalone module includes auth helpers in `github.com/gosoline-project/httpserver/auth`. Use them when they match the old embedded package behavior.

Import the standalone auth package separately:

```
import (

    "github.com/gosoline-project/httpserver"

    "github.com/gosoline-project/httpserver/auth"

)
```

Auth settings are now scoped to the named HTTP server. `httpserver.RunDefaultServer` uses the server name `default`, so auth config belongs below `httpserver.default.auth`. If you register `httpserver.NewServer("admin", DefineRouter)`, the auth settings belong below `httpserver.admin.auth`, and you pass `"admin"` to auth constructors.

Common config key migrations:

| Old gosoline key               | New standalone key for server `default`       |
| ------------------------------ | --------------------------------------------- |
| `api_auth_keys`                | `httpserver.default.auth.keys`                |
| `api_auth_basic_users`         | `httpserver.default.auth.basic.users`         |
| `api_auth_bearer_id_header`    | `httpserver.default.auth.bearer.id_header`    |
| `api_auth_bearer_token_header` | `httpserver.default.auth.bearer.token_header` |
| `httpserver.<name>.auth.jwt.*` | `httpserver.<name>.auth.jwt.*`                |

### API Key Auth[​](#api-key-auth "Direct link to API Key Auth")

Old application-owned middleware often looked like this:

```
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

```
httpserver:

  default:

    auth:

      keys:

        - ${env:API_KEY}
```

```
apiKeyAuth, err := auth.NewConfigKeyHandler(config, logger, "default", auth.ProvideValueFromHeader(auth.HeaderApiKey))

if err != nil {

    return err

}



api := router.Group("/api")

api.Use(apiKeyAuth)
```

Use `auth.ProvideValueFromQueryParam("api_key")` or `auth.ProvideValueFromUriPath("apiKey")` if your old middleware read the key from a query parameter or path parameter instead of a header.

For reusable middleware registration that should use the current server name automatically, register the settings-aware factory:

```
api := router.Group("/api")

api.UseFactory(auth.ConfigKeyHandlerFactory(auth.ProvideValueFromHeader(auth.HeaderApiKey)))
```

### Basic Auth[​](#basic-auth "Direct link to Basic Auth")

Move Basic Auth users below the named server:

```
httpserver:

  default:

    auth:

      basic:

        users:

          - admin:${env:BASIC_AUTH_ADMIN_PASSWORD}
```

Register the middleware with the same server name:

```
basicAuth, err := auth.NewBasicAuthHandler(config, logger, "default")

if err != nil {

    return err

}



router.Group("/admin").Use(basicAuth)
```

Or use the settings-aware factory:

```
router.Group("/admin").UseFactory(auth.BasicAuthHandlerFactory)
```

Unauthorized Basic Auth responses use the configured gosoline app identity name as the Basic realm.

### Bearer Token Auth[​](#bearer-token-auth "Direct link to Bearer Token Auth")

Move bearer header configuration below the named server:

```
httpserver:

  default:

    auth:

      bearer:

        id_header: X-BEARER-ID

        token_header: X-BEARER-TOKEN
```

Then pass the server name and your bearer provider:

```
bearerAuth, err := auth.NewTokenBearerHandler(config, logger, "default", provider)

if err != nil {

    return err

}



router.Group("/api").Use(bearerAuth)
```

Or use the settings-aware factory:

```
router.Group("/api").UseFactory(auth.TokenBearerHandlerFactory(provider))
```

### JWT Auth[​](#jwt-auth "Direct link to JWT Auth")

JWT settings live below `httpserver.<name>.auth.jwt`:

```
httpserver:

  default:

    auth:

      jwt:

        signingSecret: ${env:JWT_SIGNING_SECRET}

        issuer: my-service

        expireDuration: 15m
```

```
jwtAuth, err := auth.NewJwtAuthHandler(config, "default")

if err != nil {

    return err

}



router.Group("/api").Use(jwtAuth)
```

Or use the settings-aware factory:

```
router.Group("/api").UseFactory(auth.JwtAuthHandlerFactory)
```

The standalone JWT helper validates `Authorization: Bearer <token>` headers using HS256 and requires an `email` claim for the authenticated subject.

### Auth Subjects and Chains[​](#auth-subjects-and-chains "Direct link to Auth Subjects and Chains")

Successful authenticators attach an `auth.Subject` to the request context. In bound handlers, retrieve it with `auth.GetSubject(ctx)`:

```
func (h *UserHandler) Me(ctx context.Context) (httpserver.Response, error) {

    subject := auth.GetSubject(ctx)



    return httpserver.NewJsonResponse(subject), nil

}
```

If a route accepts multiple auth methods, build authenticators and combine them with `auth.NewChainHandler`. You can optionally restrict enabled methods per server with `httpserver.<name>.auth.allowedAuthenticators`.

```
httpserver:

  default:

    auth:

      allowedAuthenticators:

        - apiKey

        - jwtAuth
```

```
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

See [Authenticate requests](/docs/pr-1/how-to/http-server/authentication.md) for the full standalone auth guide.

## 12. Review Configuration[​](#12-review-configuration "Direct link to 12. Review Configuration")

The main server configuration still lives under `httpserver.<name>`:

```
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

| Area               | Notes                                                                                                                                                                                                             |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Server name        | `RunDefaultServer` reads `httpserver.default`. `NewServer("admin", ...)` reads `httpserver.admin`.                                                                                                                |
| Routes             | `/health` is still registered by the server. Unhealthy modules are reported as `"unhealthy"`; underlying error strings are logged but not exposed in the response.                                                |
| Compression        | Exclude SSE endpoints from compression.                                                                                                                                                                           |
| Request body limit | `max_body_bytes` defaults to `10485760` bytes (10 MiB). Set it to `0` to disable the limit or raise it for large upload endpoints. The limit is applied after request decompression.                              |
| Binding errors     | `Bind` and `BindSse` binding or validation failures return `400 Bad Request`.                                                                                                                                     |
| Error responses    | The default handler exposes 4xx error messages and returns `{"err":"internal server error"}` for 5xx responses. Set `httpserver.<name>.errors.privacy` to `public` only when 5xx error details should be exposed. |
| CORS               | Move old `api_cors_*` keys to `httpserver.<name>.cors.*`. `allowed_origin_pattern` is matched against the full `Origin` value, so partial regex matches do not allow origins.                                     |
| Profiling          | Profiling is still configured under `profiling` and binds to `127.0.0.1:<port>`.                                                                                                                                  |
| Auth               | Auth config is scoped below `httpserver.<name>.auth`; old global auth keys must be moved below the server name.                                                                                                   |

## 13. Migrate Tests[​](#13-migrate-tests "Direct link to 13. Migrate Tests")

Old test suites usually implemented definition-based setup:

```
func (s *UserApiSuite) SetupApiDefinitions() httpserver.Definer {

    return DefineRouter

}
```

New httpserver test cases use router factories:

```
func (s *UserApiSuite) SetupHttpServerRouter() httpserver.RouterFactory {

    return DefineRouter

}
```

For handler-level tests, prefer testing handler methods directly where possible:

```
func TestGetUser(t *testing.T) {

    handler := &UserHandler{store: fakeStore}



    response, err := handler.GetUser(context.Background(), &GetUserInput{Id: 1})



    require.NoError(t, err)

    assert.Equal(t, http.StatusOK, response.StatusCode())

}
```

For binding and routing tests, use the package test helpers or an application test case with `SetupHttpServerRouter`.

## 14. Complete Before and After Example[​](#14-complete-before-and-after-example "Direct link to 14. Complete Before and After Example")

### Old[​](#old-3 "Direct link to Old")

```
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

### New[​](#new-3 "Direct link to New")

```
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



func (h *UserHandler) ListUsers(ctx context.Context) (httpserver.Response, error) {

    users, err := h.store.List(ctx)

    if err != nil {

        return nil, err

    }



    return httpserver.NewJsonResponse(users), nil

}



func (h *UserHandler) GetUser(ctx context.Context, input *GetUserInput) (httpserver.Response, error) {

    user, err := h.store.Get(ctx, input.Id)

    if errors.Is(err, ErrUserNotFound) {

        return httpserver.GetErrorHandler()(http.StatusNotFound, err), nil

    }

    if err != nil {

        return nil, err

    }



    return httpserver.NewJsonResponse(user), nil

}



func (h *UserHandler) CreateUser(ctx context.Context, input *CreateUserInput) (httpserver.Response, error) {

    user, err := h.store.Create(ctx, input.Name, input.Email)

    if err != nil {

        return nil, err

    }



    return httpserver.NewJsonResponse(user, httpserver.WithStatusCode(http.StatusCreated)), nil

}



func (h *UserHandler) DeleteUser(ctx context.Context, input *GetUserInput) (httpserver.Response, error) {

    if err := h.store.Delete(ctx, input.Id); err != nil {

        return nil, err

    }



    return httpserver.NewStatusResponse(http.StatusNoContent), nil

}
```

## Migration Checklist[​](#migration-checklist "Direct link to Migration Checklist")

* Replace `github.com/justtrackio/gosoline/pkg/httpserver` imports with `github.com/gosoline-project/httpserver`.
* Replace `Definer` functions returning `*Definitions` with `RouterFactory` functions accepting `*Router`.
* Replace `application.RunHttpDefaultServer` with `httpserver.RunDefaultServer`, or register `httpserver.NewServer` as a module factory.
* Replace `Definitions` route registration with direct `Router` registration.
* Replace old handler interfaces with plain methods registered through `Bind`, `BindN`, `BindR`, or `BindNR`.
* Consolidate route-specific handlers into route-group or domain handlers where the routes share dependencies.
* Replace `GetInput` and `request.Body.(*Input)` with typed input arguments.
* Replace path parameter helpers with `uri` tags on input structs.
* Replace `CreateRawHandler` and `CreateReaderHandler` with `BindR` or `BindNR` and explicit `req.Body` handling.
* Replace `request.ClientIp` with `ResolveClientIP(req)` in `BindR` or `BindNR` handlers.
* Replace low-level `NewResponse(body, contentType, statusCode, header)` calls with `NewTextResponse`, `NewJsonResponse`, `NewStatusResponse`, or `NewResponse` options.
* Replace old `pkg/httpserver/auth` imports with `github.com/gosoline-project/httpserver/auth`, move auth config below `httpserver.<name>.auth`, and pass the server name to auth constructors.
* Move old global CORS config from `api_cors_*` keys to `httpserver.<name>.cors.*`, and register package CORS with `router.UseFactory(httpserver.CorsFactory)` or `httpserver.Cors(config, name)`.
* Review `max_body_bytes`; keep the v0.5.6 default 10 MiB limit, raise it, or set it to `0` intentionally.
* Update tests for v0.5.6 behavior: binding errors return 400, default/private 5xx bodies are sanitized, `errors.privacy: public` exposes 5xx messages, CORS patterns match full origins, health responses hide module error details, and profiling binds to loopback.
* Update test suites from definition setup to router factory setup.
* Run your HTTP tests and exercise validation errors, client errors, and middleware behavior after migration.

## AI Agent Migration Instructions[​](#ai-agent-migration-instructions "Direct link to AI Agent Migration Instructions")

If you want an AI agent to perform this migration, copy the prepared migration instructions and provide them as the task prompt. The instructions are intentionally operational and include discovery steps, migration order, API mappings, validation checks, and stop conditions.

Copy AI migration instructions
