# Concurrency and connection pressure

The httpserver provides three complementary mechanisms to protect your service under load:

1. **Concurrent request limiting** — caps the number of requests being handled at the same time
2. **Connection pressure management** — caps the number of open TCP connections and proactively reclaims idle ones when the limit is approached
3. **Connection lifecycle advisor** — closes long-lived or high-traffic connections to force reconnection, distributing load across pods in Kubernetes

All three are configured under `httpserver.<name>` and are disabled by default (except the lifecycle advisor, which is enabled).

## Limiting concurrent requests[​](#limiting-concurrent-requests "Direct link to Limiting concurrent requests")

`max_requests` sets the maximum number of requests that can be handled concurrently. When the limit is reached, new requests are rejected immediately — they are never queued.

```
httpserver:

  default:

    concurrency:

      max_requests: 200

      overload_status_code: 503

      retry_after: 1s
```

When the limit is hit, the server responds with `overload_status_code` (default `503`) and a JSON body:

```
{"error": "server overloaded"}
```

If `retry_after` is set, a `Retry-After` header is included in the response with the value rounded up to the nearest second. This lets clients and upstream proxies back off before retrying.

Set `max_requests: 0` (the default) to disable this limit.

| Field                  | Type     | Default | Description                                                        |
| ---------------------- | -------- | ------- | ------------------------------------------------------------------ |
| `max_requests`         | int      | `0`     | Maximum concurrent requests. `0` disables the limit.               |
| `overload_status_code` | int      | `503`   | HTTP status code returned when the limit is reached.               |
| `retry_after`          | duration | `0`     | Value for the `Retry-After` response header. `0` omits the header. |

note

The `/health` endpoint is registered before the concurrency middleware, so health checks are never blocked by the request limit.

## Connection pressure management[​](#connection-pressure-management "Direct link to Connection pressure management")

`max_connections` sets a target ceiling on open TCP connections. When this limit is reached, the server does not immediately reject new connections. Instead it proactively closes the **oldest idle connection** to free a slot before accepting the new one.

```
httpserver:

  default:

    concurrency:

      max_connections: 1000
```

The mechanism works at the listener level:

1. When a new connection arrives and the limit is already reached, the server attempts to close the oldest idle (keep-alive) connection first.
2. If no idle connections are available, the accept blocks until a connection slot is freed — either by an existing connection closing normally, or by a connection transitioning to idle.
3. This repeats until a slot is available.

This approach prevents connection exhaustion while maximising connection reuse: active connections are never forcibly closed, only idle ones are reclaimed under pressure.

Set `max_connections: 0` (the default) to disable this limit.

| Field             | Type | Default | Description                                                                   |
| ----------------- | ---- | ------- | ----------------------------------------------------------------------------- |
| `max_connections` | int  | `0`     | Target maximum open connections. `0` disables connection pressure management. |

### Metrics[​](#metrics "Direct link to Metrics")

The server emits two gauge metrics (sampled every 10 seconds and on every change):

| Metric                   | Description                                 |
| ------------------------ | ------------------------------------------- |
| `HttpConcurrentRequests` | Number of requests currently being handled. |
| `HttpOpenConnections`    | Number of currently open TCP connections.   |

Both metrics carry a `ServerName` dimension set to the server name (e.g. `default`).

## Connection lifecycle advisor[​](#connection-lifecycle-advisor "Direct link to Connection lifecycle advisor")

In Kubernetes, TCP keep-alive connections bypass the load balancer after the initial handshake. If a pod is scaled up, existing clients keep sending requests to the old pods unless the connection is closed and re-established. The lifecycle advisor solves this by setting a `Connection: close` header once a connection has either been open too long or served too many requests, prompting the client to reconnect — and land on any available pod.

```
httpserver:

  default:

    connection_lifecycle:

      enabled: true

      max_age: 1m

      max_request_count: 0
```

The advisor tracks each client by remote address. After every request it checks the two conditions:

* **`max_age`** — if the connection has been active for longer than `max_age`, the next response includes `Connection: close`. The client must open a new TCP connection for the next request.
* **`max_request_count`** — if the connection has served at least `max_request_count` requests, the same `Connection: close` header is added.

Either condition triggers the close; both can be set simultaneously. A condition of `0` disables that check.

After signalling closure the advisor resets its tracking for that remote address, so the client can reconnect and start a fresh lifecycle.

| Field               | Type     | Default | Description                                                                                |
| ------------------- | -------- | ------- | ------------------------------------------------------------------------------------------ |
| `enabled`           | bool     | `true`  | Enable/disable the connection lifecycle advisor.                                           |
| `max_age`           | duration | `1m`    | Maximum connection age before signalling close. `0` disables age-based closing.            |
| `max_request_count` | int      | `0`     | Maximum requests per connection before signalling close. `0` disables count-based closing. |

note

The `Connection: close` header works for both HTTP/1.1 and HTTP/2. For HTTP/1.1 the client closes and reopens the TCP connection. For HTTP/2 the client initiates a graceful stream reset.

## Combined configuration example[​](#combined-configuration-example "Direct link to Combined configuration example")

```
httpserver:

  default:

    concurrency:

      max_requests: 200

      max_connections: 1000

      overload_status_code: 503

      retry_after: 1s

    connection_lifecycle:

      enabled: true

      max_age: 1m

      max_request_count: 500
```

This configuration:

* Rejects requests beyond 200 concurrent handlers with a `503` and a `Retry-After: 1` header
* Allows at most 1000 open connections, reclaiming idle ones under pressure
* Signals `Connection: close` after 1 minute or 500 requests, whichever comes first
