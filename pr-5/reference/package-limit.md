# limit package

The `pkg/limit` package provides rate limiters for operations that must wait until a limit permits them to proceed. The common `Limiter` interface exposes `Wait(ctx, prefix) error`; a fixed-window limiter can delay a caller after capacity is reached.

## API and activation[​](#api-and-activation "Direct link to API and activation")

The metric middleware is opt-in. Attach a `limit.NewMetricMiddleware()` instance through a middleware-capable limiter's `WithMiddleware` method, using a `MiddlewareFactory` that returns it. A call to `OnTake` alone does not emit a metric; the middleware records outcome callbacks.

## Metrics[​](#metrics "Direct link to Metrics")

The table lists canonical Gosoline-authored names and units. The metric writer renders them per backend: OpenTelemetry prefixes namespaced metrics with `gosoline.` and OpenTelemetry/Prometheus scale authored millisecond durations to seconds. See the [metric schema](/docs/pr-5/reference/metric-schema/.md) for backend rendering and the complete contract.

| Metric        | Instrument | Authored unit | Dimensions                                                                                                                                                                  | Description                                                                                     |
| ------------- | ---------- | ------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| `limit.takes` | Counter    | `count`       | `name`: limiter name; `prefix`: wait prefix; `outcome`: `allowed`, `throttled`, or `error`; `error.type`: `{{default}}` for allowed/throttled, `rate_limit_error` for error | `OnRelease` records `allowed`, `OnThrottle` records `throttled`, and `OnError` records `error`. |

These are the emitted attribute keys and values: `outcome` and `error.type` carry the callback result; there are no `pass`, `reject`, or `wait` outcome values. In the fixed-window limiter, reaching capacity emits `throttled` while the call waits; a successful wait then also emits `allowed`, while cancellation during that wait emits `error`. Therefore a single `Wait` can emit a throttle event and then a terminal callback event. `error.type=rate_limit_error` is the middleware's generic operation-failure label, not a Go error type.

The `prefix` value is passed to the fixed-window limiter through `Wait`; the leaky-bucket limiter instead reports the literal `take` prefix. The middleware does not seed zero-valued defaults, so these counts come from outcome callbacks.

## Related[​](#related "Direct link to Related")

* [Concurrency and connection pressure](/docs/pr-5/how-to/http-server/concurrency-and-connection-pressure/.md)
* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
