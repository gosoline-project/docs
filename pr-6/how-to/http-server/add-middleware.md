---
sidebar:
  order: 5
title: Add middleware
---

The httpserver includes a built-in middleware chain and supports custom middleware at both the router and route-group level.

## Built-in middleware

Every httpserver automatically includes these middleware, applied in order:

1. **Sampling** — Applies sampling decisions based on HTTP headers
2. **Metric** — Records request count, response time, and status code metrics
3. **Logging** — Logs request details using fingers-crossed scope (only logs on errors)
4. **Compression** — Handles gzip compression/decompression
5. **Max body size** — Limits incoming request bodies with `httpserver.<name>.max_body_bytes`
6. **Error** — Catches errors attached to the Gin context and returns JSON error responses, using `httpserver.<name>.errors`
7. **Recovery** — Catches panics and returns 500 errors
8. **Location** — Extracts location information
9. **Connection lifecycle** — Manages connection age and request count limits
10. **Concurrency** — Limits concurrent requests
11. **Chaos** — Randomly introduces failures for resilience testing (disabled by default)

You don't need to configure these — they are enabled by default (except chaos, which is opt-in).

## Request body size limit

Every server includes `MaxBodySizeMiddleware`. The default limit is 10 MiB:

```yaml
httpserver:
  default:
    max_body_bytes: 10485760
```

Set `max_body_bytes` to `0` to disable the limit. Oversized bodies fail when a handler reads the request body, so bound JSON/form handlers return a client error instead of invoking your handler.

## Error privacy

Every server includes error middleware. By default, 5xx responses hide internal error details and return `{"err":"internal server error"}`. Client errors with a status below 500 still expose their error message.

```yaml
httpserver:
  default:
    errors:
      privacy: private
```

Set `privacy` to `public` only when clients should see internal 5xx error messages, for example in local development or trusted internal services:

```yaml
httpserver:
  default:
    errors:
      privacy: public
```

## Custom middleware

Use `router.Use()` to add middleware to the entire server or a route group:

```go
router.Use(func(ginCtx *gin.Context) {
    start := time.Now()
    ginCtx.Next()
    elapsed := time.Since(start)
    logger.Info(ctx, "request completed in %s", elapsed)
})
```

Apply middleware to a specific group:

```go
api := router.Group("/api")
api.Use(authMiddleware)
api.GET("/users", listUsers)
```

## Factory-based middleware

Use `router.UseFactory()` for middleware that needs access to config, logger, or the resolved server settings. The factory is called lazily when the router is built:

```go
router.UseFactory(func(ctx context.Context, config cfg.Config, logger log.Logger, settings *httpserver.Settings) (gin.HandlerFunc, error) {
    apiKey := config.GetString("api_key")
    return func(ginCtx *gin.Context) {
        if ginCtx.GetHeader("X-API-Key") != apiKey {
            ginCtx.AbortWithStatus(401)
            return
        }
        ginCtx.Next()
    }, nil
})
```

`settings.Name` contains the server name, such as `"default"` for `RunDefaultServer` or `"admin"` for `NewServer("admin", ...)`. This is useful for reusable middleware that reads server-scoped configuration.

The `auth` package provides settings-aware middleware factories for common authenticators, so the server name does not need to be hard-coded:

```go
router.UseFactory(auth.ConfigKeyHandlerFactory(auth.ProvideValueFromHeader(auth.HeaderApiKey)))
```

## CORS

The package provides a CORS middleware that reads settings from the named HTTP server config:

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

Use the settings-aware factory when registering middleware inside a `RouterFactory`:

```go
import "github.com/gosoline-project/httpserver"

// In your RouterFactory:
router.UseFactory(httpserver.CorsFactory)
```

The factory uses the current server name, so `RunDefaultServer` reads `httpserver.default.cors`, while `NewServer("admin", ...)` reads `httpserver.admin.cors`.

If you create the middleware manually, pass the server name explicitly:

```go
corsMiddleware, err := httpserver.Cors(config, "default")
if err != nil {
    return err
}
router.Use(corsMiddleware)
```

The origin pattern is matched against the full origin. For example, `https://example\.com` allows `https://example.com`, but not `https://example.com.evil.com`.

Or use the default Gin CORS middleware directly:

```go
import "github.com/gin-contrib/cors"

router.Use(cors.Default())
```

## Embedded static file serving

`CreateEmbeddedStaticServe` is itself a middleware factory:

```go
router.UseFactory(httpserver.CreateEmbeddedStaticServe(publicFs, "public", "/api"))
```

See [Serve a frontend](/docs/pr-6/how-to/http-server/serve-a-frontend/) for full details.

## Complete example


<details>
<summary>main.go</summary>

```go title="main.go" lineNumbers
package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosoline-project/httpserver"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

func main() {
	httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
		router.Use(func(ginCtx *gin.Context) {
			start := time.Now()
			ginCtx.Next()
			elapsed := time.Since(start)
			logger.Info(ctx, "request completed in %s", elapsed)
		})

		router.UseFactory(func(ctx context.Context, config cfg.Config, logger log.Logger, settings *httpserver.Settings) (gin.HandlerFunc, error) {
			return func(ginCtx *gin.Context) {
				ginCtx.Header("X-Server-Name", settings.Name)
				ginCtx.Next()
			}, nil
		})

		router.GET("/ping", func(ginCtx *gin.Context) {
			ginCtx.JSON(200, gin.H{"message": "pong"})
		})

		return nil
	})
}
```

</details>

<details>
<summary>config.dist.yml</summary>

```yaml title="config.dist.yml" lineNumbers
app:
  env: dev
  name: middleware

httpserver:
  default:
    port: 8088
```

</details>
