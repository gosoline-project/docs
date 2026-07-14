# Create a consumer

A consumer connects a configured input to your callback. Gosoline owns polling, decoding, acknowledgement, retries, health checks, metrics, tracing, and graceful shutdown; your callback owns the business operation.

For a first walkthrough, see [Create a consumer](/docs/getting-started/create-a-consumer.md). This guide focuses on the current typed API and operational behavior.

## Implement a typed callback[​](#implement-a-typed-callback "Direct link to Implement a typed callback")

Define the expected message body and implement `stream.ConsumerCallback[M]`:

```
type OrderCreated struct {

    Id    string  `json:"id"`

    Total float64 `json:"total"`

}



func (c *consumer) Consume(ctx context.Context, order OrderCreated, attributes map[string]string) (bool, error) {

    c.logger.Info(ctx, "received order %s with total %.2f", order.Id, order.Total)



    return true, nil

}
```

The boolean controls acknowledgement:

| Return       | Meaning                                                                    |
| ------------ | -------------------------------------------------------------------------- |
| `true, nil`  | Processing succeeded; acknowledge the message                              |
| `false, nil` | Processing did not complete; do not acknowledge and retry where configured |
| `false, err` | Processing failed; record the error and retry where configured             |

Only return `true` after the business operation is complete. The transport determines how an unacknowledged message is redelivered. When the input has no native retry mechanism, configure the consumer retry handler.

## Run the consumer[​](#run-the-consumer "Direct link to Run the consumer")

For one callback named `default`, use `application.RunConsumer`:

```
application.RunConsumer(newConsumer,

    application.WithConfigFile("config.dist.yml", "yml"),

)
```

The generic type is inferred from the factory's `stream.ConsumerCallback[OrderCreated]` return type.

For several consumers accepting the same model, use `application.RunConsumers` with a `stream.ConsumerCallbackMap[M]`. Each map key must have a matching `stream.consumer.<name>` configuration. Prefer separate applications when callbacks process unrelated model types; use untyped consumers only when message attributes genuinely select among several models.

## Connect the consumer to an input[​](#connect-the-consumer-to-an-input "Direct link to Connect the consumer to an input")

The example maps consumer `default` to input `orders`:

config.dist.yml

```
app:

  env: dev

  name: order-consumer



stream:

  consumer:

    default:

      input: orders

      encoding: application/json



  input:

    orders:

      type: file

      filename: events.jsonl

      blocking: false
```

The file input expects one serialized `stream.Message` per line, not a bare domain object:

events.jsonl

```
{"body":"{\"id\":\"order-1001\",\"total\":42.5}"}

{"body":"{\"id\":\"order-1002\",\"total\":18.75}"}
```

For SQS, only the input section changes:

```
stream:

  consumer:

    default:

      input: orders

      retry:

        enabled: true



  input:

    orders:

      type: sqs

      queue_id: orders
```

## Complete example[​](#complete-example "Direct link to Complete example")

main.go

```
package main



import (

	"context"



	"github.com/justtrackio/gosoline/pkg/application"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/stream"

)



type OrderCreated struct {

	Id    string  `json:"id"`

	Total float64 `json:"total"`

}



type consumer struct {

	logger log.Logger

}



func main() {

	application.RunConsumer(newConsumer,

		application.WithConfigFile("config.dist.yml", "yml"),

	)

}



func newConsumer(_ context.Context, _ cfg.Config, logger log.Logger) (stream.ConsumerCallback[OrderCreated], error) {

	return &consumer{logger: logger}, nil

}



func (c *consumer) Consume(ctx context.Context, order OrderCreated, _ map[string]string) (bool, error) {

	c.logger.Info(ctx, "received order %s with total %.2f", order.Id, order.Total)



	return true, nil

}
```

Run it from `docs/docs/how-to/streaming-applications/src/consumer`:

```
go run .
```

The non-blocking file input closes after both records are consumed, so the example application exits by itself.

## Concurrency and batches[​](#concurrency-and-batches "Direct link to Concurrency and batches")

`stream.consumer.<name>.runner_count` controls concurrent processing for a normal consumer. Increase it only when callback operations are safe to run concurrently and message ordering is not required.

Use `application.RunBatchConsumer` or `RunBatchConsumers` when the callback should receive batches. Batch consumers have separate callback interfaces and settings; they are not equivalent to increasing `runner_count`.

## Graceful processing[​](#graceful-processing "Direct link to Graceful processing")

The consumer derives a delayed cancellation context for each callback. `consume_grace_time` gives in-flight processing a short grace period after shutdown begins. `acknowledge_grace_time` and retry grace settings similarly allow final acknowledgement or retry writes. Keep callback work bounded and always pass its context to downstream calls.

## What's next?[​](#whats-next "Direct link to What's next?")

* [Test your consumer](/docs/getting-started/testing/test-your-consumer.md)
* [Use the producer daemon](/docs/how-to/streaming-applications/use-the-producer-daemon.md)
