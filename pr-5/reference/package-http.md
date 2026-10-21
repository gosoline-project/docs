# http package

The Gosoline SDK's `pkg/http` package (import path `github.com/justtrackio/gosoline/pkg/http`) provides a configured, reusable client for outbound HTTP requests. Its `Client` interface builds requests through `NewRequest()` and exposes `Get`, `Post`, `Put`, `Patch`, and `Delete` operations.

`ProvideHttpClient(ctx, config, logger, name)` obtains the named client from the application context. The client writes its request-duration observations to the `http.client` Gosoline metric namespace.

## Metrics[​](#metrics "Direct link to Metrics")

Each validated client operation increments `http.client.requests` immediately before `req.Execute`; request construction and URL-validation errors return earlier and are not counted. `http.client.completed.requests` increments after `Execute` returns in the response, transport-error, or cancellation branch. The separate `http.client.request.duration` histogram retains latency observations; use the Counters for starts and completions, not histogram `_count` or `_sum`.

| Metric                           | Instrument | Authored unit | Dimensions                                                                                                                                                                                                                                                                        | Description                                                                                                                                                                                                                         |
| -------------------------------- | ---------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `http.client.requests`           | Counter    | count         | `request.method` — requested HTTP method                                                                                                                                                                                                                                          | One validated request immediately before `Execute`; no zero-valued defaults.                                                                                                                                                        |
| `http.client.completed.requests` | Counter    | count         | `request.method`; `response.status_code` — decimal numeric status; `response.status_class` — `1xx` through `5xx` (default outside 100–599); `error.type`                                                                                                                          | One after `Execute` returns. Responses carry the full code, bounded class, and default `error.type`; transport errors/cancellations carry default code and class with `request_failed`/`request_canceled`. No zero-valued defaults. |
| `http.client.request.duration`   | histogram  | ms            | `request.method` — requested HTTP method; `response.status_code` — decimal response status, or `{{default}}` when Execute returns an error; `error.type` — `{{default}}` for a response, `request_canceled` for a canceled request, or `request_failed` for another Execute error | Duration of a client request. The latency histogram has no `response.status_class` dimension.                                                                                                                                       |

When the response code or class is missing, the default marker is rendered per writer: as an empty Prometheus label value, an absent OpenTelemetry attribute, or an omitted CloudWatch dimension. See [Metric schema](/docs/pr-5/reference/metric-schema/.md) for backend-specific selection examples.

Metric export is an application-level concern: wire `application.WithMetrics`, set `metric.enabled: true`, and configure at least one writer under `metric.writers`. The client does not have a separate metric-emission switch. See [Metric schema](/docs/pr-5/reference/metric-schema/.md) for canonical-name and writer conversion details.

## Related references[​](#related-references "Direct link to Related references")

* [Gosoline SDK HTTP server package](/docs/pr-5/reference/package-gosoline-httpserver/.md)
* [Published HTTP server package](/docs/pr-5/reference/package-httpserver/.md)
