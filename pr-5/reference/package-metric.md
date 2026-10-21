# metric package

The Gosoline `metric` package provides the framework metric writer pipeline and an optional Prometheus scrape endpoint. It also supplies a logger handler that records warn- and error-level log records as a framework metric. The [Metric schema reference](/docs/pr-5/reference/metric-schema/.md) remains the source of truth for canonical Gosoline metric names, units, dimensions, writer transformations, and migration history; this page covers package activation and endpoint-specific exports.

## Registering metrics and the Prometheus endpoint[​](#registering-metrics-and-the-prometheus-endpoint "Direct link to Registering metrics and the Prometheus endpoint")

Register `application.WithMetrics` on the application to install the metric daemon and Prometheus metrics-server modules ([source](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/application/options.go)). The metric daemon is disabled by default, so set `metric.enabled: true` to export framework metric data. The scrape server starts only when the `prometheus` writer is selected and its API is enabled; API defaults are enabled, port `8092`, and path `/metrics` ([server source](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/metric/prometheus_metric_server.go)).

```
metric:

  enabled: true

  writers:

    - prometheus

  writer_settings:

    prometheus:

      api:

        enabled: true

        port: 8092

        path: /metrics
```

The `api.enabled` switch controls the scrape listener separately from `metric.enabled`: disabling the API removes the endpoint, while `metric.enabled` controls the metric daemon. CloudWatch and OTEL writers do not add writer-owned operational metrics to the Gosoline contract. The Prometheus writer also adds no Gosoline-authored operational metric names; the endpoint and its upstream collectors expose the separate library-owned data described below.

To count logger records, also register `application.WithLoggerMetricHandler` ([source](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/application/options.go)). This attaches the metric handler to the application logger; it is separate from enabling the metric daemon and selecting an exporter. See the [log package reference](/docs/pr-5/reference/package-log/.md) for logger and handler setup.

## Metrics[​](#metrics "Direct link to Metrics")

### Gosoline-authored metric[​](#gosoline-authored-metric "Direct link to Gosoline-authored metric")

This framework metric is included in the current 80-name inventory in the [Metric schema reference](/docs/pr-5/reference/metric-schema/.md).

| Metric               | Instrument | Authored unit | Dimensions                  | Description                                                                                                                              |
| -------------------- | ---------- | ------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `metric.log.records` | counter    | count         | `level` = `warn` or `error` | Counts only warn- and error-level records delivered to the metric logger handler, separated by level. Other log levels are filtered out. |

The handler declares count-counter defaults for both levels and writes one increment for each matching record it receives ([handler source](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/metric/logger_handler.go)).

### Prometheus HTTP-handler metrics (upstream-owned)[​](#prometheus-http-handler-metrics-upstream-owned "Direct link to Prometheus HTTP-handler metrics (upstream-owned)")

When the Prometheus API is enabled, Gosoline wraps the registry handler with `promhttp.InstrumentMetricHandler`. In the pinned Prometheus Go client `client_golang` v1.20.5, that instrumentation adds these two exports. They describe requests to the metrics handler, not application HTTP-server traffic, and are not Gosoline-authored metrics or part of the 80-name inventory.

| Metric                                       | Instrument | Upstream unit | Dimensions                | Description                                                                |
| -------------------------------------------- | ---------- | ------------- | ------------------------- | -------------------------------------------------------------------------- |
| `promhttp_metric_handler_requests_total`     | counter    | count         | `code` (HTTP status code) | Requests handled by the metrics endpoint, grouped by response status code. |
| `promhttp_metric_handler_requests_in_flight` | gauge      | count         | None                      | Requests currently being handled by the metrics endpoint.                  |

The endpoint passes an empty `promhttp.HandlerOpts`; the optional `promhttp_metric_handler_errors_total` metric is therefore not registered by this path. See the pinned [`promhttp.InstrumentMetricHandler` contract](https://pkg.go.dev/github.com/prometheus/client_golang@v1.20.5/prometheus/promhttp#InstrumentMetricHandler).

### Upstream Go and process collectors[​](#upstream-go-and-process-collectors "Direct link to Upstream Go and process collectors")

The default Prometheus registry also registers the upstream Go runtime and process collectors. Their exported metric families depend on the pinned client version, Go runtime/toolchain, and process platform; they are external collector output, not Gosoline-authored metrics, and are intentionally not enumerated as a fixed Gosoline metric set. Consult the pinned [`NewGoCollector`](https://pkg.go.dev/github.com/prometheus/client_golang@v1.20.5/prometheus/collectors#NewGoCollector) and [`NewProcessCollector`](https://pkg.go.dev/github.com/prometheus/client_golang@v1.20.5/prometheus/collectors#NewProcessCollector) documentation for that upstream contract.

`metric.schema_version` is emitted as schema metadata, not as a metric. Likewise, the two `promhttp` metrics and the Go/process collector exports above are outside the Gosoline-authored 80-name inventory. For canonical writer naming and unit conversion, see [Metric schema](/docs/pr-5/reference/metric-schema/.md); for logger integration, see [Package log](/docs/pr-5/reference/package-log/.md).
