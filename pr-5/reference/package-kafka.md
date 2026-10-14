# Kafka client hook metrics

The `pkg/kafka` package's `MetricsHook` records broker connection and throttle events plus franz-go produce/fetch wire batches. These are transport-level hook measurements; the Kafka consumer and producer packages emit separate client-operation metrics.

## Activation[​](#activation "Direct link to Activation")

Gosoline's Kafka consumer reader and producer writer install this hook when they create their franz-go clients. No separate stream setting is required. Configure Gosoline metric collection and a writer/exporter to make the emitted measurements visible.

The `client.type` dimension is `consumer` or `producer`; `client.name` is the configured client name. Broker callbacks add `broker.address` as `host:port`. Wire-batch callbacks instead add the resolved `topic.name` and decimal-string `partition.id`; they do not include `broker.address`.

## Metrics[​](#metrics "Direct link to Metrics")

These are canonical Gosoline names as authored. The hook writes them in the `kafka` namespace. OpenTelemetry prefixes the namespace with `gosoline.` and converts durations authored in milliseconds to seconds; that does not change the authored units shown here. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for exporter mappings.

| Metric                                | Instrument | Authored unit | Dimensions                                                                                                                                                                                               | Description                                                                                   |
| ------------------------------------- | ---------- | ------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| `kafka.connects`                      | Counter    | count         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `broker.address` (host<!-- -->:port<!-- -->); `error.type` = `{{default}}` for success or `connect_failed` for failure | Broker connection attempts. The error dimension distinguishes successful and failed attempts. |
| `kafka.throttles`                     | Counter    | count         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `broker.address` (host<!-- -->:port<!-- -->)                                                                           | Broker-throttled requests.                                                                    |
| `kafka.throttle.duration`             | Histogram  | milliseconds  | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `broker.address` (host<!-- -->:port<!-- -->)                                                                           | Time a broker throttled a Kafka client.                                                       |
| `kafka.produce.batch.records`         | Histogram  | count         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `topic.name` (resolved topic); `partition.id` (decimal-string partition)                                               | Records in one franz-go produce wire batch.                                                   |
| `kafka.produce.batch.size`            | Histogram  | bytes         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `topic.name` (resolved topic); `partition.id` (decimal-string partition)                                               | Uncompressed bytes in one produce wire batch.                                                 |
| `kafka.produce.batch.compressed.size` | Histogram  | bytes         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `topic.name` (resolved topic); `partition.id` (decimal-string partition)                                               | Compressed bytes in one produce wire batch.                                                   |
| `kafka.fetch.batch.records`           | Histogram  | count         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `topic.name` (resolved topic); `partition.id` (decimal-string partition)                                               | Records in one franz-go fetch wire batch.                                                     |
| `kafka.fetch.batch.size`              | Histogram  | bytes         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `topic.name` (resolved topic); `partition.id` (decimal-string partition)                                               | Uncompressed bytes in one fetch wire batch.                                                   |
| `kafka.fetch.batch.compressed.size`   | Histogram  | bytes         | `client.type` = `consumer` or `producer`; `client.name` (configured client name); `topic.name` (resolved topic); `partition.id` (decimal-string partition)                                               | Compressed bytes in one fetch wire batch.                                                     |

Only `kafka.connects` carries `error.type`; successful connections use `{{default}}`, while failed connections use `connect_failed`. The throttle metrics have only the broker dimensions shown above. None of the hook metrics reports business success or durable delivery; correlate them with the [Kafka consumer metrics](/docs/pr-5/reference/package-kafka-consumer/.md), [Kafka producer metrics](/docs/pr-5/reference/package-kafka-producer/.md), and application outcomes.

## Related[​](#related "Direct link to Related")

* [General Kafka usage](/docs/pr-5/how-to/kafka/general/.md)
* [Kafka consumer metrics](/docs/pr-5/reference/package-kafka-consumer/.md)
* [Kafka producer metrics](/docs/pr-5/reference/package-kafka-producer/.md)
* [Metric schema reference](/docs/pr-5/reference/metric-schema/.md)
