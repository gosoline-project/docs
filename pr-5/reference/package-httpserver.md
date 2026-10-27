# httpserver package

Package `httpserver` provides a modular HTTP server built on [Gin](https://gin-gonic.com/) and the gosoline kernel. It offers structured request binding, typed responses, middleware, SSE streaming, and dependency-injected handler registration.

```
import "github.com/gosoline-project/httpserver"
```

## Application helpers[​](#application-helpers "Direct link to Application helpers")

### [RunDefaultServer()](https://github.com/gosoline-project/httpserver/blob/main/application_run.go)[​](#rundefaultserver "Direct link to rundefaultserver")

#### Usage[​](#usage "Direct link to Usage")

```
httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

    router.GET("/ping", handler)

    return nil

})
```

#### Description[​](#description "Direct link to Description")

Runs a single HTTP server named `"default"`, reading configuration from `httpserver.default`. This is the simplest way to start a server.

### [RunServers()](https://github.com/gosoline-project/httpserver/blob/main/application_run.go)[​](#runservers "Direct link to runservers")

#### Usage[​](#usage-1 "Direct link to Usage")

```
httpserver.RunServers(map[string]httpserver.RouterFactory{

    "public": publicDefiner,

    "admin":  adminDefiner,

})
```

#### Description[​](#description-1 "Direct link to Description")

Runs multiple named HTTP servers. Each server reads its configuration from `httpserver.<name>`.

## Server construction[​](#server-construction "Direct link to Server construction")

### [NewServer()](https://github.com/gosoline-project/httpserver/blob/main/server_module.go)[​](#newserver "Direct link to newserver")

#### Usage[​](#usage-2 "Direct link to Usage")

```
application.WithModuleFactory("http", httpserver.NewServer("default", myRouterFactory))
```

#### Description[​](#description-2 "Direct link to Description")

Creates a `kernel.ModuleFactory` for an HTTP server. The server is registered as a kernel module and runs at the Service stage. The `name` parameter determines both the config key (`httpserver.<name>`) and the module name (`httpserver-<name>`).

## Router[​](#router "Direct link to Router")

### [RouterFactory](https://github.com/gosoline-project/httpserver/blob/main/router.go)[​](#routerfactory "Direct link to routerfactory")

```
type RouterFactory func(ctx context.Context, config cfg.Config, logger log.Logger, router *Router) error
```

Implement this function to define your routes. Return an error to prevent the server from starting.

### Router methods[​](#router-methods "Direct link to Router methods")

#### [Handle()](https://github.com/gosoline-project/httpserver/blob/main/router.go)[​](#handle "Direct link to handle")

```
router.Handle(httpMethod, relativePath, handlers...)
```

Registers a route for the given HTTP method and path.

#### [GET() / POST() / PUT() / DELETE() / PATCH() / OPTIONS()](https://github.com/gosoline-project/httpserver/blob/main/router.go)[​](#get--post--put--delete--patch--options "Direct link to get--post--put--delete--patch--options")

```
router.GET("/users", handler)

router.POST("/users", handler)
```

Convenience methods for registering routes with specific HTTP methods.

#### [Group()](https://github.com/gosoline-project/httpserver/blob/main/router.go)[​](#group "Direct link to group")

```
api := router.Group("/api")

api.GET("/users", handler)
```

Creates a route group with a common path prefix. Groups can be nested.

#### [Use()](https://github.com/gosoline-project/httpserver/blob/main/router.go)[​](#use "Direct link to use")

```
router.Use(middlewareFunc)
```

Adds middleware to the router or group.

#### [UseFactory()](https://github.com/gosoline-project/httpserver/blob/main/router.go)[​](#usefactory "Direct link to usefactory")

```
router.UseFactory(func(ctx context.Context, config cfg.Config, logger log.Logger, settings *httpserver.Settings) (gin.HandlerFunc, error) {

    // create middleware with access to config, logger, and server settings

})
```

Adds a middleware factory that is resolved lazily when the router is built. `settings.Name` contains the server name used for server-scoped configuration.

#### [HandleWith()](https://github.com/gosoline-project/httpserver/blob/main/router.go)[​](#handlewith "Direct link to handlewith")

```
router.Group("/api").HandleWith(httpserver.With(NewHandler, func(r *httpserver.Router, h *Handler) {

    r.GET("/items", httpserver.Bind(h.ListItems))

}))
```

Registers routes using a `RegisterFactoryFunc` (typically created with `With()`).

## Handler registration[​](#handler-registration "Direct link to Handler registration")

### [With\[H\]()](https://github.com/gosoline-project/httpserver/blob/main/with.go)[​](#withh "Direct link to withh")

#### Usage[​](#usage-3 "Direct link to Usage")

```
router.HandleWith(httpserver.With(NewMyHandler, func(r *httpserver.Router, h *MyHandler) {

    r.GET("/items", httpserver.Bind(h.ListItems))

    r.POST("/items", httpserver.Bind(h.CreateItem))

}))
```

#### Description[​](#description-3 "Direct link to Description")

Generic function that creates a handler instance via the factory function, then calls the registration function with both a sub-router and the handler. This is the primary pattern for dependency-injected route registration.

The handler factory must have this signature:

```
func NewMyHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*MyHandler, error)
```

Since httpserver v0.6.4, handlers registered through `With` that implement `io.Closer` are closed after the server stops serving requests. Implement `Close() error` as blocking, idempotent cleanup. See the [released registration contract](https://github.com/gosoline-project/httpserver/blob/v0.6.4/with.go).

## Request binding[​](#request-binding "Direct link to Request binding")

### Bind functions[​](#bind-functions "Direct link to Bind functions")

| Function          | Handler signature                                                       | Use case                                      |
| ----------------- | ----------------------------------------------------------------------- | --------------------------------------------- |
| `Bind[I, O any]`  | `func(ctx context.Context, input *I) (O, error)`                        | Bind request to input and render typed output |
| `BindR[I, O any]` | `func(ctx context.Context, request *http.Request, input *I) (O, error)` | Bind with raw request access                  |
| `BindN[O any]`    | `func(ctx context.Context) (O, error)`                                  | No input binding                              |
| `BindNR[O any]`   | `func(ctx context.Context, request *http.Request) (O, error)`           | No input, with raw request                    |
| `BindSse[I]`      | `(ctx, *I, *SseWriter) error`                                           | SSE with input binding                        |
| `BindSseR[I]`     | `(ctx, *http.Request, *I, *SseWriter) error`                            | SSE with input + raw request                  |
| `BindSseN`        | `(ctx, *SseWriter) error`                                               | SSE with no input                             |
| `BindSseNR`       | `(ctx, *http.Request, *SseWriter) error`                                | SSE with raw request, no input                |

The four response binding helpers accept a typed result `O` and negotiate its representation from the request's `Accept` header. The default renderer supports JSON only, uses JSON when `Accept` is missing, and returns `406 Not Acceptable` when no accepted media type is supported. See [Customizing responses in Build an HTTP service](/docs/pr-5/how-to/http-server/build-an-http-service/.md#customizing-responses) for response metadata and explicit `Response` overrides.

Typed outputs are negotiated before their `StatusCode()` metadata is applied. For a bodyless `204 No Content` response that must be preserved when `Accept` is unsupported, return `NewStatusResponse(http.StatusNoContent)`; explicit responses bypass negotiation. Other typed outputs, including those with custom status codes, remain negotiated.

### Binding tags[​](#binding-tags "Direct link to Binding tags")

Input structs use standard tags to specify the data source:

| Tag       | Source                    | Example                    |
| --------- | ------------------------- | -------------------------- |
| `uri`     | Path parameters           | `uri:"id"`                 |
| `json`    | JSON body                 | `json:"name"`              |
| `form`    | Query string or form body | `form:"limit"`             |
| `xml`     | XML body                  | `xml:"item"`               |
| `yaml`    | YAML body                 | `yaml:"config"`            |
| `header`  | HTTP headers              | `header:"X-Request-Id"`    |
| `binding` | Validation rules          | `binding:"required,email"` |

Binding and validation failures return `400 Bad Request`. Validation bind failures are logged at warning level, not error level, with field, tag, invalid value, and parameter details; invalid values are truncated before logging.

## Responses[​](#responses "Direct link to Responses")

### Response constructors[​](#response-constructors "Direct link to Response constructors")

| Constructor                                                           | Description                                         |
| --------------------------------------------------------------------- | --------------------------------------------------- |
| `NewJsonResponse[T](body T, opts ...ResponseOption) *jsonResponse[T]` | JSON response with `Content-Type: application/json` |
| `NewTextResponse(text string, opts ...ResponseOption) *response`      | Plain text response                                 |
| `NewStatusResponse(statusCode int, opts ...ResponseOption) *response` | Status-only response (no body)                      |
| `NewResponse(opts ...ResponseOption) *response`                       | Base response constructor                           |

### Response options[​](#response-options "Direct link to Response options")

| Option                             | Description              |
| ---------------------------------- | ------------------------ |
| `WithStatusCode(code int)`         | Set the HTTP status code |
| `WithHeader(key, value string)`    | Add a single header      |
| `WithHeaders(headers http.Header)` | Merge multiple headers   |
| `WithBody(body []byte)`            | Set raw response body    |

## SSE[​](#sse "Direct link to SSE")

### [SseWriter](https://github.com/gosoline-project/httpserver/blob/main/sse_writer.go)[​](#ssewriter "Direct link to ssewriter")

#### Methods[​](#methods "Direct link to Methods")

| Method                            | Description                         |
| --------------------------------- | ----------------------------------- |
| `Send(data string) error`         | Send a simple SSE data event        |
| `SendEvent(event SseEvent) error` | Send a structured SSE event         |
| `Close()`                         | Close the writer and stop heartbeat |

### SseEvent[​](#sseevent "Direct link to SseEvent")

| Field   | Type   | Description               |
| ------- | ------ | ------------------------- |
| `Event` | string | Event type name           |
| `Data`  | string | Event payload             |
| `Id`    | string | Event ID for reconnection |
| `Retry` | int    | Reconnection hint (ms)    |

## Error handling[​](#error-handling "Direct link to Error handling")

### [WithErrorHandler()](https://github.com/gosoline-project/httpserver/blob/main/server_options.go)[​](#witherrorhandler "Direct link to witherrorhandler")

```
httpserver.WithErrorHandler(func(statusCode int, err error) any {

    return map[string]any{"error": err.Error()}

})
```

Pass this option to `NewServer` or `NewServerWithSettings`, or add it to `ServerDefinition.Options` for `RunServers`. The callback runs after middleware selects a status from the last recorded error.

Status selection checks `ErrorWithStatus` first, then the first matching `WithErrorMapper`. Validation errors use `400 Bad Request`. All other errors use `500 Internal Server Error`.

The callback returns the body to send to the client, not another Go error. The endpoint still returns its error through its normal `(output, error)` result. A returned struct or map keeps the selected HTTP status. Private 5xx privacy passes `internal server error` to the callback.

A struct, map, or other serializable value that does not implement `httpserver.Response` uses the server's response negotiator. It selects a format from `Accept`, serializes the value, and sets `Content-Type`. JSON is the default. XML needs a negotiator that includes `XMLRepresentation()` and a value that `encoding/xml` can encode. In the pinned package version, error middleware falls back to JSON with the selected status if negotiation or encoding fails.

An explicit `httpserver.Response` bypasses negotiation and sends its format regardless of `Accept`. It owns the status and headers, including `Content-Type`. It does not inherit the selected error status. For fixed JSON with that status, return `httpserver.NewJsonResponse(body, httpserver.WithStatusCode(statusCode))`.

### [NewErrorWithStatus()](https://github.com/gosoline-project/httpserver/blob/main/error.go)[​](#newerrorwithstatus "Direct link to newerrorwithstatus")

Return a zero output and a wrapped error to set an expected client-error status:

```
return nil, httpserver.NewErrorWithStatus(http.StatusBadRequest, err)
```

For a struct output, replace `nil` with that struct's zero value.

Middleware can attach the same wrapped error to the Gin context:

```
ginCtx.Error(httpserver.NewErrorWithStatus(http.StatusBadRequest, err))
```

## Middleware[​](#middleware "Direct link to Middleware")

### [Cors()](https://github.com/gosoline-project/httpserver/blob/main/cors.go)[​](#cors "Direct link to cors")

```
corsMiddleware, err := httpserver.Cors(config, "default")
```

Creates a CORS middleware from `httpserver.<name>.cors` settings.

### [CorsFactory()](https://github.com/gosoline-project/httpserver/blob/main/cors.go)[​](#corsfactory "Direct link to corsfactory")

```
router.UseFactory(httpserver.CorsFactory)
```

Creates CORS middleware as a settings-aware middleware factory. The factory uses `settings.Name`, so reusable route setup automatically reads the CORS settings for the current named server.

### [CreateEmbeddedStaticServe()](https://github.com/gosoline-project/httpserver/blob/main/middleware_embedded_static_serve.go)[​](#createembeddedstaticserve "Direct link to createembeddedstaticserve")

```
router.UseFactory(httpserver.CreateEmbeddedStaticServe(fs, "public", "/api"))
```

Creates a middleware that serves embedded static files with SPA fallback.

### Built-in middleware[​](#built-in-middleware "Direct link to Built-in middleware")

The following middleware is automatically applied to every server:

| Middleware           | Purpose                                                                                                                    |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| Sampling             | Apply sampling decisions                                                                                                   |
| Metric               | Record request-duration and response-status observations, request/rejection counters, and active-request/connection gauges |
| Logging              | Log requests (fingers-crossed)                                                                                             |
| Compression          | Gzip compression                                                                                                           |
| Max body size        | Limit incoming request bodies                                                                                              |
| Error                | Catch and format errors                                                                                                    |
| Recovery             | Panic recovery                                                                                                             |
| Connection lifecycle | Manage connection age                                                                                                      |

## Metrics[​](#metrics "Direct link to Metrics")

**Source status:** The HTTP duration boundaries and server-wide CloudWatch rejection projection described below are present in the published HTTPServer pseudo-version pinned by this documentation, `v0.6.5-0.20261026100922-13fdf9916b14`. This describes module source only; it does not claim that any application has adopted or deployed this version.

The two request counters count matched requests after the middleware chain unwinds; the two response counters count matched responses at that point. Matched requests aborted by authentication, concurrency, or chaos middleware are included, while unmatched paths emit none. `http.server.requests` has a zero-valued default per configured server; `http.server.routed.requests` has one per registered application route/method pair. The separately registered `/health` route has no pre-request default. The response counters have no zero defaults. The duration histogram retains the full status code without a class. The detailed rejection Counter remains statusless and route-scoped, with its existing registered-route zero defaults; CloudWatch also receives a same-name server-wide `KindTotal` `Count` projection with only `{name}` and a zero default per configured server. The projection is CloudWatch-only and is not a second OTEL or Prometheus series. See [Metric schema](/docs/pr-5/reference/metric-schema/.md) for the v1→v2 shapes. No additional instrumentation hook is required.

| Metric                             | Instrument                        | Authored unit | Dimensions                                                                                                                                                                                                                         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| ---------------------------------- | --------------------------------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `http.server.requests`             | monotonic counter (high priority) | Count         | `name` — configured server name                                                                                                                                                                                                    | One count for each matched request, including requests aborted after route matching. Unmatched requests are excluded. Configured servers receive a zero-valued default.                                                                                                                                                                                                                                                                                  |
| `http.server.routed.requests`      | monotonic counter (high priority) | Count         | `name` — configured server name; `route` — normalized Gin route template; `request.method` — HTTP method                                                                                                                           | One count for each matched route, including `/health` and requests aborted after route matching. Unmatched requests are excluded. Registered application route/method pairs receive zero-valued defaults; the separately registered `/health` route has no pre-request zero default.                                                                                                                                                                     |
| `http.server.responses`            | monotonic counter (high priority) | Count         | `name` — configured server name; `response.status_code` — decimal numeric status; `response.status_class` — `1xx` through `5xx` (default outside 100–599)                                                                          | One per matched response after the middleware chain unwinds, including responses from aborted requests and matched `/health`. Unmatched routes are excluded. No zero-valued defaults.                                                                                                                                                                                                                                                                    |
| `http.server.routed.responses`     | monotonic counter (high priority) | Count         | `name` — configured server name; `route` — normalized Gin route template; `request.method` — HTTP method; `response.status_code` — decimal numeric status; `response.status_class` — `1xx` through `5xx` (default outside 100–599) | One per matched response after the middleware chain unwinds, including responses from aborted requests and matched `/health`. Unmatched routes are excluded. No zero-valued defaults.                                                                                                                                                                                                                                                                    |
| `http.server.request.duration`     | histogram                         | ms            | `name` — configured server name; `route` — normalized Gin route template; `request.method` — HTTP method; `response.status_code` — decimal response status                                                                         | Duration of a matched HTTP request. The histogram retains the full response code but does not carry `response.status_class`; keep it for latency distributions.                                                                                                                                                                                                                                                                                          |
| `http.server.rejected.requests`    | counter                           | count         | `name` — configured server name; `route` — registered Gin route; `request.method` — HTTP method                                                                                                                                    | Detailed raw Counter for requests explicitly rejected by the concurrency middleware. Ordinary 4xx/5xx responses are not necessarily rejections; the Counter has no response-status dimension, and registered routes receive zero-valued defaults. CloudWatch also receives a same-name server-wide `KindTotal` Count projection keyed only by `name`, with a zero-valued default per configured server. The projection has no OTEL or Prometheus series. |
| `http.server.active_request.count` | gauge                             | count         | `name` — configured server name                                                                                                                                                                                                    | Current number of tracked active requests for the named server. Updated on request start/completion and sampled at startup, every 10 seconds, and during shutdown.                                                                                                                                                                                                                                                                                       |
| `http.server.connection.count`     | gauge                             | count         | `name` — configured server name                                                                                                                                                                                                    | Current number of tracked open connections for the named server. Updated on connection state changes and sampled at startup, every 10 seconds, and during shutdown.                                                                                                                                                                                                                                                                                      |

The published HTTPServer module version documented by this page uses the HTTP semantic convention's recommended duration boundaries: `[0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10]` seconds ([HTTP metrics](https://opentelemetry.io/docs/specs/semconv/http/http-metrics/)). The middleware authors the equivalent `[5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000]` millisecond bounds; the writers scale those explicit bounds and observations to seconds for export. The millisecond representation preserves these recommended values exactly. Buckets set histogram resolution rather than exact quantiles: values between bounds share a bucket, and values above 10 seconds fall beyond the final finite bound.

Metric export requires the application metric daemon: wire `application.WithMetrics`, set `metric.enabled: true`, and configure at least one writer under `metric.writers`. These are authored Gosoline names and units; see [Metric schema](/docs/pr-5/reference/metric-schema/.md) for writer-specific names and duration conversion.

The documentation's HTTPServer pin is `v0.6.5-0.20261026100922-13fdf9916b14`; its [metric middleware source](https://github.com/gosoline-project/httpserver/blob/13fdf9916b1462b89562b458ccfb03cdda72e1a9/middleware_metric.go) includes the standard HTTP duration profile and server-wide CloudWatch rejection projection described above. The library and the Gosoline SDK's separately implemented [HTTP server package](/docs/pr-5/reference/package-gosoline-httpserver/.md) use the same eight canonical names and dimension keys. If both implementations send identical dimension values to the same backend, the series have no source-module dimension that distinguishes them. For backend-specific projection behavior and CloudWatch queries, see [Metric schema](/docs/pr-5/reference/metric-schema/.md). For operational details, see [Add middleware](/docs/pr-5/how-to/http-server/add-middleware/.md) and [Concurrency and connection pressure](/docs/pr-5/how-to/http-server/concurrency-and-connection-pressure/.md).

## Validation[​](#validation "Direct link to Validation")

### [AddCustomValidators()](https://github.com/gosoline-project/httpserver/blob/main/custom_validators.go)[​](#addcustomvalidators "Direct link to addcustomvalidators")

```
httpserver.AddCustomValidators([]httpserver.CustomValidator{

    {Name: "phone", Validator: validatePhone},

})
```

Register custom field validators for use in `binding` tags.

### Related functions[​](#related-functions "Direct link to Related functions")

* `AddStructValidators([]StructValidator)` — Register whole-struct validators
* `AddCustomTypeFuncs([]CustomTypeFunc)` — Register custom type functions
* `AddValidateAlias([]ValidateAlias)` — Register validation tag aliases

## Configuration[​](#configuration "Direct link to Configuration")

### Settings[​](#settings "Direct link to Settings")

Config key: `httpserver.<name>`

| Field                  | Type                               | Default     | Description                                                          |
| ---------------------- | ---------------------------------- | ----------- | -------------------------------------------------------------------- |
| `port`                 | string                             | `"8080"`    | Listening port. Use `"0"` for random port.                           |
| `mode`                 | string                             | `"release"` | Gin mode: `debug`, `release`, `test`.                                |
| `compression`          | CompressionSettings                | -           | Gzip configuration.                                                  |
| `cors`                 | CorsSettings                       | -           | CORS middleware configuration.                                       |
| `timeout`              | TimeoutSettings                    | -           | IO timeouts.                                                         |
| `logging`              | LoggingSettings                    | -           | Request logging configuration.                                       |
| `errors`               | ErrorsSettings                     | -           | Error response privacy configuration.                                |
| `router`               | RouterSettings                     | -           | Router configuration.                                                |
| `max_body_bytes`       | int                                | `10485760`  | Maximum incoming request body size in bytes. `0` disables the limit. |
| `concurrency`          | ConcurrencySettings                | -           | Request and connection pressure limits.                              |
| `connection_lifecycle` | ConnectionLifeCycleAdvisorSettings | -           | Connection age and request count limits.                             |
| `chaos`                | ChaosSettings                      | -           | Chaos middleware for resilience testing.                             |

### TimeoutSettings[​](#timeoutsettings "Direct link to TimeoutSettings")

| Field      | Type     | Default | Description                                      |
| ---------- | -------- | ------- | ------------------------------------------------ |
| `read`     | duration | `60s`   | Max duration for reading entire request. Min 1s. |
| `write`    | duration | `60s`   | Max duration for writing response. Min 1s.       |
| `idle`     | duration | `60s`   | Max idle time for keep-alive. Min 1s.            |
| `drain`    | duration | `0s`    | Wait time before graceful shutdown.              |
| `shutdown` | duration | `60s`   | Max time for graceful shutdown. Min 1s.          |

### CompressionSettings[​](#compressionsettings "Direct link to CompressionSettings")

| Field           | Type                       | Default     | Description                                             |
| --------------- | -------------------------- | ----------- | ------------------------------------------------------- |
| `level`         | string                     | `"default"` | `"none"`, `"default"`, `"fast"`, `"best"`, `"0"`-`"9"`. |
| `decompression` | bool                       | `true`      | Decompress gzip request bodies.                         |
| `exclude`       | CompressionExcludeSettings | -           | Paths/extensions to exclude.                            |

### CompressionExcludeSettings[​](#compressionexcludesettings "Direct link to CompressionExcludeSettings")

| Field        | Type         | Description                 |
| ------------ | ------------ | --------------------------- |
| `extension`  | string array | File extensions to exclude. |
| `path`       | string array | Exact paths to exclude.     |
| `path_regex` | string array | Regex patterns to exclude.  |

### CorsSettings[​](#corssettings "Direct link to CorsSettings")

| Field                    | Type         | Description                                                 |
| ------------------------ | ------------ | ----------------------------------------------------------- |
| `allowed_origin_pattern` | string       | Regular expression matched against the full `Origin` value. |
| `allowed_headers`        | string array | Allowed request headers for CORS requests.                  |
| `allowed_methods`        | string array | Allowed HTTP methods for CORS requests.                     |

The origin pattern is anchored internally. For example, `https://example\.com` allows `https://example.com`, but not `https://example.com.evil.com`.

### LoggingSettings[​](#loggingsettings "Direct link to LoggingSettings")

| Field                 | Type         | Default | Description                                                              |
| --------------------- | ------------ | ------- | ------------------------------------------------------------------------ |
| `request_body`        | bool         | `false` | Log request bodies.                                                      |
| `request_body_base64` | bool         | `false` | Base64-encode request body in logs.                                      |
| `request_headers`     | string array | `[]`    | Header names to include in logs. Only the configured headers are logged. |

### ErrorsSettings[​](#errorssettings "Direct link to ErrorsSettings")

| Field     | Type   | Default     | Description                                                                          |
| --------- | ------ | ----------- | ------------------------------------------------------------------------------------ |
| `privacy` | string | `"private"` | `"private"` hides internal error details for 5xx responses. `"public"` exposes them. |

### RouterSettings[​](#routersettings "Direct link to RouterSettings")

| Field          | Type | Default | Description                                            |
| -------------- | ---- | ------- | ------------------------------------------------------ |
| `use_raw_path` | bool | `false` | Use the raw URL path (not decoded) for route matching. |

### ConcurrencySettings[​](#concurrencysettings "Direct link to ConcurrencySettings")

| Field                  | Type     | Default | Description                                                                        |
| ---------------------- | -------- | ------- | ---------------------------------------------------------------------------------- |
| `max_requests`         | int      | `0`     | Maximum concurrent requests. `0` disables the limit.                               |
| `max_connections`      | int      | `0`     | Target maximum open TCP connections. `0` disables connection pressure management.  |
| `overload_status_code` | int      | `503`   | HTTP status code returned when `max_requests` is reached.                          |
| `retry_after`          | duration | `0`     | Value for the `Retry-After` response header when overloaded. `0` omits the header. |

See [Concurrency and connection pressure](/docs/pr-5/how-to/http-server/concurrency-and-connection-pressure/.md) for detailed behaviour.

### ConnectionLifeCycleAdvisorSettings[​](#connectionlifecycleadvisorsettings "Direct link to ConnectionLifeCycleAdvisorSettings")

| Field               | Type     | Default | Description                                                                            |
| ------------------- | -------- | ------- | -------------------------------------------------------------------------------------- |
| `enabled`           | bool     | `true`  | Enable connection lifecycle management.                                                |
| `max_age`           | duration | `1m`    | Max connection age before signalling close. `0` disables age-based closing.            |
| `max_request_count` | int      | `0`     | Max requests per connection before signalling close. `0` disables count-based closing. |

See [Concurrency and connection pressure](/docs/pr-5/how-to/http-server/concurrency-and-connection-pressure/.md) for detailed behaviour.

### ChaosSettings[​](#chaossettings "Direct link to ChaosSettings")

| Field           | Type                      | Default | Description                   |
| --------------- | ------------------------- | ------- | ----------------------------- |
| `enabled`       | bool                      | `false` | Enable chaos middleware.      |
| `reject`        | ChaosRejectSettings       | -       | Random HTTP error responses.  |
| `delay`         | ChaosDelaySettings        | -       | Random pre-processing delays. |
| `drop`          | ChaosDropSettings         | -       | Random TCP connection drops.  |
| `slow_response` | ChaosSlowResponseSettings | -       | Random trickle responses.     |
| `truncate`      | ChaosTruncateSettings     | -       | Random response truncation.   |

### ChaosRejectSettings[​](#chaosrejectsettings "Direct link to ChaosRejectSettings")

| Field          | Type      | Default                     | Description                         |
| -------------- | --------- | --------------------------- | ----------------------------------- |
| `percent`      | int       | `3`                         | Probability (0–100) of rejection.   |
| `status_codes` | int array | `[499, 500, 502, 503, 504]` | Status codes to randomly pick from. |

### ChaosDelaySettings[​](#chaosdelaysettings "Direct link to ChaosDelaySettings")

| Field          | Type     | Default | Description                   |
| -------------- | -------- | ------- | ----------------------------- |
| `percent`      | int      | `3`     | Probability (0–100) of delay. |
| `min_duration` | duration | `0`     | Minimum delay.                |
| `max_duration` | duration | `60s`   | Maximum delay.                |

### ChaosDropSettings[​](#chaosdropsettings "Direct link to ChaosDropSettings")

| Field     | Type | Default | Description                             |
| --------- | ---- | ------- | --------------------------------------- |
| `percent` | int  | `3`     | Probability (0–100) of connection drop. |

### ChaosSlowResponseSettings[​](#chaosslowresponsesettings "Direct link to ChaosSlowResponseSettings")

| Field        | Type     | Default | Description                                |
| ------------ | -------- | ------- | ------------------------------------------ |
| `percent`    | int      | `3`     | Probability (0–100) of throttled response. |
| `delay`      | duration | `1s`    | Pause between chunks.                      |
| `chunk_size` | int      | `64`    | Bytes per chunk.                           |

### ChaosTruncateSettings[​](#chaostruncatesettings "Direct link to ChaosTruncateSettings")

| Field       | Type | Default | Description                         |
| ----------- | ---- | ------- | ----------------------------------- |
| `percent`   | int  | `3`     | Probability (0–100) of truncation.  |
| `max_bytes` | int  | `512`   | Maximum body bytes before dropping. |

See the [Chaos middleware how-to](/docs/pr-5/how-to/http-server/chaos-middleware/.md) for usage examples and safety notes.
