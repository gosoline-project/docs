# Chaos middleware

The httpserver includes an optional chaos middleware for resilience testing. When enabled, it randomly introduces failures to exercise client retry logic, circuit breakers, and timeout handling in non-production environments.

The middleware is applied **after** the health check route, so Kubernetes liveness and readiness probes are never affected.

## Enabling chaos[​](#enabling-chaos "Direct link to Enabling chaos")

Chaos is disabled by default. Enable it under `httpserver.<name>.chaos`:

```
httpserver:

  default:

    chaos:

      enabled: true
```

When `enabled` is `false` (the default), the middleware is a no-op with zero overhead.

## Failure scenarios[​](#failure-scenarios "Direct link to Failure scenarios")

The middleware evaluates scenarios in this order on every request:

1. **Drop** → **Delay** → **Reject** → **Slow response** → **Truncate** → normal processing

Each scenario has an independent `percent` probability (0–100). **Drop** and **reject** are terminal — if they trigger, the request ends immediately. The other scenarios combine: a single request can be delayed *and* get a slow response *and* be truncated. This produces realistic compound failures (e.g. a slow upstream that also crashes mid-response).

### Drop[​](#drop "Direct link to Drop")

Hijacks the TCP connection and closes it immediately without sending any response. The client sees `connection reset by peer` or `EOF`.

**Simulates:** OOM-killed pod, network partition, hard crash.

```
chaos:

  enabled: true

  drop:

    percent: 2
```

| Field     | Type | Default | Description                                       |
| --------- | ---- | ------- | ------------------------------------------------- |
| `percent` | int  | `3`     | Probability (0–100) that a connection is dropped. |

### Delay[​](#delay "Direct link to Delay")

Adds random latency before the request is processed. The delay is uniformly distributed between `min_duration` and `max_duration`.

**Simulates:** Garbage collection pauses, overloaded upstream, network congestion.

```
chaos:

  enabled: true

  delay:

    percent: 5

    min_duration: 1s

    max_duration: 30s
```

| Field          | Type     | Default | Description                                    |
| -------------- | -------- | ------- | ---------------------------------------------- |
| `percent`      | int      | `3`     | Probability (0–100) that a request is delayed. |
| `min_duration` | duration | `0`     | Minimum delay before processing.               |
| `max_duration` | duration | `60s`   | Maximum delay before processing.               |

The delay respects context cancellation — if the client disconnects, the delay is interrupted immediately.

### Reject[​](#reject "Direct link to Reject")

Responds with a random HTTP error status code, chosen from the configured list.

**Simulates:** Backend returning 5xx errors under load.

```
chaos:

  enabled: true

  reject:

    percent: 5

    status_codes:

      - 500

      - 502

      - 503

      - 504
```

| Field          | Type   | Default                     | Description                                                                |
| -------------- | ------ | --------------------------- | -------------------------------------------------------------------------- |
| `percent`      | int    | `3`                         | Probability (0–100) that a request is rejected.                            |
| `status_codes` | \[]int | `[499, 500, 502, 503, 504]` | Status codes to respond with. One is chosen randomly per rejected request. |

### Slow response[​](#slow-response "Direct link to Slow response")

Wraps the response writer so the response body is trickled to the client in small chunks with configurable delays between them. Headers are sent immediately, but the body takes a very long time to arrive.

**Simulates:** Overloaded upstream sending bytes slowly, defeating simple read timeouts because bytes *are* arriving.

```
chaos:

  enabled: true

  slow_response:

    percent: 3

    delay: 1s

    chunk_size: 64
```

| Field        | Type     | Default | Description                                       |
| ------------ | -------- | ------- | ------------------------------------------------- |
| `percent`    | int      | `3`     | Probability (0–100) that a response is throttled. |
| `delay`      | duration | `1s`    | Pause between each chunk written to the client.   |
| `chunk_size` | int      | `64`    | Number of bytes per chunk.                        |

### Truncate[​](#truncate "Direct link to Truncate")

Sends the response headers and a partial body, then forcibly closes the connection. The client receives a valid HTTP status and headers, starts reading the body, then gets `unexpected EOF`.

**Simulates:** Upstream crash mid-response, load balancer timeout after partial forward.

```
chaos:

  enabled: true

  truncate:

    percent: 2

    max_bytes: 512
```

| Field       | Type | Default | Description                                                                                                             |
| ----------- | ---- | ------- | ----------------------------------------------------------------------------------------------------------------------- |
| `percent`   | int  | `3`     | Probability (0–100) that a response is truncated.                                                                       |
| `max_bytes` | int  | `512`   | Maximum body bytes sent before dropping the connection. A random value between 1 and `max_bytes` is chosen per request. |

## Full configuration example[​](#full-configuration-example "Direct link to Full configuration example")

A sandbox configuration that produces a realistic mix of failures:

```
httpserver:

  default:

    chaos:

      enabled: true

      drop:

        percent: 2

      delay:

        percent: 5

        min_duration: 500ms

        max_duration: 10s

      reject:

        percent: 5

        status_codes:

          - 499

          - 500

          - 502

          - 503

          - 504

      slow_response:

        percent: 3

        delay: 1s

        chunk_size: 64

      truncate:

        percent: 2

        max_bytes: 512
```

With this configuration, roughly 17% of requests are affected by some failure (2 + 5 + 5 + 3 + 2, though the exact probability is slightly lower since earlier scenarios short-circuit later ones).

## Using chaos per route group[​](#using-chaos-per-route-group "Direct link to Using chaos per route group")

The built-in chaos middleware is applied to the entire server. If you need chaos only on specific routes, use `router.UseFactory` or `router.Use` on a group:

```
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

    chaosSettings := httpserver.ChaosSettings{

        Enabled: true,

        Drop:    httpserver.ChaosDropSettings{Percent: 10},

    }



    api := router.Group("/api")

    api.Use(httpserver.ChaosMiddleware(ctx, logger, chaosSettings))



    api.GET("/users", httpserver.BindN(listUsers))



    return nil

}
```

## Safety notes[​](#safety-notes "Direct link to Safety notes")

warning

Never enable chaos middleware in production. The middleware is designed for sandbox and staging environments only. Use environment-specific config files or environment variable overrides to ensure `chaos.enabled` is always `false` in production.

note

The `/health` endpoint is registered before the chaos middleware, so health checks are never affected — pods will not be killed by liveness probes hitting a dropped connection.
