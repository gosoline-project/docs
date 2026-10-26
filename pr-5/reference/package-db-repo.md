# db-repo package

`pkg/db-repo` provides a model-oriented SQL repository interface with create, read, update, delete, and query operations. It also contains the model-event publisher notifier used to send repository notifications.

## API and activation[​](#api-and-activation "Direct link to API and activation")

Repository-operation timing and counting are explicit opt-ins: wrap a repository with `db_repo.NewMetricRepository(config, logger, repo)`. The wrapper records `Create`, `Read`, `Update`, `Delete`, and `Query`; it does not override `Count`, and the package does not automatically wrap every repository. Model-event notification counts are emitted by the separate `db_repo.NewPublisherNotifier` path after each publisher result. See [`NewMetricRepository`](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/db-repo/metric_repo.go) and [`NewPublisherNotifier`](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/db-repo/notification_publisher.go).

## Metrics[​](#metrics "Direct link to Metrics")

The table lists names and units authored by Gosoline. The OpenTelemetry writer exports names with a `gosoline.` prefix and converts authored `ms` durations to the OpenTelemetry duration unit. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for the writer contract and `{{default}}` dimension behavior.

| Metric                              | Instrument | Authored unit | Dimensions                                                                                                                                                            | Description                                                                                                                                                                                        |
| ----------------------------------- | ---------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `db.repo.operation.duration`        | Histogram  | `ms`          | `operation.name`: `create`, `read`, `update`, `delete`, `query`; `model.id`: repository model ID; `error.type`: `{{default}}` on success, `operation_failed` on error | Records one duration observation after each wrapped repository operation. Keep this histogram for latency; use the Counter for operation volume. `Count` is not wrapped by this metric repository. |
| `db.repo.operations`                | Counter    | `count`       | `operation.name`: `create`, `read`, `update`, `delete`, `query`; `model.id`: repository model ID; `error.type`: `{{default}}` on success, `operation_failed` on error | Adds one after each wrapped call returns. High-priority zero defaults exist for both outcomes per model and operation.                                                                             |
| `db.repo.model_event.notifications` | Counter    | `count`       | `model.id`: configured model ID; `error.type`: `{{default}}` when publish succeeds, `notify_failed` when it returns an error, including cancellation                  | Adds one after each model-event `Publisher.Publish` result. Notification type is not a dimension; the counter records publish outcomes rather than the underlying database operation.              |

## Related references[​](#related-references "Direct link to Related references")

* [SQL database client package](/docs/pr-5/reference/package-db/.md)
* [Working with relational databases](/docs/pr-5/how-to/databases-sql/.md)
* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
* [Repository metric implementation](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/db-repo/metric_repo.go)
