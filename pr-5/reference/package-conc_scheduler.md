# conc/scheduler package

The `pkg/conc/scheduler` package groups keyed jobs into batches and sends each batch to a caller-provided runner. Each `ScheduleJob` call receives the result associated with its key.

## API and activation[​](#api-and-activation "Direct link to API and activation")

Create a named scheduler with `NewScheduler[T](config, batchRunner, name)`. Its configuration is read from `scheduler.<name>`, where `name` is the constructor argument. The `BatchRunner[T]` receives the batch's keys and provider functions and returns a result map keyed by those keys, or an error.

Start `Run(ctx)` to consume batches. Calling `ScheduleJob(ctx, key, provider)` queues the job and waits for its result.

| Setting          | Default | Effect                                                          |
| ---------------- | ------- | --------------------------------------------------------------- |
| `batch_timeout`  | `10ms`  | Flushes a partial batch after this interval from its first job. |
| `runner_count`   | `25`    | Number of concurrent batch runners.                             |
| `max_batch_size` | `25`    | Flushes a batch when it reaches this many jobs.                 |

Batches also flush when the `Run` context is cancelled, so pending jobs can be passed to the runner before shutdown.

## Metrics[​](#metrics "Direct link to Metrics")

The table lists canonical Gosoline-authored names and units. The metric writer renders them per backend: OpenTelemetry prefixes namespaced metrics with `gosoline.` and OpenTelemetry/Prometheus scale authored millisecond durations to seconds. See the [metric schema](/docs/pr-5/reference/metric-schema/.md) for backend rendering and the complete contract.

| Metric                               | Instrument | Authored unit | Dimensions             | Description                                                                                                                                                                                                           |
| ------------------------------------ | ---------- | ------------- | ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `conc.scheduler.batch.tasks`         | Histogram  | `count`       | `name`: scheduler name | One observation per batch, written before the batch runner is called; the value is the number of jobs in the batch, including batches whose runner returns an error.                                                  |
| `conc.scheduler.task.queue.duration` | Histogram  | `ms`          | `name`: scheduler name | One observation per job when its result arrives. It spans from entry to `ScheduleJob` through queueing, batch dispatch, and the entire batch-runner call; it is recorded before the success/error result is returned. |

The standard `NewScheduler` writer is initialized with zero-valued defaults for both metrics, tagged with the scheduler `name`. The direct `NewSchedulerWithSettings` constructor accepts a writer supplied by the caller.

## Related[​](#related "Direct link to Related")

* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
