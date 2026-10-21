# Gosoline SDK HTTP server package

The Gosoline SDK's `pkg/httpserver` package integrates an HTTP server with the application kernel. Import it as `github.com/justtrackio/gosoline/pkg/httpserver`; `httpserver.New(name, definer)` returns a kernel module factory and reads server configuration from `httpserver.<name>`. The SDK installs its request-metric middleware and server metric recorder as part of constructing the server.

This is distinct from the published `github.com/gosoline-project/httpserver` module documented in the [published HTTP server reference](/docs/pr-5/reference/package-httpserver/.md). The modules are separate implementations, though they currently author the same four canonical metric names and dimensions.

## Metrics[​](#metrics "Direct link to Metrics")

The request-duration histogram records matched routes; its observation count is the request count for each dimension set, not a separate status metric. Rejected-request observations are emitted when the concurrency middleware marks a request rejected; an ordinary 4xx or 5xx response is not by itself a rejection. The rejection counter is initialized with zero-valued defaults for registered routes and does not include a response-status dimension.

| Metric                             | Instrument | Authored unit | Dimensions                                                                                                                                                 | Description                                                                                                                                                                                                          |
| ---------------------------------- | ---------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `http.server.request.duration`     | histogram  | ms            | `name` — configured server name; `route` — normalized Gin route template; `request.method` — HTTP method; `response.status_code` — decimal response status | Duration of a matched HTTP request. The route, method, and response status distinguish observations.                                                                                                                 |
| `http.server.rejected.requests`    | counter    | count         | `name` — configured server name; `route` — normalized Gin route template; `request.method` — HTTP method                                                   | Requests explicitly rejected by the concurrency middleware. Ordinary 4xx/5xx responses are not necessarily rejections; this metric has no response-status dimension. Registered routes receive zero-valued defaults. |
| `http.server.active_request.count` | gauge      | count         | `name` — configured server name                                                                                                                            | Current number of requests tracked as active for the named server. Updated on request start/completion and sampled at startup, every 10 seconds, and during shutdown.                                                |
| `http.server.connection.count`     | gauge      | count         | `name` — configured server name                                                                                                                            | Current number of tracked open connections for the named server. Updated on connection state changes and sampled at startup, every 10 seconds, and during shutdown.                                                  |

The SDK installs the metric middleware by default; exporting the observations still requires the application metric daemon. Wire `application.WithMetrics`, set `metric.enabled: true`, and configure at least one writer under `metric.writers`. The table shows canonical Gosoline names and authored units; see [Metric schema](/docs/pr-5/reference/metric-schema/.md) for writer-specific names and duration conversion.

If this SDK server and the published HTTP server both write to the same backend with identical dimension values, the metrics do not carry a source-module dimension that separates the implementations.

## Related guides[​](#related-guides "Direct link to Related guides")

* [Add HTTP server middleware](/docs/pr-5/how-to/http-server/add-middleware/.md)
* [Concurrency and connection pressure](/docs/pr-5/how-to/http-server/concurrency-and-connection-pressure/.md)
* [Published HTTP server package](/docs/pr-5/reference/package-httpserver/.md)
