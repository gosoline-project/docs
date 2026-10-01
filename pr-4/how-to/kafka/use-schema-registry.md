# How to use the Kafka Schema Registry

Gosoline provides the option to use the Kafka schema registry when reading from or writing to a Kafka topic. The registry checks schema registration against the subject's configured compatibility mode. Gosoline resolves the supplied schema and configures its serializer during initialization; this is not broker-side validation of every message.

Using the schema registry is optional; without it, Gosoline uses the configured encoding without registry compatibility checks.

Moreover, keep in mind that the Apache Kafka broker does not have any server side schema validation and will still let you write anything to a topic if you do not configure gosoline to use the schema registry, even if there are already existing consumers and producers of that topic which do use the schema registry.

## Configure the registry connection[​](#configure-the-registry-connection "Direct link to Configure the registry connection")

Set `schema_registry_address` on the Kafka connection selected by the input or output. Both select `default` unless their `connection` setting specifies another name:

```
kafka:

  connection:

    analytics:

      brokers:

        - kafka.example.com:9093

      tls_enabled: true

      schema_registry_address: https://schemas.example.com



stream:

  output:

    producerName:

      type: kafka

      connection: analytics

      topic_id: example-events



  input:

    example-events:

      type: kafka

      connection: analytics

      topic_id: example-events

      group_id: example-reader



  consumer:

    default:

      input: example-events
```

The registry uses an HTTP client configured from this URL; Kafka broker `username`, `password`, `tls_enabled`, and `insecure_skip_verify` are not forwarded to it. Broker SASL credentials therefore do not authenticate registry requests. See [connection configuration](/docs/pr-4/how-to/kafka/general/.md#connection-configuration) for broker settings.

## Registering a schema[​](#registering-a-schema "Direct link to Registering a schema")

By default, Gosoline looks up the supplied schema and expects it to be registered already. A missing schema causes initialization to fail. Register schemas and check compatibility externally, for example in a CI/CD pipeline, to detect incompatibilities before deploying.

Alternatively, set `AutoRegister: true` on `stream.SchemaSettings` to look up or create the schema during producer or consumer initialization:

```
schemaSettings := stream.SchemaSettings{

    Subject:      "exampleEvent",

    Schema:       exampleEventSchema,

    Model:        &exampleEvent{},

    AutoRegister: true,

}
```

This opt-in requires registry write permissions and can fail initialization if the registry rejects the schema under its compatibility policy. Consumers can opt in through the settings returned by `GetSchemaSettings()` as well.

## Configure a Producer/Publisher to use the schema registry[​](#configure-a-producerpublisher-to-use-the-schema-registry "Direct link to Configure a Producer/Publisher to use the schema registry")

As usual, you need to specify the encoding in your config according to your schema type.

In addition to the already natively supported `application/json` and `application/x-protobuf` encodings, you also have the option to use `application/avro` when using the Kafka schema registry.

```
stream:

  producer:

    producerName:

      encoding: application/avro
```

Now you need to provide the schema settings via the `stream.WithSchemaSettings` option when creating your Producer/Publisher.

```
//go:embed ExampleEvent.avsc

var exampleEventSchema string



schemaSettings := stream.SchemaSettings{

    Subject: "exampleEvent",

    Schema:  exampleEventSchema,

    Model:   &exampleEvent{},

}



producer, err := stream.NewProducer(ctx, config, logger, "producerName", stream.WithSchemaSettings(schemaSettings))
```

Where the `ExampleEvent.avsc` file would contain the avro schema.

Note that depending on whether your model is specified as a struct (`exampleEvent{}`) or a pointer to a struct (`&exampleEvent{}`) you will also need to write your events accordingly via `producer.Write`/`producer.WriteOne` because the internal serializer of the franz-go kafka library will otherwise complain that there is no registered encoder for your model if you mix it up.

## Protobuf message selection[​](#protobuf-message-selection "Direct link to Protobuf message selection")

For `application/x-protobuf`, `SchemaSettings.ProtobufMessageIndex` selects the message within the registered `.proto` schema. Indices are zero-based and follow declaration order. Omitting the field selects `[0]`, the first top-level message.

For example, this schema contains two top-level messages:

```
syntax = "proto3";

package events;



message OrderCreated {

  string id = 1;

}



message OrderCancelled {

  string id = 1;

}
```

To write `OrderCancelled`, use index `[1]` and the matching generated Go model (here imported as `eventspb`):

```
//go:embed events.proto

var eventsSchema string



schemaSettings := stream.SchemaSettings{

    Subject:              "order-events",

    Schema:               eventsSchema,

    Model:                &eventspb.OrderCancelled{},

    ProtobufMessageIndex: []int{1},

}



producer, err := stream.NewProducer(ctx, config, logger, "producerName",

    stream.WithSchemaSettings(schemaSettings),

)
```

Set `stream.producer.producerName.encoding: application/x-protobuf` and write `*eventspb.OrderCancelled` values. The consumer must select the matching index/model in `GetSchemaSettings()` and use the same encoding. For nested messages, provide the full declaration path, for example `[1, 0]` for the first nested message inside the second top-level message. The path is encoded in the registry wire framing; it is not a topic partition, schema version, or Protobuf field number.

## Configure a Consumer to use the schema registry[​](#configure-a-consumer-to-use-the-schema-registry "Direct link to Configure a Consumer to use the schema registry")

Configure the encoding in your config.

```
stream:

  consumer:

    default:

      encoding: application/avro
```

Like the producer, the consumer needs to provide the schema settings but the consumer has to do this by implementing the `stream.SchemaSettingsAwareCallback` interface, i.e. by implementing the `GetSchemaSettings()` method.

```
func (c consumer) GetSchemaSettings() (*stream.SchemaSettings, error) {

	return &stream.SchemaSettings{

        Subject: "exampleEvent",

        Schema:  exampleEventSchema,

        Model:   &exampleEvent{},

    }, nil

}
```

Implement this method on the callback passed to the typed or untyped consumer factory. The typed callback wrapper forwards the schema settings automatically.

Schema registration compatibility and historical-record decoding are separate concerns. Gosoline registers the resolved schema ID and model with its serializer during initialization; do not assume that every historical schema version is automatically registered for decoding. Verify representative old records against your consumer before deployment.

Kafka offsets are committed for records handed to processing even if decoding or the callback fails. Decoding errors occur before the normal callback-retry path, so `retry.enabled` alone does not guarantee retention of malformed or incompatible records. See [offsets, failures, and retries](/docs/pr-4/how-to/kafka/general/.md#offsets-failures-and-retries).

## Configure a Subscriber to use the schema registry[​](#configure-a-subscriber-to-use-the-schema-registry "Direct link to Configure a Subscriber to use the schema registry")

This case works similar to the consumer case.

Again configure the encoding in your config.

```
mdlsub:

  subscribers:

    targetModelName:

      input: kafka

      output: ddb

      source: { group: source-group, name: exampleEvent }



stream:

  consumer:

    subscriber-targetModelName:

      encoding: application/avro
```

And again you need to provide the schema settings by implementing the `stream.SchemaSettingsAwareCallback` interface but in this case the interface has to be implemented by your transformer.

```
func (t transformer) GetSchemaSettings() (*stream.SchemaSettings, error) {

	return &stream.SchemaSettings{

        Subject: "exampleEvent",

        Schema:  exampleEventSchema,

        Model:   &exampleEvent{},

    }, nil

}
```

Gosoline allows only one transformer per input model when using the schema registry. Evolve the schema and transformer together, and verify that the deployed transformer can handle the records it will encounter.
