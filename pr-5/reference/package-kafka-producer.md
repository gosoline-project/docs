# Kafka producer metrics

The `pkg/kafka/producer` component sends records synchronously and reports per-call results. A stream output configured with `type: kafka` uses this producer path; the public `producer.NewProducer` constructor also creates one directly. These metrics describe producer-call outcomes and timing, distinct from Kafka broker and wire-batch hook measurements.

## Activation[​](#activation "Direct link to Activation")

Configure a Kafka output under `stream.output.<name>`; Gosoline creates the producer and its `kafka.producer` metric writer. Enable Gosoline metric collection and configure a writer/exporter to observe the measurements. See [Kafka producer configuration](/docs/pr-5/how-to/kafka/general/.md#producer-configuration).

## Metrics[​](#metrics "Direct link to Metrics")

These are canonical Gosoline names as authored in the `kafka.producer` namespace. OpenTelemetry prefixes the namespace with `gosoline.` and converts durations authored in milliseconds to seconds; that does not change the authored units shown here. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for exporter mappings.

| Metric                            | Instrument | Authored unit | Dimensions                                                                                                                                                     | Description                                                                                                                                                                                                       |
| --------------------------------- | ---------- | ------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kafka.producer.sent.messages`    | Counter    | count         | `client.type` = `producer`; `client.name` (configured producer/output name); `topic.name` (resolved full topic); `error.type` = `{{default}}` or `send_failed` | Records returned by a synchronous producer call split into successful (`{{default}}`) and failed (`send_failed`) outcomes. The buckets are disjoint; for a partial failure their sum is the records in that call. |
| `kafka.producer.batch.records`    | Histogram  | count         | `client.type` = `producer`; `client.name` (configured producer/output name); `topic.name` (resolved full topic)                                                | Number of records passed to one synchronous producer call; this is not a franz-go wire-batch count.                                                                                                               |
| `kafka.producer.produce.duration` | Histogram  | milliseconds  | `client.type` = `producer`; `client.name` (configured producer/output name); `topic.name` (resolved full topic)                                                | Duration of one synchronous producer call.                                                                                                                                                                        |

## Interpretation notes[​](#interpretation-notes "Direct link to Interpretation notes")

`kafka.producer.batch.records` measures records passed to one synchronous producer call, not broker wire batches; the separate [Kafka client-hook metrics](/docs/pr-5/reference/package-kafka/.md) describe wire batches and broker activity. With the [producer daemon](/docs/pr-5/how-to/streaming-applications/use-the-producer-daemon/.md), the stream daemon's asynchronous output metrics are a separate set and do not turn a daemon write into a synchronous delivery confirmation. Producer result metrics are not evidence of durable application effects.

## Related[​](#related "Direct link to Related")

* [General Kafka usage](/docs/pr-5/how-to/kafka/general/.md)
* [Kafka client-hook metrics](/docs/pr-5/reference/package-kafka/.md)
* [Kafka consumer metrics](/docs/pr-5/reference/package-kafka-consumer/.md)
* [Use the producer daemon](/docs/pr-5/how-to/streaming-applications/use-the-producer-daemon/.md)
* [Metric schema reference](/docs/pr-5/reference/metric-schema/.md)
