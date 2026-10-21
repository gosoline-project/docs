# kvstore package

`pkg/kvstore` provides generic model-keyed stores with in-memory, Redis, and DynamoDB backends. Its interface supports individual and batch reads, writes, and deletes.

## API and activation[​](#api-and-activation "Direct link to API and activation")

When `kvstore.Settings.MetricsEnabled` is true, `kvstore.NewMetricStoreWithInterfaces` wraps the configured backend. `kvstore.reads`, `kvstore.writes`, and `kvstore.deletes` describe successful wrapper calls; `kvstore.attempted.reads` counts requested read items before delegation, including delegated failures. For Redis, batch counters use caller item counts while the backend may split batches into Redis command chunks according to `BatchSize`—they are not command or chunk counts. The separate `kvstore.item.count` series samples an optional backend size estimate, not an operation batch size.

## Metrics[​](#metrics "Direct link to Metrics")

The table lists names and units authored by Gosoline. The OpenTelemetry writer exports names with a `gosoline.` prefix and converts authored `ms` durations to the OpenTelemetry duration unit. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for the writer contract and dimension conventions.

| Metric                    | Instrument | Authored unit | Dimensions                                                                                        | Description                                                                                                                                                                                                                                                                                         |
| ------------------------- | ---------- | ------------- | ------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kvstore.reads`           | Counter    | `count`       | `model.id`: configured model ID; `store.type`: concrete backend Go type; `hit`: `true` or `false` | For each successful `Contains`, `Get`, or `GetBatch`, records item counts split into items found by this store (`true`) and not found (`false`). Both series are written, including a zero-valued side; errors and input-conversion failures emit no read points.                                   |
| `kvstore.attempted.reads` | Counter    | `count`       | `model.id`: configured model ID; `store.type`: concrete backend Go type                           | Counts requested read items before delegation: `Contains` and `Get` add one before calling the store; `GetBatch` adds the number of keys after successful conversion and before delegation. Delegated errors count; key-conversion failures do not. High-priority zero default per model and store. |
| `kvstore.writes`          | Counter    | `count`       | `model.id`: configured model ID; `store.type`: concrete backend Go type                           | On success, `Put` adds one and `PutBatch` adds the number of input entries. The batch value is the caller's item count, not the number of Redis commands/chunks or newly created keys.                                                                                                              |
| `kvstore.deletes`         | Counter    | `count`       | `model.id`: configured model ID; `store.type`: concrete backend Go type                           | On success, `Delete` adds one and `DeleteBatch` adds the number of input keys. The batch value is the caller's item count, not the number of Redis commands/chunks; errors emit no delete points.                                                                                                   |
| `kvstore.item.count`      | Gauge      | `count`       | `model.id`: configured model ID; `store.type`: concrete backend Go type                           | Samples the optional `SizedStore.EstimateSize()` result once per minute after the sampler starts. A backend must implement `SizedStore`; a `nil` estimate emits no value and is not treated as zero. Redis uses `DBSIZE`; the in-memory backend reports its tracked estimate.                       |

The `store.type` dimension is the concrete Go type string, not a normalized backend label. A DDB backend may also emit metrics owned by `pkg/ddb`; see the [DDB package reference](/docs/pr-5/reference/package-ddb/.md) for that separate contract.

## Related references[​](#related-references "Direct link to Related references")

* [DynamoDB package](/docs/pr-5/reference/package-ddb/.md)
* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
* [Metric wrapper and size sampler](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/kvstore/metric.go)
* [`SizedStore` interface](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/kvstore/kvstore.go)
