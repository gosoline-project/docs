# mdlsub package

The `pkg/mdlsub` package adapts model subscription messages to the stream consumer callback interface. It reads model and version details from message attributes, selects a transformer from the supplied model core, and persists transformed output.

## API and activation[​](#api-and-activation "Direct link to API and activation")

Use `NewSubscriberCallbackFactory(core, sourceModel, persistGraceTime)` to create a `stream.UntypedConsumerCallbackFactory` for a stream consumer. The `SubscriberCore` supplies known model IDs, transformers, and outputs; `persistGraceTime` controls the delayed-cancellation window while persistence is in progress.

## Metrics[​](#metrics "Direct link to Metrics")

The table lists canonical Gosoline-authored names and units. The metric writer renders them per backend: OpenTelemetry prefixes namespaced metrics with `gosoline.` and OpenTelemetry/Prometheus scale authored millisecond durations to seconds. See the [metric schema](/docs/pr-5/reference/metric-schema/.md) for backend rendering and the complete contract.

| Metric          | Instrument | Authored unit | Dimensions                                                                                                    | Description                                                                                                                                                                                                                                               |
| --------------- | ---------- | ------------- | ------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mdlsub.events` | Counter    | `count`       | `model.id`: parsed model ID; `outcome`: `applied` or `skipped`; `error.type`: `{{default}}` or `apply_failed` | Counts applied/skipped model events and errors after model-spec parsing. Successful persistence emits `applied` with the default error type; a nil transformed model emits `skipped`; a returned error after parsing emits `applied` with `apply_failed`. |

The callback parses the model specification before installing its error-metric defer, so malformed or missing model attributes do not produce a metric. Delay errors and request cancellation that the callback handles as a retry also return without an error metric. Other returned errors after parsing—including transformer lookup, output lookup, or persistence errors—are counted as `outcome=applied,error.type=apply_failed`. Thus `applied` by itself does not imply success; read it with `error.type`.

The writer seeds zero-valued defaults for `applied` and `skipped` with `error.type={{default}}` for each model ID known to the core. Those defaults are not consumed events; no zero-valued `apply_failed` series is seeded.

## Related[​](#related "Direct link to Related")

* [Create a consumer](/docs/pr-5/how-to/streaming-applications/create-a-consumer/.md)
* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
