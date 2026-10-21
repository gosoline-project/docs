# grpcserver package

The Gosoline SDK's `pkg/grpcserver` package (import path `github.com/justtrackio/gosoline/pkg/grpcserver`) wraps a gRPC server as a kernel module. `grpcserver.New(name, definer, middlewares...)` returns a module factory, reads server settings from `grpc_server.<name>`, and registers the defined services. Register the factory with the application using `application.WithModuleFactory`.

The server's statistics handler is enabled by default through `grpc_server.<name>.stats.enabled`. Disabling it disables both statistics logging and the handler's metric observations. Metric export also requires the application metric daemon: wire `application.WithMetrics`, set `metric.enabled: true`, and configure a writer under `metric.writers`.

## Metrics[​](#metrics "Direct link to Metrics")

At each gRPC stats end event, the handler records one duration observation for the RPC. The full method name is split into `service` and `method`; for a name without a slash separating a method, `method` is `{{default}}`. Errors do not create a separate status or error dimension, and the histogram observation count provides the RPC count.

| Metric                        | Instrument | Authored unit | Dimensions                                                                                                                                        | Description                                                                                                                                                     |
| ----------------------------- | ---------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `rpc.server.request.duration` | histogram  | ms            | `service` — service component of the full method name; `method` — method component, or `{{default}}` when the full method has no method separator | Duration of a completed gRPC server call. One observation is recorded for each stats end event; there is no separate request counter or status/error dimension. |

The table lists Gosoline-authored names and units. Configured writers may transform exported names and duration units; see [Metric schema](/docs/pr-5/reference/metric-schema/.md).
