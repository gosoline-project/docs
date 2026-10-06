---
sidebar:
  order: 1
title: Understand inputs and outputs
---

The `stream` package separates your application logic from the system carrying its messages. A streaming application typically has this flow:

```text
model -> Producer -> Output -> queue, topic, or stream -> Input -> Consumer -> callback
```

- An **output** writes encoded messages to a transport.
- A **producer** encodes Go values and delegates to an output.
- An **input** reads messages from a transport.
- A **consumer** decodes those messages, calls your callback, selects the retry mechanism, and handles health checks, metrics, tracing, and the shared processing shutdown deadline. The input executes transport acknowledgement or checkpointing.

Most applications should use `Producer` and a typed consumer rather than operate an `Input` or `Output` directly.

## Name each component

Inputs, outputs, producers, and consumers are configured independently and connected by name:

```yaml
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

`stream.consumer.orders.input` selects the input named `orders`. When omitted, a consumer's input defaults to the literal name `consumer`, regardless of the consumer's name. `stream.producer.notifications.output` selects the output named `notifications`; when omitted, a producer's output defaults to the producer's own name.

This indirection lets you change transports without changing application code. A producer named `orders` is created the same way whether its configured output is a file, SQS queue, SNS topic, Kinesis stream, Kafka topic, or Redis list.

## Supported transports

| Type | Input | Output | Typical use |
|---|---:|---:|---|
| `sqs` | Yes | Yes | Work queues with acknowledgement and redelivery |
| `sns` | Yes | Yes | Fan-out through an SNS topic and managed SQS subscription |
| `kinesis` | Yes | Yes | Partitioned event streams |
| `kafka` | Yes | Yes | Kafka topics and consumer groups |
| `redis` | Yes | Yes | Redis lists |
| `file` | Yes | Yes | Local development and simple fixtures |
| `inMemory` | Yes | Yes | Tests in the same process |
| `multiple` | No | Yes | Fan-out to several configured outputs |
| `noop` | No | Yes | Discard messages intentionally |

Each transport has additional settings such as `queue_id`, `topic_id`, `stream_name`, or `key`. The component name is still the stable link used by producers and consumers.

## Input and output interfaces

An input delivers `*stream.Message` values through a processing callback and exposes lifecycle methods:

```go
type InputProcess func(ctx context.Context, msg *Message) (ack bool)

type Input interface {
    Run(ctx context.Context, process InputProcess) error
    Stop(ctx context.Context)
    IsHealthy() bool
}
```

The consumer runs the configured input and supplies the decoding and business-processing callback. The input owns fetching, concurrency, and transport acknowledgement. `Stop` stops fetching and allows already fetched work to drain; finite inputs such as files can finish on their own. `Run` should only be called once.

The callback result is transport-specific: SQS/SNS can leave failed messages undeleted for native redelivery, while Kafka/Kinesis commit or checkpoint records handed to processing even when processing fails. Redis reads remove messages from the list. See [consumer retry, concurrency, and shutdown behavior](/docs/pr-6/how-to/streaming-applications/create-a-consumer) before selecting a transport.

An output accepts values that already implement `WritableMessage`:

```go
type Output interface {
    WriteOne(ctx context.Context, msg WritableMessage) error
    Write(ctx context.Context, batch []WritableMessage) error
}
```

Use `stream.NewConfigurableOutput` when you already have encoded stream messages or are building infrastructure. For ordinary domain events, use `stream.NewProducer`; it provides encoding, compression, tracing, schema options, and optional producer-daemon integration.

## Production transports

Here are minimal output configurations. Resource identity fields such as `application`, `env`, and `tags` are optional and default from `app`.

```yaml
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

```yaml
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

For Kafka-specific behavior and schema registry configuration, see the [Kafka guides](/docs/pr-6/how-to/kafka/general).

## What's next?

- [Create a producer](/docs/pr-6/how-to/streaming-applications/create-a-producer/)
- [Create a consumer](/docs/pr-6/how-to/streaming-applications/create-a-consumer/)
- [Use the producer daemon](/docs/pr-6/how-to/streaming-applications/use-the-producer-daemon/)
