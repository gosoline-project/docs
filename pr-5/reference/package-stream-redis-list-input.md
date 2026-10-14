# Redis list input

The `stream` package's Redis list input implements the stream `Input` interface. It blocks on a configured Redis list with `BLPOP`, decodes each payload as a `stream.Message`, and passes decoded messages into the consumer pipeline.

## Activation[​](#activation "Direct link to Activation")

Configure an input with `type: redis` and a Redis server name and list key:

```
stream:

  input:

    orders:

      type: redis

      server_name: default

      key: orders

      wait_time: 5s
```

The exported constructor is `stream.NewRedisListInput(ctx, config, logger, settings)`; it returns a `stream.Input`. The `RedisListInputSettings` includes the Redis server name, key, wait time, health-check timeout, and runner count. A zero wait time is rejected when the input runs. For the consumer wiring and transport behavior, see [Create a consumer](/docs/pr-5/how-to/streaming-applications/create-a-consumer/.md).

## Metrics[​](#metrics "Direct link to Metrics")

The table lists canonical Gosoline names as authored, not backend-exported names. The metric writer places these in the `stream` namespace. OpenTelemetry prefixes the namespace with `gosoline.` and converts durations authored in milliseconds to seconds; that conversion does not change the authored units shown here. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for exporter mappings.

| Metric                 | Instrument | Authored unit | Dimensions                          | Description                                                                                                                                                                             |
| ---------------------- | ---------- | ------------- | ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `stream.message.count` | Gauge      | count         | `list.name` = `<server_name>-<key>` | Current Redis-list length, sampled immediately when the metric loop starts and then once per second. If Redis `LLEN` fails, the input logs the error and emits no sample for that read. |
| `stream.reads`         | Counter    | count         | `list.name` = `<server_name>-<key>` | Increments after a message has been removed by `BLPOP`, decoded, and handed to the input's processing channel. It measures intake, not callback success.                                |

## Operational notes[​](#operational-notes "Direct link to Operational notes")

Redis removes a list item as part of `BLPOP`; it is not redelivered by the input if decoding or later processing fails. A malformed payload is logged and skipped before `stream.reads` is incremented. Use the stream consumer retry handler when application failures need a separate retry path; it does not restore the original Redis list entry.

The `list.name` dimension is the configured server name and key joined with a hyphen. These metrics do not include a success/failure dimension for callback processing.

## Related[​](#related "Direct link to Related")

* [Understand inputs and outputs](/docs/pr-5/how-to/streaming-applications/understand-inputs-and-outputs/.md)
* [Create a consumer](/docs/pr-5/how-to/streaming-applications/create-a-consumer/.md)
* [Redis list output metrics](/docs/pr-5/reference/package-stream-redis-list-output/.md)
* [Metric schema reference](/docs/pr-5/reference/metric-schema/.md)
