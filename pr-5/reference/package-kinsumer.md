# Kinsumer

`Kinsumer` is the current exported Kinesis consumer API in `github.com/justtrackio/gosoline/pkg/cloud/aws/kinesis`. It coordinates shard ownership and checkpointing across clients, runs per-shard readers, and delivers record payloads to a `RecordHandler`. Construct it with `NewKinsumer(ctx, config, logger, settings)` and a `*Settings` (`Settings.StreamName` is required), then call `Run(ctx, handler)` to consume. The `Kinsumer` interface also exposes `Stop` and `IsHealthy` for lifecycle management.

These gauges are emitted by the current exported `Kinsumer` API through `NewKinsumer` and `Run`. Shard readers emit their own metrics on the [Kinesis shard reader page](/docs/pr-5/reference/package-cloud_aws_kinesis-reader/.md).

## Metrics[​](#metrics "Direct link to Metrics")

Names below are Gosoline-authored canonical names. The OTEL writer exports them with a `gosoline.` prefix; the table shows the units authored by the emitter. See the [metric schema](/docs/pr-5/reference/metric-schema/.md) for writer conversion details.

| Metric                           | Instrument | Authored unit | Dimensions                       | Description                                                                                                                                                |
| -------------------------------- | ---------- | ------------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cloud.aws.kinesis.shard.count`  | Gauge      | count         | `stream.name` (full stream name) | Number of shards in the current shard-distribution snapshot for this stream. Emitted when the consumer group starts and every minute while it runs.        |
| `cloud.aws.kinesis.client.count` | Gauge      | count         | `stream.name` (full stream name) | Number of Kinsumer clients known for this stream in the same distribution snapshot. Emitted when the consumer group starts and every minute while it runs. |

Both values are gauges, not cumulative counters. The only dimension is the full stream name; neither metric adds a shard ID, client ID, or error dimension. The two values can be compared per stream to understand shard-to-client distribution, but they do not identify individual shard assignments or clients.

## Related pages[​](#related-pages "Direct link to Related pages")

* [Kinesis shard reader](/docs/pr-5/reference/package-cloud_aws_kinesis-reader/.md)
* [Kinesis RecordWriter](/docs/pr-5/reference/package-cloud_aws_kinesis-writer/.md)
* [AWS package configuration](/docs/pr-5/reference/package-cloud_aws/.md)
* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
* [Creating a consumer](/docs/pr-5/how-to/streaming-applications/create-a-consumer/.md)
