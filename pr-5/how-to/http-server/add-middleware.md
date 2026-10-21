# Add middleware

The httpserver includes a built-in middleware chain and supports custom middleware at both the router and route-group level.

## Built-in middleware[​](#built-in-middleware "Direct link to Built-in middleware")

Every httpserver automatically includes these middleware, applied in order:

1. **Sampling** — Applies sampling decisions based on HTTP headers
2. **Metric** — Records response-duration observations, explicit request rejections, request/response counters, and active-request/connection gauges
3. **Logging** — Logs request details using fingers-crossed scope (only logs on errors)
4. **Compression** — Handles gzip compression/decompression
5. **Max body size** — Limits incoming request bodies with `httpserver.<name>.max_body_bytes`
6. **Error** — Catches errors attached to the Gin context and returns JSON error responses, using `httpserver.<name>.errors`
7. **Recovery** — Catches panics and returns 500 errors
8. **Location** — Extracts location information
9. **Connection lifecycle** — Manages connection age and request count limits
10. **Concurrency** — Limits concurrent requests
11. **Chaos** — Randomly introduces failures for resilience testing (disabled by default)

With metric export enabled, the middleware emits the high-priority monotonic `Count` counters after a matched route's middleware chain unwinds. `http.server.requests` (`name`) and `http.server.routed.requests` (`name`, `route`, `request.method`) remain statusless and keep their existing zero defaults: one per configured server and per registered application route/method pair. The new `http.server.responses` and `http.server.routed.responses` counters retain the full numeric `response.status_code` and add bounded `response.status_class` values (`1xx` through `5xx`; default outside 100–599); they have no zero defaults. All count matched responses, including middleware-aborted requests and matched `/health`; unmatched paths emit none. The duration histogram keeps the full code without a class, while `http.server.rejected.requests` remains statusless. See the [metric schema](/docs/pr-5/reference/metric-schema/.md) for writer-specific selection and the V1→V2 migration.

These are canonical Gosoline metric names; the configured writer determines their exported names and units. The middleware is installed by default; metric export requires the application to wire `application.WithMetrics`, set `metric.enabled: true`, and configure at least one writer under `metric.writers`. See the [published HTTP server metrics reference](/docs/pr-5/reference/package-httpserver/.md) for the complete table; the SDK's separate [HTTP server implementation](/docs/pr-5/reference/package-gosoline-httpserver/.md) emits the same eight canonical names. For backend-specific name and unit conversion, see [Metric schema](/docs/pr-5/reference/metric-schema/.md). Chaos is the only opt-in built-in middleware.

## Request body size limit[​](#request-body-size-limit "Direct link to Request body size limit")

Every server includes `MaxBodySizeMiddleware`. The default limit is 10 MiB:

```
httpserver:

  default:

    max_body_bytes: 10485760
```

Set `max_body_bytes` to `0` to disable the limit. Oversized bodies fail when a handler reads the request body, so bound JSON/form handlers return a client error instead of invoking your handler.

## Error privacy[​](#error-privacy "Direct link to Error privacy")

Every server includes error middleware. By default, 5xx responses hide internal error details and return `{"err":"internal server error"}`. Client errors with a status below 500 still expose their error message.

```
httpserver:

  default:

    errors:

      privacy: private
```

Set `privacy` to `public` only when clients should see internal 5xx error messages, for example in local development or trusted internal services:

```
httpserver:

  default:

    errors:

      privacy: public
```

## Custom middleware[​](#custom-middleware "Direct link to Custom middleware")

Use `router.Use()` to add middleware to the entire server or a route group:

```
router.Use(func(ginCtx *gin.Context) {

    start := time.Now()

    ginCtx.Next()

    elapsed := time.Since(start)

    logger.Info(ctx, "request completed in %s", elapsed)

})
```

Apply middleware to a specific group:

```
api := router.Group("/api")

api.Use(authMiddleware)

api.GET("/users", listUsers)
```

## Factory-based middleware[​](#factory-based-middleware "Direct link to Factory-based middleware")

Use `router.UseFactory()` for middleware that needs access to config, logger, or the resolved server settings. The factory is called lazily when the router is built:

```
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

```
router.UseFactory(auth.ConfigKeyHandlerFactory(auth.ProvideValueFromHeader(auth.HeaderApiKey)))
```

## CORS[​](#cors "Direct link to CORS")

The package provides a CORS middleware that reads settings from the named HTTP server config:

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

Use the settings-aware factory when registering middleware inside a `RouterFactory`:

```
import "github.com/gosoline-project/httpserver"



// In your RouterFactory:

router.UseFactory(httpserver.CorsFactory)
```

The factory uses the current server name, so `RunDefaultServer` reads `httpserver.default.cors`, while `NewServer("admin", ...)` reads `httpserver.admin.cors`.

If you create the middleware manually, pass the server name explicitly:

```
corsMiddleware, err := httpserver.Cors(config, "default")

if err != nil {

    return err

}

router.Use(corsMiddleware)
```

The origin pattern is matched against the full origin. For example, `https://example\.com` allows `https://example.com`, but not `https://example.com.evil.com`.

Or use the default Gin CORS middleware directly:

```
import "github.com/gin-contrib/cors"



router.Use(cors.Default())
```

## Embedded static file serving[​](#embedded-static-file-serving "Direct link to Embedded static file serving")

`CreateEmbeddedStaticServe` is itself a middleware factory:

```
router.UseFactory(httpserver.CreateEmbeddedStaticServe(publicFs, "public", "/api"))
```

See [Serve a frontend](/docs/pr-5/how-to/http-server/serve-a-frontend/.md) for full details.

## Complete example[​](#complete-example "Direct link to Complete example")

<!-- -->

main.go

main.go

```
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

config.dist.yml

config.dist.yml

```
app:

  env: dev

  name: middleware



httpserver:

  default:

    port: 8088
```
