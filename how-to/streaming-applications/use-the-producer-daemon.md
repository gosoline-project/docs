# Use the producer daemon

The producer daemon is a background module placed between a `Producer` and its output. It can buffer writes, combine messages into transport batches, aggregate several stream messages into one message, compress aggregates, and write through several background runners.

Use it when fewer, larger transport requests improve throughput or cost. Do not enable it when each call must be durably written before `WriteOne` returns: a daemon write confirms that the message entered the in-process buffer, not that the remote transport accepted it.

## Enable and register the daemon[​](#enable-and-register-the-daemon "Direct link to Enable and register the daemon")

Enable the daemon for a named producer:

```
stream:

  producer:

    orders:

      daemon:

        enabled: true
```

Also register producer daemons with the application:

```
application.RunModule("producer", newModule,

    application.WithConfigFile("config.dist.yml", "yml"),

    application.WithProducerDaemon,

)
```

`stream.NewProducer` detects `daemon.enabled` and writes through the shared daemon automatically. Do not construct or write to a producer daemon directly.

## Configure buffering and batching[​](#configure-buffering-and-batching "Direct link to Configure buffering and batching")

config.dist.yml

```
app:

  env: dev

  name: order-producer-daemon



stream:

  producer:

    orders:

      output: orders

      encoding: application/json

      daemon:

        enabled: true

        interval: 1s

        buffer_size: 2

        runner_count: 1

        batch_size: 2

        aggregation_size: 3



  output:

    orders:

      type: file

      filename: orders.jsonl

      mode: append
```

The main settings are:

| Setting                  | Default  | Purpose                                                                 |
| ------------------------ | -------- | ----------------------------------------------------------------------- |
| `interval`               | `1m`     | Flush incomplete aggregates and batches after inactivity                |
| `buffer_size`            | `10`     | Number of completed batches allowed in the in-process output channel    |
| `runner_count`           | `10`     | Background writers draining completed batches                           |
| `batch_size`             | `10`     | Maximum messages in one output call                                     |
| `batch_max_size`         | `258048` | Maximum total bytes in one output batch; `0` disables this daemon limit |
| `aggregation_size`       | `1`      | Stream messages combined into one aggregate message                     |
| `aggregation_max_size`   | `65536`  | Maximum aggregate bytes; `0` disables this daemon limit                 |
| `partition_bucket_count` | `128`    | Aggregation buckets for partitioned outputs                             |
| `message_attributes`     | none     | Attributes added to emitted messages                                    |

Batching and aggregation are different:

* **Batching** sends several messages in one output API call.
* **Aggregation** serializes several stream messages into one transport message. Gosoline consumers detect and unpack these aggregates.

Leave `aggregation_size` at `1` when a non-Gosoline consumer reads the output or when the transport does not support Gosoline aggregate messages.

## Complete example[​](#complete-example "Direct link to Complete example")

The example queues five orders. With `aggregation_size: 3`, the daemon emits one aggregate after the third order and flushes the remaining aggregate during shutdown:

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

	Id string `json:"id"`

}



type producerModule struct {

	producer stream.Producer

}



func main() {

	application.RunModule("producer", newModule,

		application.WithConfigFile("config.dist.yml", "yml"),

		application.WithProducerDaemon,

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

	for i := 1; i <= 5; i++ {

		order := OrderCreated{Id: fmt.Sprintf("order-%04d", i)}

		if err := m.producer.WriteOne(ctx, order); err != nil {

			return fmt.Errorf("publish order: %w", err)

		}

	}



	return nil

}
```

Run it from `docs/docs/how-to/streaming-applications/src/producer-daemon`:

```
rm -f orders.jsonl

go run .

cat orders.jsonl
```

The file output makes the lifecycle visible without requiring a queue. In production, replace it with an SQS, SNS, Kinesis, Kafka, or Redis output.

## Backpressure and errors[​](#backpressure-and-errors "Direct link to Backpressure and errors")

When all background runners are busy and the output channel reaches `buffer_size`, producer writes block until capacity becomes available. This is intentional backpressure; size the buffer and runner count for expected traffic instead of making the buffer unbounded.

Remote output errors happen in a background runner after the producer call has returned. The daemon logs these errors and emits metrics including message count, batch size, aggregate size, and idle duration. Monitor daemon logs and metrics because callers cannot receive asynchronous output failures.

On graceful kernel shutdown, the daemon flushes pending aggregates and batches before closing its output workers. Abrupt process termination can still lose buffered messages.

## Transport limitations[​](#transport-limitations "Direct link to Transport limitations")

Output capabilities can reduce configured batch and message sizes automatically. For partitioned outputs such as Kinesis, the daemon uses partition buckets so messages with the same partition key remain assigned consistently enough for shard processing.

Kafka does not support Gosoline aggregation in the producer daemon. Kafka's own producer performs batching and compression, and its `max_batch_size` and `max_batch_bytes` settings override daemon batch limits. See [general Kafka usage](/docs/how-to/kafka/general.md#producer-daemon-usage).

## Choosing settings[​](#choosing-settings "Direct link to Choosing settings")

Start with aggregation disabled and a short interval. Enable aggregation only after confirming every consumer understands Gosoline aggregate messages. Tune using output request rate, message size, queue latency, daemon metrics, and shutdown requirements rather than increasing all limits together.
