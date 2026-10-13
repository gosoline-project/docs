# Metric schema

This page is the contract for metrics authored by Gosoline. The canonical names, instrument types, authored units, dimensions, and descriptions below describe what the emitters actually record; they are not a copy of OpenTelemetry's metric inventory. Gosoline follows OpenTelemetry semantic conventions for metric grammar, units, and attribute shape where useful, but emits its own names. For example, `http.server.request.duration` is authored by Gosoline and exported by OTEL as `gosoline.http.server.request.duration`, not as an unqualified semantic-convention metric.

## Schema version[​](#schema-version "Direct link to Schema version")

The current metric contract is **`v2.0`** (`metric.SchemaVersion`). When the metric daemon is enabled, Gosoline publishes that value as metadata under `metric.schema_version`. Removing or renaming an authored metric name, dimension key, or unit representation is a breaking change; purely additive metric changes are minor. The current V2 release is the single migration from V1 and emits no dual versions. A new build's schema version does not automatically migrate dashboards, alerts, queries, or autoscaling policies.

## Authored names and writer output[​](#authored-names-and-writer-output "Direct link to Authored names and writer output")

A Gosoline metric is authored as a canonical namespace and leaf, joined with a dot—for example, `http.server` + `request.duration` becomes `http.server.request.duration`. The namespace identifies the emitting subsystem; it is not the application identity. A namespace-less application metric instead uses its leaf as its complete authored name.

Each writer renders that canonical name for its backend:

| Authored metric                           | CloudWatch metric name      | Prometheus metric name                          | OTEL metric name                        |
| ----------------------------------------- | --------------------------- | ----------------------------------------------- | --------------------------------------- |
| `http.server.request.duration`            | `HttpServerRequestDuration` | `gosoline_http_server_request_duration_seconds` | `gosoline.http.server.request.duration` |
| Namespace-less counter `my.custom.metric` | `MyCustomMetric`            | `gosoline_my_custom_metric_total`               | `my.custom.metric`                      |

The first row's duration is authored in milliseconds; Prometheus and OTEL values are scaled to seconds, and both names show the seconds base unit. The Prometheus histogram exposition adds its normal sample suffixes to the base name. Prometheus dimension keys also replace dots with underscores, so the HTTP `request.method`, `route`, and `response.status_code` attributes become `request_method`, `route`, and `response_status_code` labels. The default-dimension marker `{{default}}` is rendered as an empty Prometheus label value.

For Gosoline metrics, the Prometheus writer always uses the fixed `gosoline` prefix. It names the framework that authored the metric, not the application. Do not configure an application-specific Prometheus prefix: `metric.writer_settings.prometheus.naming.namespace_pattern` and `metric.writer_settings.prometheus.naming.namespace_delimiter` no longer control metric names. Put application identity on the scrape target through service discovery/target labels; Gosoline does not copy app identity into each metric name.

CloudWatch keeps its separate, still-supported namespace setting:

```
metric:

  writer_settings:

    cloudwatch:

      naming:

        namespace_pattern: "{app.namespace}-{app.name}"

        namespace_delimiter: "/"
```

This setting controls the CloudWatch **namespace** in AWS, not the metric name. CloudWatch metric names are rendered from the authored namespace and leaf as one PascalCased name; it does not add `gosoline` to the metric name. The datum's metric namespace and the CloudWatch service namespace are separate concepts. Consequently, writing an application-authored metric with an empty Gosoline namespace does not remove the configured CloudWatch namespace: only its metric name is PascalCased from the leaf.

For a namespace-less custom name, CloudWatch converts canonical separators in the leaf too: `my.custom.metric` becomes `MyCustomMetric`. Prometheus still applies its fixed `gosoline` writer prefix, converts dots in the leaf to underscores, and appends the unit/counter suffix (the example is a counter). OTEL normalizes and exports that namespace-less leaf without a `gosoline.` prefix. These behaviors are intentional; re-key CloudWatch dashboards or alarms using namespace-less custom metrics if they previously depended on the unrendered dotted leaf.

## Units and descriptions[​](#units-and-descriptions "Direct link to Units and descriptions")

The inventory's unit is the unit the emitter authors. Gosoline durations are generally recorded in milliseconds, and Prometheus and OTEL scale those values to seconds and render the base-unit suffix or instrument unit. `cloud.aws.kinesis.acquire.duration` is the one duration authored in seconds. Byte histograms are authored in bytes. Count metrics use count units.

Prometheus appends `_seconds` to duration histograms and `_bytes` to byte histograms; counters append `_total` (a count counter therefore ends in `_total`, without an extra `_count` unit suffix). OTEL puts the unit on the instrument rather than in its metric name. It scales milliseconds to seconds and uses `By` for byte values. For count instruments whose final leaf is plural, it derives a UCUM item annotation from that leaf's singular form: `stream.consumed.messages` uses `{message}`, `kafka.consumer.polls` uses `{poll}`, and `kafka.connects` uses `{connect}`. Nonplural leaves such as `.count` use unit `1`. CloudWatch receives the authored CloudWatch unit. The source inventory below distinguishes a counter from a gauge or histogram: their equal `count` units do not make their measurements interchangeable.

Every framework metric has a registered description. Prometheus and OTEL use that help text; a per-datum `Kind.WithHelp` value overrides the registered description. If no help has been registered for an application-authored metric, the writer falls back to `unit: <unit>`.

## Current framework inventory[​](#current-framework-inventory "Direct link to Current framework inventory")

The V2 inventory contains **62 metrics across 18 namespaces**.

The dimensions listed are the actual attributes emitted with that metric. Entries in `error.type` or `outcome` describe separate series, not separate metric names. The `kafka.consumer` and `kafka.producer` operation metrics are topic-level (no partition attribute). Franz-go wire-batch metrics carry `topic.name` and `partition.id`; connection/throttle metrics instead carry `broker.address`. SQL pool state and SQL open attempts are separate metrics with different semantics.

### `stream`[​](#stream "Direct link to stream")

| Canonical metric            | Type      | Authored unit | Dimensions                    | Description / actual measurement                                                                                                                                          |
| --------------------------- | --------- | ------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `stream.process.duration`   | histogram | ms            | `consumer.name`               | Duration of one consumer processing operation; a batch consumer records one observation for the whole batch.                                                              |
| `stream.consumed.messages`  | counter   | count         | `consumer.name`, `error.type` | Messages a consumer took in from its input, by error type.                                                                                                                |
| `stream.retry.operations`   | counter   | count         | `consumer.name`, `operation`  | Retry-queue `get` and `put` attempts; not a count of retried records or proof that a queue operation succeeded.                                                           |
| `stream.produced.messages`  | counter   | count         | `producer.name`               | Input messages in calls to a running daemon's `Write`, counted before aggregation and batching; this does not establish that those stages or remote delivery succeeded.   |
| `stream.batch.messages`     | histogram | count         | `producer.name`               | Messages passed to the configured output per batch after aggregation and batching; the observation is recorded after the output call returns even if it reports an error. |
| `stream.aggregate.messages` | histogram | count         | `producer.name`               | Messages combined by a producer daemon into one aggregate.                                                                                                                |
| `stream.idle.duration`      | histogram | ms            | `producer.name`               | Time an output worker waits for the next batch on the daemon's output channel, capped at the configured `interval`; it is not output or broker-request latency.           |
| `stream.message.count`      | gauge     | count         | `list.name`                   | Messages currently held by a Redis list input.                                                                                                                            |
| `stream.reads`              | counter   | count         | `list.name`                   | Read operations performed by a Redis list input.                                                                                                                          |
| `stream.writes`             | counter   | count         | `list.name`                   | Messages supplied to a Redis-list output batch, counted after the list's `RPush` calls; not a Redis command count or confirmation that each item was stored.              |

For `stream.consumed.messages`, `error.type="{{default}}"` is the all-processed total, including callback failures; `error.type="processing_failed"` is an overlapping per-record subset. Do not add or subtract those series to infer successful messages.

### `kafka.consumer`[​](#kafkaconsumer "Direct link to kafkaconsumer")

| Canonical metric                   | Type      | Authored unit | Dimensions                                               | Description / actual measurement                                                                                                                                                                                         |
| ---------------------------------- | --------- | ------------- | -------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `kafka.consumer.consumed.messages` | counter   | count         | `client.type`, `client.name`, `topic.name`, `error.type` | Records a Kafka consumer took in from its topic, by error type.                                                                                                                                                          |
| `kafka.consumer.process.duration`  | histogram | ms            | `client.type`, `client.name`, `topic.name`               | Duration of one Kafka consumer processing cycle, including configured consume-delay waiting, record-callback processing, and offset-commit work.                                                                         |
| `kafka.consumer.polls`             | counter   | count         | `client.type`, `client.name`, `topic.name`               | Poll calls counted when the following processing/commit path completes successfully; errors or cancellation that abort that path suppress the sample (successful shutdown drains may still count).                       |
| `kafka.consumer.poll.duration`     | histogram | ms            | `client.type`, `client.name`, `topic.name`               | Duration of a `poll()` request, recorded when the following processing/commit path completes successfully; errors or cancellation that abort that path suppress the sample (successful shutdown drains may still count). |
| `kafka.consumer.commit.duration`   | histogram | ms            | `client.type`, `client.name`, `topic.name`               | Duration of a Kafka consumer offset commit.                                                                                                                                                                              |
| `kafka.consumer.sleep.duration`    | histogram | ms            | `client.type`, `client.name`, `topic.name`               | Completed per-record configured consume-delay waits before processing; canceled waits produce no observation.                                                                                                            |
| `kafka.consumer.rebalances`        | counter   | count         | `client.type`, `client.name`, `topic.name`               | Consumer-group rebalances a Kafka consumer took part in.                                                                                                                                                                 |

These consumer metrics are topic-level and do not include a partition dimension.

### `kafka.producer`[​](#kafkaproducer "Direct link to kafkaproducer")

| Canonical metric                  | Type      | Authored unit | Dimensions                                               | Description / actual measurement                                                                      |
| --------------------------------- | --------- | ------------- | -------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `kafka.producer.sent.messages`    | counter   | count         | `client.type`, `client.name`, `topic.name`, `error.type` | Records a Kafka producer handed to the broker, by error type.                                         |
| `kafka.producer.produce.duration` | histogram | ms            | `client.type`, `client.name`, `topic.name`               | Duration of a Kafka producer send.                                                                    |
| `kafka.producer.batch.records`    | histogram | count         | `client.type`, `client.name`, `topic.name`               | Records passed to one synchronous producer call; this is not a wire-batch count or proof of delivery. |

Producer metrics are topic-level and do not include a partition dimension.

### `kafka` (franz-go hooks)[​](#kafka-franz-go-hooks "Direct link to kafka-franz-go-hooks")

The broker connection/throttle metrics carry `client.type`, `client.name`, and `broker.address`. Only `kafka.connects` also carries `error.type` (`connect_failed` or `{{default}}`); `kafka.throttles` and `kafka.throttle.duration` do not. Produce/fetch wire-batch metrics carry `client.type`, `client.name`, `topic.name`, and `partition.id` instead; they do not carry `broker.address`.

| Canonical metric                      | Type      | Authored unit | Dimensions                                                   | Description / actual measurement                                             |
| ------------------------------------- | --------- | ------------- | ------------------------------------------------------------ | ---------------------------------------------------------------------------- |
| `kafka.connects`                      | counter   | count         | `client.type`, `client.name`, `broker.address`, `error.type` | Broker connection attempts a Kafka client observed, by `error.type` outcome. |
| `kafka.throttles`                     | counter   | count         | `client.type`, `client.name`, `broker.address`               | Requests a broker throttled.                                                 |
| `kafka.throttle.duration`             | histogram | ms            | `client.type`, `client.name`, `broker.address`               | Time a broker throttled a Kafka client for.                                  |
| `kafka.produce.batch.records`         | histogram | count         | `client.type`, `client.name`, `topic.name`, `partition.id`   | Records a Kafka client produced per batch.                                   |
| `kafka.produce.batch.size`            | histogram | bytes         | `client.type`, `client.name`, `topic.name`, `partition.id`   | Uncompressed size of a batch a Kafka client produced.                        |
| `kafka.produce.batch.compressed.size` | histogram | bytes         | `client.type`, `client.name`, `topic.name`, `partition.id`   | Compressed size of a batch a Kafka client produced.                          |
| `kafka.fetch.batch.records`           | histogram | count         | `client.type`, `client.name`, `topic.name`, `partition.id`   | Records a Kafka client fetched per batch.                                    |
| `kafka.fetch.batch.size`              | histogram | bytes         | `client.type`, `client.name`, `topic.name`, `partition.id`   | Uncompressed size of a batch a Kafka client fetched.                         |
| `kafka.fetch.batch.compressed.size`   | histogram | bytes         | `client.type`, `client.name`, `topic.name`, `partition.id`   | Compressed size of a batch a Kafka client fetched.                           |

### `cloud.aws.kinesis`[​](#cloudawskinesis "Direct link to cloudawskinesis")

| Canonical metric                      | Type      | Authored unit | Dimensions                                  | Description / actual measurement                                                                                                                                                                    |
| ------------------------------------- | --------- | ------------- | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cloud.aws.kinesis.consumed.messages` | counter   | count         | `stream.name`, `partition.id`, `error.type` | Records processed by a Kinesis shard reader; the default series is the completed-batch total and `processing_failed` is an overlapping per-record subset.                                           |
| `cloud.aws.kinesis.sent.messages`     | counter   | count         | `stream.name`, `error.type`                 | Records sent to a Kinesis stream per `PutRecords` attempt, by Kinesis error reason.                                                                                                                 |
| `cloud.aws.kinesis.process.duration`  | histogram | ms            | `stream.name`, `partition.id`               | Duration of processing one batch of Kinesis records.                                                                                                                                                |
| `cloud.aws.kinesis.reads`             | counter   | count         | `stream.name`, `partition.id`               | `GetRecords` calls a Kinesis shard reader performed.                                                                                                                                                |
| `cloud.aws.kinesis.lag`               | histogram | ms            | `stream.name`, `partition.id`               | AWS-reported milliseconds behind the latest record for this shard reader, from Kinesis `GetRecordsOutput.MillisBehindLatest`; it is not wall-clock age of the last processed record.                |
| `cloud.aws.kinesis.acquire.duration`  | histogram | s             | `stream.name`, `partition.id`               | Time a Kinesis shard reader waited to acquire a shard.                                                                                                                                              |
| `cloud.aws.kinesis.sleep.duration`    | histogram | ms            | `stream.name`, `partition.id`               | Completed per-record configured `ConsumeDelay` waits before processing; canceled waits produce no observation.                                                                                      |
| `cloud.aws.kinesis.wait.duration`     | histogram | ms            | `stream.name`, `partition.id`               | Computed delay scheduled before the next poll: normally the remaining `WaitTime` after the time since the previous `GetRecords` call, and zero when reported lag exceeds `WaitTime + ConsumeDelay`. |
| `cloud.aws.kinesis.shard.count`       | gauge     | count         | `stream.name`                               | Shards a Kinesis stream currently has.                                                                                                                                                              |
| `cloud.aws.kinesis.client.count`      | gauge     | count         | `stream.name`                               | Clients currently consuming a Kinesis stream.                                                                                                                                                       |
| `cloud.aws.kinesis.batch.records`     | histogram | count         | `stream.name`                               | Records included in one `PutRecords` request attempt (including retries).                                                                                                                           |

For Kinesis `consumed.messages`, the `{{default}}` batch-total series includes failed records; `error.type="processing_failed"` is an overlapping per-record subset. For `sent.messages`, `error.type` carries the reason Kinesis reported. `lag` keeps millisecond samples and CloudWatch maximum aggregation; its OTEL histogram exposes the cumulative maximum per stream and shard.

### `kvstore`[​](#kvstore "Direct link to kvstore")

| Canonical metric     | Type    | Authored unit | Dimensions                      | Description / actual measurement                                                                                                                         |
| -------------------- | ------- | ------------- | ------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kvstore.reads`      | counter | count         | `model.id`, `store.type`, `hit` | Read operations against a key-value store, by whether the store served them itself.                                                                      |
| `kvstore.writes`     | counter | count         | `model.id`, `store.type`        | Write operations performed against a key-value store.                                                                                                    |
| `kvstore.deletes`    | counter | count         | `model.id`, `store.type`        | Delete operations performed against a key-value store.                                                                                                   |
| `kvstore.item.count` | gauge   | count         | `model.id`, `store.type`        | Estimated item count from a store's optional `SizedStore.EstimateSize()` implementation; the metric is not reported when that capability is unavailable. |

### `db.client`, `db.repo`, and `ddb`[​](#dbclient-dbrepo-and-ddb "Direct link to dbclient-dbrepo-and-ddb")

| Canonical metric                    | Type      | Authored unit | Dimensions                                 | Description / actual measurement                                                                                               |
| ----------------------------------- | --------- | ------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `db.client.connection.count`        | gauge     | count         | `connection.state` (`used`, `idle`)        | Current SQL pool connections by used/idle state. The framework emits the two states; it does not emit a separate `open` state. |
| `db.client.connections`             | counter   | count         | none                                       | SQL driver `Open` attempts since process start; one is recorded before each underlying driver's `Open` call.                   |
| `db.repo.operation.duration`        | histogram | ms            | `operation.name`, `model.id`, `error.type` | Duration of a SQL repository operation.                                                                                        |
| `db.repo.model_event.notifications` | counter   | count         | `model.id`, `error.type`                   | Model-event notifications a repository published, by error type.                                                               |
| `ddb.operation.duration`            | histogram | ms            | `operation.name`, `model.id`, `error.type` | Duration of a DynamoDB repository operation.                                                                                   |

The framework's SQL metrics are separate from optional SQLC instrumentation. The published SQLC integration remains a legacy exception: it emits `DbConnectionCount` with an empty Gosoline metric namespace, `Type="new"` once per driver `Open` attempt, and `Type="open"`, `"inUse"`, and `"idle"` connection-pool measurements each minute. That is not an emission of `db.client.connections` or `db.client.connection.count`, and is not part of the 62-name framework V2 inventory above.

### `mdlsub`[​](#mdlsub "Direct link to mdlsub")

| Canonical metric | Type    | Authored unit | Dimensions                                                 | Description / actual measurement                                                                                                                    |
| ---------------- | ------- | ------------- | ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mdlsub.events`  | counter | count         | `model.id`, `outcome` (`applied`, `skipped`), `error.type` | Subscriber events counted on applied/skipped outcomes or an apply failure (`error.type=apply_failed`); deferred/cancelled events produce no metric. |

### `http.server`[​](#httpserver "Direct link to httpserver")

| Canonical metric                   | Type      | Authored unit | Dimensions                                                | Description / actual measurement                        |
| ---------------------------------- | --------- | ------------- | --------------------------------------------------------- | ------------------------------------------------------- |
| `http.server.request.duration`     | histogram | ms            | `name`, `route`, `request.method`, `response.status_code` | Duration of an HTTP server request.                     |
| `http.server.rejected.requests`    | counter   | count         | `name`, `route`, `request.method`                         | HTTP requests the server rejected before handling them. |
| `http.server.active_request.count` | gauge     | count         | `name`                                                    | HTTP requests the server is currently handling.         |
| `http.server.connection.count`     | gauge     | count         | `name`                                                    | Connections the HTTP server currently holds open.       |

### `http.client`[​](#httpclient "Direct link to httpclient")

| Canonical metric               | Type      | Authored unit | Dimensions                                             | Description / actual measurement    |
| ------------------------------ | --------- | ------------- | ------------------------------------------------------ | ----------------------------------- |
| `http.client.request.duration` | histogram | ms            | `request.method`, `response.status_code`, `error.type` | Duration of an HTTP client request. |

A canceled client request is recorded with `error.type="request_canceled"` before its cancellation error is returned.

### `rpc.server`[​](#rpcserver "Direct link to rpcserver")

| Canonical metric              | Type      | Authored unit | Dimensions          | Description / actual measurement |
| ----------------------------- | --------- | ------------- | ------------------- | -------------------------------- |
| `rpc.server.request.duration` | histogram | ms            | `service`, `method` | Duration of a gRPC server call.  |

### Remaining namespaces[​](#remaining-namespaces "Direct link to Remaining namespaces")

| Canonical metric                     | Type      | Authored unit | Dimensions                                                                  | Description / actual measurement                                                                                                                      |
| ------------------------------------ | --------- | ------------- | --------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `blob.batch.operations`              | counter   | count         | `operation`                                                                 | S3 `GetObject`, `PutObject`, `CopyObject`, and `DeleteObject` calls counted after each call returns; this counts attempts and does not imply success. |
| `conc.scheduler.batch.tasks`         | histogram | count         | `name`                                                                      | Number of tasks in the batch supplied to one `batchRunner` call, recorded before the call; not a successful-task count.                               |
| `conc.scheduler.task.queue.duration` | histogram | ms            | `name`                                                                      | Elapsed time from `ScheduleJob` start before enqueue until its `batchRunner` result is received, including queueing and batch execution.              |
| `limit.takes`                        | counter   | count         | `name`, `prefix`, `outcome` (`allowed`, `throttled`, `error`), `error.type` | Attempts to take a rate-limiter token, by outcome and error type; recorded when the take ends.                                                        |
| `smpl.decisions`                     | counter   | count         | `sampled`                                                                   | Sampling decisions made, by outcome.                                                                                                                  |
| `metric.log.records`                 | counter   | count         | `level`                                                                     | Log records written at warn or error level, by level.                                                                                                 |

## Folded outcomes and removed names[​](#folded-outcomes-and-removed-names "Direct link to Folded outcomes and removed names")

The V2 contract records operation outcomes as dimensions instead of separate outcome metrics. Re-key queries that used these removed names:

| Removed metric name(s)                                                     | Current metric                        | Dimension(s)                                                                              |
| -------------------------------------------------------------------------- | ------------------------------------- | ----------------------------------------------------------------------------------------- |
| `stream.errors`                                                            | `stream.consumed.messages`            | `error.type`                                                                              |
| `kafka.producer.send.errors`                                               | `kafka.producer.sent.messages`        | `error.type`                                                                              |
| `cloud.aws.kinesis.send.errors`                                            | `cloud.aws.kinesis.sent.messages`     | `error.type` (the Kinesis-reported reason)                                                |
| `cloud.aws.kinesis.consume.errors`                                         | `cloud.aws.kinesis.consumed.messages` | `error.type`                                                                              |
| `kvstore.hits`                                                             | `kvstore.reads`                       | `hit` (`true`, `false`)                                                                   |
| `limit.releases`, `limit.throttles`, `limit.errors`                        | `limit.takes`                         | `outcome` (`allowed`, `throttled`, `error`) and `error.type`                              |
| `mdlsub.consumed.events`, `mdlsub.skipped.events`, `mdlsub.consume.errors` | `mdlsub.events`                       | `outcome` (`applied`, `skipped`) and `error.type`                                         |
| `kafka.consumer.consume.errors`                                            | `kafka.consumer.consumed.messages`    | `error.type`                                                                              |
| `db.repo.model_event.notify.errors`                                        | `db.repo.model_event.notifications`   | `error.type`; the former error-only metric is not emitted and there is no `success` label |

For exclusive outcomes, attempts are the sum over the dimension values. `stream.consumed.messages` and `cloud.aws.kinesis.consumed.messages` are exceptions: their `{{default}}` series is the all- processed total and `processing_failed` is an overlapping subset. Never sum or subtract those values to infer success. `error.type` is a bounded action/reason such as `send_failed`, `connect_failed`, `request_failed`, `request_canceled`, `operation_failed`, `notify_failed`, `apply_failed`, or `processing_failed`, not a Go error type or message.

### Other V1 name changes[​](#other-v1-name-changes "Direct link to Other V1 name changes")

| V1 name                        | V2 name                                                                                               | Description                                           |
| ------------------------------ | ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| `metric.records`               | `metric.log.records`                                                                                  | The leaf now identifies log-volume data.              |
| `db.client.operation.duration` | `db.repo.operation.duration` for SQL repositories; `ddb.operation.duration` for DynamoDB repositories | Repository telemetry has distinct V2 names by source. |

V2 no longer emits `autoscaling.per_runner.stream.messages`, `autoscaling.per_runner.http.server.requests`, `stream.available.messages`, or `stream.sent.messages`. Those historical calculator/per-runner metrics are not current autoscaling signals; use application- specific signals instead. The active Kinsumer autoscale leader-election table naming pattern remains a separate resource-naming feature; see the [Naming Patterns reference](/docs/pr-5/reference/naming-patterns/.md#kinsumer-autoscale-leader-election-table).
