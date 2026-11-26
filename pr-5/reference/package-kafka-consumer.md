# Kafka consumer metrics

The `pkg/kafka/consumer` component polls Kafka records, coordinates callback processing, and commits offsets. A stream input configured with `type: kafka` uses this consumer path; the public `consumer.NewConsumer` constructor also creates a consumer directly. These measurements describe the Kafka client processing cycle, not an application's business outcomes.

## Activation[​](#activation "Direct link to Activation")

Configure a Kafka input under `stream.input.<name>`; Gosoline creates the consumer and its `kafka.consumer` metric writer. Enable Gosoline metric collection and configure a writer/exporter to observe the measurements. See [Kafka consumer configuration](/docs/pr-5/how-to/kafka/general/.md#consumer-configuration) for input settings and shutdown behavior.

## Metrics[​](#metrics "Direct link to Metrics")

These are canonical Gosoline names as authored in the `kafka.consumer` namespace. OpenTelemetry prefixes the namespace with `gosoline.` and converts durations authored in milliseconds to seconds; that does not change the authored units shown here. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for exporter mappings.

| Metric                             | Instrument | Authored unit | Dimensions                                                                                                                                                          | Description                                                                                                                                                                                                         |
| ---------------------------------- | ---------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kafka.consumer.consumed.messages` | Counter    | count         | `client.type` = `consumer`; `client.name` (configured consumer/input name); `topic.name` (resolved full topic); `error.type` = `{{default}}` or `processing_failed` | Records in a poll's work set. Each record is counted once: `processing_failed` if its callback failed, or `{{default}}` otherwise. During shutdown, `{{default}}` can include records whose callbacks did not run.  |
| `kafka.consumer.polls`             | Counter    | count         | `client.type` = `consumer`; `client.name` (configured consumer/input name); `topic.name` (resolved full topic)                                                      | One sample for a poll whose processing and offset-commit path returns nil. A poll that drains during graceful shutdown can still emit a sample; an error or cancellation that aborts this path suppresses it.       |
| `kafka.consumer.poll.duration`     | Histogram  | milliseconds  | `client.type` = `consumer`; `client.name` (configured consumer/input name); `topic.name` (resolved full topic)                                                      | Duration of the `poll()` request, recorded when the following processing and offset-commit path returns nil. A graceful drain can still emit a sample; an error or cancellation that aborts the path suppresses it. |
| `kafka.consumer.process.duration`  | Histogram  | milliseconds  | `client.type` = `consumer`; `client.name` (configured consumer/input name); `topic.name` (resolved full topic)                                                      | One observation for a poll's processing cycle, including configured `consume_delay` waits, callback processing, and offset-commit work.                                                                             |
| `kafka.consumer.commit.duration`   | Histogram  | milliseconds  | `client.type` = `consumer`; `client.name` (configured consumer/input name); `topic.name` (resolved full topic)                                                      | Duration of a successfully completed offset commit. Failed commits return before the metric is written, so they have no duration sample.                                                                            |
| `kafka.consumer.sleep.duration`    | Histogram  | milliseconds  | `client.type` = `consumer`; `client.name` (configured consumer/input name); `topic.name` (resolved full topic)                                                      | Completed per-record `consume_delay` waits before processing; canceled waits produce no observation.                                                                                                                |
| `kafka.consumer.rebalances`        | Counter    | count         | `client.type` = `consumer`; `client.name` (configured consumer/input name); `topic.name` (resolved topic)                                                           | One increment for each topic entry in the consumer's partition-loss or revocation callback; it indicates partition ownership changes, not a count of every partition.                                               |

## Interpretation notes[​](#interpretation-notes "Direct link to Interpretation notes")

Poll count and duration include empty polls and polls drained during graceful shutdown if the processing path returns without error. The per-record delay histogram includes only waits that finish without cancellation.

## Related[​](#related "Direct link to Related")

* [General Kafka usage](/docs/pr-5/how-to/kafka/general/.md)
* [Kafka client-hook metrics](/docs/pr-5/reference/package-kafka/.md)
* [Kafka producer metrics](/docs/pr-5/reference/package-kafka-producer/.md)
* [Create a consumer](/docs/pr-5/how-to/streaming-applications/create-a-consumer/.md)
* [Metric schema reference](/docs/pr-5/reference/metric-schema/.md)
