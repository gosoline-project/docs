# blob package

`pkg/blob` provides batched object-storage access through named stores and a worker-backed `BatchRunner`. The runner uses the configured S3 client and processes each object in a batch through separate read, write, copy, and delete workers.

## API and activation[​](#api-and-activation "Direct link to API and activation")

Register `blob.ProvideBatchRunner(name)` as a kernel module and provide the matching `blob.ProvideStore` with the same name. The store queues object operations for that named runner; the runner loads its settings from `blob.<name>`. See the pinned [runner registration](https://github.com/justtrackio/gosoline/blob/dd90ce372ff3/pkg/blob/runner.go#L45-L50) and [store provider](https://github.com/justtrackio/gosoline/blob/dd90ce372ff3/pkg/blob/store.go#L118-L145).

## Metrics[​](#metrics "Direct link to Metrics")

The table lists names and units authored by Gosoline. The OpenTelemetry writer exports names with a `gosoline.` prefix and converts authored `ms` durations to the OpenTelemetry duration unit. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for the writer contract and dimension conventions.

| Metric                  | Instrument | Authored unit | Dimensions                                     | Description                                                                                                                                                                                                                            |
| ----------------------- | ---------- | ------------- | ---------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `blob.batch.operations` | Counter    | `count`       | `operation`: `Read`, `Write`, `Copy`, `Delete` | Adds one after each object-level S3 runner operation returns, including calls that return an error. This is not one count per batch API call or per network/retry attempt; queued work cancelled before a worker runs is not recorded. |

## Related references[​](#related-references "Direct link to Related references")

* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
* [Batch operation implementation](https://github.com/justtrackio/gosoline/blob/dd90ce372ff3/pkg/blob/runner.go#L139-L253)
