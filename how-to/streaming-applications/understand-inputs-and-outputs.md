# Understand inputs and outputs

The `stream` package separates your application logic from the system carrying its messages. A streaming application typically has this flow:

```
model -> Producer -> Output -> queue, topic, or stream -> Input -> Consumer -> callback
```

* An **output** writes encoded messages to a transport.
* A **producer** encodes Go values and delegates to an output.
* An **input** reads messages from a transport.
* A **consumer** decodes those messages, calls your callback, and handles acknowledgement, retries, health checks, metrics, and tracing.

Most applications should use `Producer` and a typed consumer rather than operate an `Input` or `Output` directly.

## Name each component[​](#name-each-component "Direct link to Name each component")

Inputs, outputs, producers, and consumers are configured independently and connected by name:

```
stream:

  input:

    orders:

      type: sqs

      queue_id: orders



  consumer:

    orders:

      input: orders



  output:

    notifications:

      type: sns

      topic_id: notifications



  producer:

    notifications:

      output: notifications

      encoding: application/json
```

`stream.consumer.orders.input` selects the input named `orders`. `stream.producer.notifications.output` selects the output named `notifications`. When omitted, both links default to the component's own name.

This indirection lets you change transports without changing application code. A producer named `orders` is created the same way whether its configured output is a file, SQS queue, SNS topic, Kinesis stream, Kafka topic, or Redis list.

## Supported transports[​](#supported-transports "Direct link to Supported transports")

| Type       | Input | Output | Typical use                                               |
| ---------- | ----- | ------ | --------------------------------------------------------- |
| `sqs`      | Yes   | Yes    | Work queues with acknowledgement and redelivery           |
| `sns`      | Yes   | Yes    | Fan-out through an SNS topic and managed SQS subscription |
| `kinesis`  | Yes   | Yes    | Partitioned event streams                                 |
| `kafka`    | Yes   | Yes    | Kafka topics and consumer groups                          |
| `redis`    | Yes   | Yes    | Redis lists                                               |
| `file`     | Yes   | Yes    | Local development and simple fixtures                     |
| `inMemory` | Yes   | Yes    | Tests in the same process                                 |
| `multiple` | No    | Yes    | Fan-out to several configured outputs                     |
| `noop`     | No    | Yes    | Discard messages intentionally                            |

Each transport has additional settings such as `queue_id`, `topic_id`, `stream_name`, or `key`. The component name is still the stable link used by producers and consumers.

## Input and output interfaces[​](#input-and-output-interfaces "Direct link to Input and output interfaces")

An input exposes a channel of `*stream.Message` and lifecycle methods:

```
type Input interface {

    Run(ctx context.Context) error

    Stop(ctx context.Context)

    Data() <-chan *Message

    IsHealthy() bool

}
```

The consumer owns this lifecycle. It runs the configured input, decodes its messages, and acknowledges them when the transport supports acknowledgement.

An output accepts values that already implement `WritableMessage`:

```
type Output interface {

    WriteOne(ctx context.Context, msg WritableMessage) error

    Write(ctx context.Context, batch []WritableMessage) error

}
```

Use `stream.NewConfigurableOutput` when you already have encoded stream messages or are building infrastructure. For ordinary domain events, use `stream.NewProducer`; it provides encoding, compression, tracing, schema options, and optional producer-daemon integration.

## Production transports[​](#production-transports "Direct link to Production transports")

Here are minimal output configurations. Resource identity fields such as `application`, `env`, and `tags` are optional and default from `app`.

```
stream:

  output:

    jobs:

      type: sqs

      queue_id: jobs



    events:

      type: sns

      topic_id: events



    records:

      type: kinesis

      stream_name: records



    audit:

      type: kafka

      topic_id: audit



    cache_updates:

      type: redis

      server_name: default

      key: cache-updates
```

Input configuration follows the same naming pattern. For example:

```
stream:

  input:

    jobs:

      type: sqs

      queue_id: jobs



    records:

      type: kinesis

      application: record-producer

      stream_name: records



    audit:

      type: kafka

      topic_id: audit

      group_id: audit-worker
```

For Kafka-specific behavior and schema registry configuration, see the [Kafka guides](/docs/how-to/kafka/general.md).

## What's next?[​](#whats-next "Direct link to What's next?")

* [Create a producer](/docs/how-to/streaming-applications/create-a-producer.md)
* [Create a consumer](/docs/how-to/streaming-applications/create-a-consumer.md)
* [Use the producer daemon](/docs/how-to/streaming-applications/use-the-producer-daemon.md)
