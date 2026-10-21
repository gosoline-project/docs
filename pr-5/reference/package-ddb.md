# ddb package

`pkg/ddb` provides a DynamoDB repository interface for model persistence. Gosoline's `mdlsub.NewOutputDdb` wraps its repository with `ddb.NewMetricRepository`; the wrapper instruments `PutItem`, while delete operations use `DeleteItem` and do not pass through this metric hook.

## API and activation[​](#api-and-activation "Direct link to API and activation")

`ddb.NewMetricRepository(config, logger, repo)` creates the metric-wrapped repository. In the standard DDB model output, both `Create` and `Update` persistence route through `PutItem` and are recorded with `operation.name=save`; deletes are not recorded by this wrapper. See [`NewMetricRepository`](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/ddb/metric_repo.go) and the [model output wiring](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/mdlsub/output_ddb.go).

## Metrics[​](#metrics "Direct link to Metrics")

The table lists names and units authored by Gosoline. The OpenTelemetry writer exports names with a `gosoline.` prefix and converts authored `ms` durations to the OpenTelemetry duration unit. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for the writer contract and `{{default}}` dimension behavior.

| Metric                   | Instrument | Authored unit | Dimensions                                                                                                                     | Description                                                                                                                                                                                                                                                                                                                                             |
| ------------------------ | ---------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ddb.operation.duration` | Histogram  | `ms`          | `operation.name`: `save`; `model.id`: repository model ID; `error.type`: `{{default}}` on success, `operation_failed` on error | Records after each wrapped `PutItem` result, including errors. **Source caveat:** the current implementation initializes the start time to `time.Time{}` and computes `time.Since(start)` after `PutItem`; the value is therefore measured from Go's zero time, not the call's elapsed duration. Do not interpret it as a normal per-operation latency. |
| `ddb.operations`         | Counter    | `count`       | `operation.name`: `save`; `model.id`: repository model ID; `error.type`: `{{default}}` on success, `operation_failed` on error | Adds one after each wrapped `PutItem`/save result, including failures. High-priority zero defaults exist for both outcomes per model and operation.                                                                                                                                                                                                     |

## Related references[​](#related-references "Direct link to Related references")

* [Key-value store package](/docs/pr-5/reference/package-kvstore/.md)
* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
* [DynamoDB metric implementation](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/ddb/metric_repo.go)
