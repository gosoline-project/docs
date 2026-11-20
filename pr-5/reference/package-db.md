# db package

`pkg/db` creates managed SQL connections as `*sqlx.DB` values from named database settings. It installs a driver wrapper for connection-attempt metrics and reads live pool state from `database/sql` statistics.

## API and activation[​](#api-and-activation "Direct link to API and activation")

Use `db.ProvideConnection(ctx, config, logger, name)` to provide a named connection; it resolves the corresponding settings and creates the managed connection. The open-attempt counter is part of this connection setup, while the pool-state sampler starts after connection setup, lifecycle registration, and migrations succeed. There is no separate database-metrics switch: export still depends on the application's configured Gosoline metric writer. See [`ProvideConnection`](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/db/connection.go) and the [metric wrapper](https://github.com/justtrackio/gosoline/blob/e1a14ac09eab0404cf620e1f3ee8a0fd5db259c0/pkg/db/metrics.go).

## Metrics[​](#metrics "Direct link to Metrics")

The table lists names and units authored by Gosoline. The OpenTelemetry writer exports names with a `gosoline.` prefix and converts authored `ms` durations to the OpenTelemetry duration unit. See the [metric schema reference](/docs/pr-5/reference/metric-schema/.md) for the writer contract and dimension conventions.

| Metric                       | Instrument | Authored unit | Dimensions                         | Description                                                                                                                                                                                      |
| ---------------------------- | ---------- | ------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `db.client.connection.count` | Gauge      | `count`       | `connection.state`: `used`, `idle` | Samples `sql.DB.Stats().InUse` and `.Idle` immediately, then once per minute. Both state series are written, including zero values; there is no separate `open` or total-state series.           |
| `db.client.connections`      | Counter    | `count`       | None                               | Adds one before each underlying driver `Open` call, including failed opens. It accumulates open attempts since process start; it is not a live connection count and is not decremented on close. |

## SQLC connections[​](#sqlc-connections "Direct link to SQLC connections")

In the standalone `sqlc` package, `NewDBFromSettings` installs the driver-open counter and starts the pool sampler after migrations succeed. `NewDBWithSettings` installs the counter but does not start the sampler.

Both metrics are high priority. The counter uses `UnitCount`, and the gauge uses `UnitCountAverage`. See the [SQLC metric implementation](https://github.com/gosoline-project/sqlc/blob/89ede23e657a11d223f9b24f16dc42abfbb059cc/metrics.go).

SQLC has no separate switch for these metrics. To export them, add `application.WithMetrics`, set `metric.enabled: true`, and configure at least one writer under `metric.writers`.

## Related references[​](#related-references "Direct link to Related references")

* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
* [SQL client guide](/docs/pr-5/how-to/databases-sql/sqlc/.md)
* [SQL repository package](/docs/pr-5/reference/package-db-repo/.md)
* [Working with relational databases](/docs/pr-5/how-to/databases-sql/.md)
