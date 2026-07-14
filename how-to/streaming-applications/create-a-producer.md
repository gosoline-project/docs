# Create a producer

A `stream.Producer` turns Go values into stream messages and writes them to a configured output. The producer owns encoding and compression, so application code works with domain models instead of serialized messages.

## Define the producer[​](#define-the-producer "Direct link to Define the producer")

Create the producer in your module factory and retain the `stream.Producer` interface:

```
func newModule(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.Module, error) {

    producer, err := stream.NewProducer(ctx, config, logger, "orders")

    if err != nil {

        return nil, fmt.Errorf("create orders producer: %w", err)

    }



    return &producerModule{producer: producer}, nil

}
```

The name `orders` resolves configuration below `stream.producer.orders`. If that section does not specify `output`, it also selects `stream.output.orders`.

## Write models[​](#write-models "Direct link to Write models")

Use `WriteOne` for one model and `Write` for a slice:

```
if err := producer.WriteOne(ctx, OrderCreated{

    Id:    "order-1001",

    Total: 42.50,

}); err != nil {

    return fmt.Errorf("publish order: %w", err)

}



orders := []OrderCreated{

    {Id: "order-1002", Total: 18.75},

    {Id: "order-1003", Total: 91.20},

}

if err := producer.Write(ctx, orders); err != nil {

    return fmt.Errorf("publish orders: %w", err)

}
```

Both methods accept optional message attributes after the model argument:

```
err := producer.WriteOne(ctx, order, map[string]string{

    "tenant": "acme",

})
```

Attributes carry transport metadata and application routing information. Transport-specific helpers or constants should be used where available, such as Kafka keys or Kinesis partition keys.

## Configure encoding and output[​](#configure-encoding-and-output "Direct link to Configure encoding and output")

The complete example uses a file output so it runs without external infrastructure:

config.dist.yml

```
app:

  env: dev

  name: order-producer



stream:

  producer:

    orders:

      output: orders

      encoding: application/json



  output:

    orders:

      type: file

      filename: orders.jsonl

      mode: append
```

The producer defaults to JSON. Set `stream.producer.<name>.encoding` when a consumer expects another supported encoding. Compression is also configured on the producer; some transports, notably Kafka, handle it natively.

## Complete example[​](#complete-example "Direct link to Complete example")

This application publishes one order and then a batch of two orders:

main.go

```
package main



import (

	"context"

	"fmt"



	"github.com/justtrackio/gosoline/pkg/application"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/kernel"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/stream"

)



type OrderCreated struct {

	Id    string  `json:"id"`

	Total float64 `json:"total"`

}



type producerModule struct {

	producer stream.Producer

}



func main() {

	application.RunModule("producer", newModule,

		application.WithConfigFile("config.dist.yml", "yml"),

	)

}



func newModule(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.Module, error) {

	producer, err := stream.NewProducer(ctx, config, logger, "orders")

	if err != nil {

		return nil, fmt.Errorf("create orders producer: %w", err)

	}



	return &producerModule{producer: producer}, nil

}



func (m *producerModule) Run(ctx context.Context) error {

	if err := m.producer.WriteOne(ctx, OrderCreated{

		Id:    "order-1001",

		Total: 42.50,

	}); err != nil {

		return fmt.Errorf("publish order: %w", err)

	}



	orders := []OrderCreated{

		{Id: "order-1002", Total: 18.75},

		{Id: "order-1003", Total: 91.20},

	}



	if err := m.producer.Write(ctx, orders); err != nil {

		return fmt.Errorf("publish orders: %w", err)

	}



	return nil

}
```

Run it from `docs/docs/how-to/streaming-applications/src/producer`:

```
rm -f orders.jsonl

go run .

cat orders.jsonl
```

The file contains one stream message per line. Replace only the output configuration to publish the same models to SQS:

```
stream:

  output:

    orders:

      type: sqs

      queue_id: orders
```

## Producer or output?[​](#producer-or-output "Direct link to Producer or output?")

Prefer `Producer` for application events. Use `Output` directly only when the data is already represented as a `stream.WritableMessage`, for example when forwarding an encoded message without decoding it.

## What's next?[​](#whats-next "Direct link to What's next?")

* [Create a consumer](/docs/how-to/streaming-applications/create-a-consumer.md) to process the produced models
* [Use the producer daemon](/docs/how-to/streaming-applications/use-the-producer-daemon.md) to buffer, batch, or aggregate writes
