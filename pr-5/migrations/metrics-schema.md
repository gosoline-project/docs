# Migrate from metrics schema v1 to v2

This guide takes an application from metrics schema `v1.0` (Gosoline `v0.68.3`) to `v2.0` ([Gosoline `v0.69.0`](https://github.com/justtrackio/gosoline/releases/tag/v0.69.0)). The metrics schema version is separate from the Go module version. The metric daemon reports it as `metric.schema_version` in application metadata when enabled. Gosoline does not emit both schemas or provide aliases for the old names.

## 1. Find what uses the old metrics[​](#1-find-what-uses-the-old-metrics "Direct link to 1. Find what uses the old metrics")

* Check the Gosoline version in the application and the metrics each deployed writer exports. If a separate module emits metrics, check its version and output too.
* Find metric producers, names, and dimensions in Go code and tests. Search application config, dashboards, recording rules, alerts, and autoscaling policy definitions, including those generated from templates or maintained in a separate repository.
* For each query, record what it measures, its unit, which labels it selects, and how it handles missing samples. Use actual exporter output for the old names; source names and exported names are not always the same.

## 2. Update Go code and configuration[​](#2-update-go-code-and-configuration "Direct link to 2. Update Go code and configuration")

Upgrade the Gosoline dependency in your application:

```
go get github.com/justtrackio/gosoline@v0.69.0
```

If you use older versions of independently released producers, upgrade them in the same `go get` command. Compatible releases include `github.com/gosoline-project/httpserver@v0.6.5` and `github.com/gosoline-project/sqlc@v0.5.0`. Include only the modules your application uses. Keep newer compatible versions and check the selected versions with `go list -m all`.

* Update `metric.NewWriter` and `metric.NewWriterWithInterfaces` to take a namespace. Use `metric.NewWriter("", defaults...)` to keep existing custom names as full leaves. To use namespace `orders`, set the datum leaf to `created`, not `orders.created`. A writer supplies its namespace only when `Datum.Namespace` is empty. For writes directly to a backend writer, specify the intended namespace, instrument kind, and unit on the datum.
* Update calls to `metric.NewPrometheusWriterWithInterfaces`: remove the namespace argument. Remove `metric.writer_settings.prometheus.naming.namespace_delimiter` and `metric.writer_settings.prometheus.naming.namespace_pattern` from config. The prefix is `gosoline`; select the application with scrape-target labels. The CloudWatch namespace setting is unchanged by this schema migration.
* If you call `kafka.MetricPair`, pass a `kafka.MetricSpec` instead of positional arguments. Use `kafka.DimensionClientType` with `kafka.ClientTypeConsumer` or `kafka.ClientTypeProducer`. See the [release notes](https://github.com/justtrackio/gosoline/releases/tag/v0.69.0) for the complete list of Go API and constant changes.

The metric calculator was removed in Gosoline `v0.69.0`.

## 3. Update queries and alerts[​](#3-update-queries-and-alerts "Direct link to 3. Update queries and alerts")

The left column lists old source-level names, not literal exporter names. Look up the matching names in your backend. The right column uses the [canonical names in the metrics reference](/docs/pr-5/reference/metric-schema/.md).

| Previous metric or family                                               | What to query after the upgrade                                                                                                                                                                                                                                |
| ----------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HTTP server `HttpRequestCount` and `HttpRequestCountPerRoute`           | `http.server.requests` for server counts and `http.server.routed.requests` for route/method counts. Replace `ServerName`, `Path`, and `Method` with `name`, `route`, and `request.method` where applicable.                                                    |
| HTTP server `HttpStatus1XX` through `HttpStatus5XX`                     | `http.server.responses` or `http.server.routed.responses` for counts by `response.status_code` or `response.status_class`. See [CloudWatch server response queries](/docs/pr-5/reference/metric-schema/.md#cloudwatch-server-response-queries) for 5xx alarms. |
| `HttpRequestResponseTime` and `HttpRequestResponseTimePerRoute`         | Use `http.server.request.duration`. For a server-wide Prometheus percentile, sum bucket rates by `name` and `le` before applying `histogram_quantile`. For a per-route view, also keep `route` and `request_method`.                                           |
| HTTP client `HttpClientRequest` and `HttpClientError`                   | `http.client.requests` for request starts. Use `http.client.completed.requests` with `error.type="request_failed"` for transport failures. Cancellations have a separate value, `request_canceled`.                                                            |
| Kafka `RecordsConsumedFailed` and `RecordsSentFailed`                   | Select `processing_failed` on `kafka.consumer.consumed.messages` or `send_failed` on `kafka.producer.sent.messages`. In each metric the failure and default series are disjoint.                                                                               |
| Logger leaves `warn` and `error`                                        | `metric.log.records`, filtered by the `level` dimension.                                                                                                                                                                                                       |
| Repository `DbAccessLatency` and `DdbAccessLatency`                     | `db.repo.operation.duration` for SQL and `ddb.operation.duration` for DynamoDB. Use `db.repo.operations` or `ddb.operations` for operation counts, not histogram samples.                                                                                      |
| Stream `ProcessedCount`, `stream.Error`, and `stream.UnknownModelError` | `stream.consumed.messages`: the default series counts collected input, not successful processing. Select `processing_failed` or `model_lookup_failed` directly; the diagnostic series can overlap each other and the total.                                    |

Use the [release notes](https://github.com/justtrackio/gosoline/releases/tag/v0.69.0) for other former metric families and the [metrics reference](/docs/pr-5/reference/metric-schema/.md) for current measurement details. Check the operation boundary before treating any mapping as equivalent.

### Check writer-specific names, units, and dimensions[​](#check-writer-specific-names-units-and-dimensions "Direct link to Check writer-specific names, units, and dimensions")

* **CloudWatch:** Use the [rendered metric name](/docs/pr-5/reference/metric-schema/.md#authored-names-and-writer-output), the configured namespace, and the exact dimension set. Choose either a detailed series or a CloudWatch-only aggregate of the same observations; do not sum both.
* **OpenTelemetry:** Framework names have a `gosoline.` prefix. Default dimensions such as a successful operation's `error.type` are absent, not literal markers. Check the unit on the instrument.
* **Prometheus:** Use the rendered `gosoline_` name, the counter or unit suffix where applicable, and underscore labels such as `error_type`. To select a default error dimension, use `error_type=""`.
* **Latency:** In Prometheus and OpenTelemetry, millisecond measurements export in seconds. Review thresholds and use histogram queries for framework duration metrics that were summaries. Existing samples are not rebucketed. Use a dedicated counter for event volume when the histogram measures a different event.

## 4. Verify before rollout[​](#4-verify-before-rollout "Direct link to 4. Verify before rollout")

1. Build and run the affected application tests. Update assertions that depend on metric names, dimensions, units, or types. Keep application behavior unchanged unless the migration requires it.
2. Run the application in a test environment with its configured writers. Generate representative success and failure traffic. Inspect the exported samples, check `metric.schema_version` when the metric daemon is enabled, and confirm that dashboard queries and alert expressions select the expected series and units.
3. Test alarms for missing data as well as threshold crossings. Coordinate the query and deployment changes: a mixed rollout can have both schemas across instances, but one instance does not write aliases for the other. Keep the prior query configuration available if you need to roll back.

## Copyable AI migration instructions[​](#copyable-ai-migration-instructions "Direct link to Copyable AI migration instructions")

The instructions ask an agent to inspect the application and its monitoring definitions, verify each mapping, and report anything it cannot verify.

Copy AI migration instructions
