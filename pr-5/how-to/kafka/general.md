# Notes on general Kafka usage

In general, using Kafka with gosoline works like any other input/output type supported by the `stream` package.

However, there are some caveats and minor differences to be considered as described below.

## Connection configuration[​](#connection-configuration "Direct link to Connection configuration")

Define broker connections under `kafka.connection.<name>`. Kafka inputs and outputs select them with their `connection` setting, which defaults to `default`:

```
kafka:

  connection:

    default:

      brokers:

        - kafka.example.com:9093

      tls_enabled: true

      dial_timeout: 10s



stream:

  output:

    audit:

      type: kafka

      connection: default

      topic_id: audit
```

| Connection setting     | Default  | Purpose                                                     |
| ---------------------- | -------- | ----------------------------------------------------------- |
| `brokers`              | required | Bootstrap broker addresses                                  |
| `tls_enabled`          | `true`   | Enable broker TLS, with a minimum version of TLS 1.2        |
| `insecure_skip_verify` | `false`  | Skip TLS certificate and hostname verification when enabled |
| `dial_timeout`         | `10s`    | Timeout for establishing a broker connection                |
| `username`, `password` | unset    | Enable SASL SCRAM-SHA-512 when **both** are nonempty        |

Supply credentials through your deployment's configuration/secrets mechanism. Setting only one credential does not enable SASL. For a local plaintext broker, set `tls_enabled: false` explicitly. These settings configure broker clients; schema registry HTTP configuration is covered in the [schema registry guide](/docs/pr-5/how-to/kafka/use-schema-registry/.md#configure-the-registry-connection).

`topic_id` and `group_id` are formatted through Gosoline's naming patterns rather than necessarily used verbatim. See the [Kafka naming reference](/docs/pr-5/reference/naming-patterns/.md#kafka) for topic/group patterns, delimiters, and resource identity overrides.

## Message format, keys, and headers[​](#message-format-keys-and-headers "Direct link to Message format, keys, and headers")

Kafka record values contain the encoded model directly, not a JSON envelope containing Gosoline `Body` and `Attributes` fields. Without a schema registry, a JSON producer writes the domain object's JSON as the value. With a schema registry, the value includes the registry serializer's framing and schema ID.

Use `stream.AttributeKafkaKey` (the attribute name `KafkaKey`) to set the record key. Other message attributes become Kafka headers:

```
err := producer.WriteOne(ctx, order, map[string]string{

    stream.AttributeKafkaKey: order.Id,

    "tenant":                 "acme",

})
```

franz-go uses the key for partition routing. Records with the same key are routed consistently while the topic's partition layout is unchanged; changing the partition count can change their destination. A key alone does not serialize callback execution: select `processing_mode: ordered` when processing order within a partition matters.

On consumption, the record value becomes the message body, headers become string attributes, and a non-nil record key is exposed as `stream.AttributeKafkaKey`:

```
func (c *consumer) Consume(ctx context.Context, order OrderCreated, attributes map[string]string) (bool, error) {

    key := attributes[stream.AttributeKafkaKey]

    tenant := attributes["tenant"]

    c.logger.Info(ctx, "received order %s with key %s for tenant %s", order.Id, key, tenant)



    return true, nil

}
```

Header order is not preserved by the attributes map. Duplicate incoming header names collapse to the last value. `KafkaKey` is reserved: output uses it as the record key rather than a header, and an incoming non-nil key overrides any header of that name. Keys and headers are exposed as strings; topic, partition, offset, and timestamp are not automatically added to callback attributes.

## Consumer configuration[​](#consumer-configuration "Direct link to Consumer configuration")

Consumer settings select the input, encoding, application retry handler, and processing shutdown deadline. Kafka-specific settings belong to the named input:

```
stream:

  consumer:

    audit:

      input: audit

      grace_time: 10s

      retry:

        enabled: true



  input:

    audit:

      type: kafka

      topic_id: audit

      group_id: audit-worker

      runner_count: 4

      processing_mode: ordered

      max_poll_records: 100

      consume_delay: 0s

      rebalance_timeout: 60s

      grace_time: 10s
```

This example uses a separate SQS retry queue for application failures, so it also requires AWS/SQS configuration. Kafka polling and offset-commit retries are independent of this application retry setting.

| Input setting       | Default     | Purpose                                                                                                                             |
| ------------------- | ----------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `runner_count`      | `1`         | Callback concurrency limit for this input instance; in ordered mode, limits concurrently processed partitions. Must be at least `1` |
| `processing_mode`   | `unordered` | `ordered` processes records sequentially within each topic-partition, with concurrency across partitions                            |
| `max_poll_records`  | `100`       | Maximum records returned by one poll; callbacks still receive individual records                                                    |
| `consume_delay`     | `0`         | Minimum record age before processing                                                                                                |
| `rebalance_timeout` | `60s`       | Consumer-group rebalance budget                                                                                                     |
| `grace_time`        | `10s`       | Offset-commit grace window after processing finishes                                                                                |
| `start_offset`      | `last`      | Starting/reset position when the group has no usable committed offset; `first` starts at the beginning                              |

### Runner count and processing order[​](#runner-count-and-processing-order "Direct link to Runner count and processing order")

`runner_count` is a concurrency limit for one Kafka input instance, not a count of consumer-group members or a setting that changes partition assignment. Its effective concurrency depends on `processing_mode`:

* **`unordered`:** each record is a separate work unit. Up to `runner_count` callbacks can execute concurrently, including multiple records from the same partition. Callback completion order is not guaranteed.
* **`ordered`:** each topic-partition's records from the current poll form one work unit. A runner processes those records sequentially before releasing its slot. At most one callback per partition runs at a time, while different partitions can run concurrently.

In ordered mode, maximum callback concurrency is:

```
min(runner_count, assigned partitions with records in the current poll)
```

| Ordered-mode configuration            | Behavior                                                                                                    |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| 3 assigned partitions, 8 runners      | At most 3 callbacks run concurrently. Extra runner capacity goes unused; partition ordering remains intact. |
| 8 assigned partitions, 3 runners      | Up to 3 partitions are processed concurrently. Other partition work units wait for a slot.                  |
| Several assigned partitions, 1 runner | Partition work units execute sequentially, but their cross-partition order is not guaranteed.               |

Only partitions assigned to this instance matter, not the topic's total partition count. Concurrency can be lower if some assigned partitions have no records in the current poll. In unordered mode, eight runners can process eight records concurrently even from a single partition.

A slow partition holds its ordered-mode slot until its fetched records finish processing. Reducing runner count can therefore increase waiting and lag for other partitions. Gosoline waits for all admitted work from the current poll to finish and commits before polling again.

Keep callbacks safe for concurrent use even in ordered mode. A separate retry input can invoke the same callback independently, adding concurrency beyond the Kafka runner limit; ordered Kafka processing does not order work redelivered through that retry queue.

Gosoline waits for admitted work from a poll to finish before committing and allowing a rebalance. Tune poll size, runner count, delay, and callback duration together so the full processing and commit cycle fits the rebalance budget. Consumer poll batches are independent of producer-daemon batches.

### Isolation and group tuning[​](#isolation-and-group-tuning "Direct link to Isolation and group tuning")

These additional settings belong under `stream.input.<name>`:

| Setting                 | Default                | Purpose                                                                                                           |
| ----------------------- | ---------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `fetch_isolation_level` | `read_uncommitted`     | Use `read_committed` to exclude aborted transactional records and wait for transactions to resolve                |
| `balancers`             | `[cooperative-sticky]` | Ordered list of supported partition assignment strategies: `cooperative-sticky`, `sticky`, `round-robin`, `range` |
| `session_timeout`       | `45s`                  | Consumer-group session timeout for detecting missing members                                                      |
| `heartbeat_interval`    | `3s`                   | Interval between group heartbeats; keep below the session timeout                                                 |
| `idle_wait_time`        | `500ms`                | Additional wait after a poll returns no records                                                                   |
| `healthcheck.timeout`   | `5m`                   | Input progress health timeout                                                                                     |

Choose balancers compatible with the other members of the group and timing values accepted by the broker. `read_committed` concerns visibility of transactional Kafka records; it does not make callbacks, retries, and offset commits an atomic transaction or provide exactly-once business processing.

Idle polling, rebalancing, and intentional consume delays are handled so they do not by themselves indicate an unhealthy input. Consumer-level `stream.consumer.<name>.healthcheck.timeout` separately checks each active message's processing duration (default `5m`); one stuck callback can make the consumer unhealthy even while other callbacks finish. Health checks report a problem rather than imposing a callback deadline.

### Delayed consumption[​](#delayed-consumption "Direct link to Delayed consumption")

`consume_delay` holds a record until its timestamp is old enough; it does not add a fixed sleep to every record. Old backlog passes immediately. For Kafka `CreateTime` timestamps, age depends on the producer's clock. Future timestamps are clamped so the wait is at most `consume_delay`; records without timestamps are not delayed.

The delay must be nonnegative and strictly less than `rebalance_timeout`. That validation alone does not ensure the whole poll batch completes within the rebalance budget. Shutdown can interrupt the delay and allow already admitted work to process early.

## Offsets, failures, and retries[​](#offsets-failures-and-retries "Direct link to Offsets, failures, and retries")

Kafka inputs join a consumer group and disable automatic offset commits. Gosoline explicitly commits offsets after processing the admitted records from a poll.

**A successful offset commit does not mean the business operation succeeded.** Records handed to processing are commit-eligible even when the callback returns `false` or panics. Returning `false` does not rewind the Kafka partition or prevent its offset commit.

Enable `stream.consumer.<name>.retry.enabled` for application failures that should be retried. This sends failed callback messages to a separate retry queue rather than seeking Kafka offsets backwards. Retry publication is not atomic with committing Kafka offsets: publication failure is logged and does not prevent the commit. Model-selection or decoding errors return before the normal callback-retry path, so malformed records are not guaranteed to reach that queue. Make processing idempotent, since failed commits or process termination can also cause Kafka redelivery.

Missing topics fail consumption immediately. Retryable partition fetch errors are logged while records accompanying the errors are retained for processing; nonretryable fetch errors terminate consumption. Fetch and commit operations use their respective retry executors and the input's `backoff` settings. These retries address transport failures, not failed business operations.

`start_offset` supplies the starting/reset offset when no usable committed group offset exists. It does not override valid committed offsets on every startup.

## Shutdown[​](#shutdown "Direct link to Shutdown")

`stream.consumer.<name>.grace_time` is the shared processing deadline after shutdown begins, including processing from the retry input. Kafka stops admitting new work and drains admitted work. When that deadline expires, callback contexts are canceled.

The input's separate `stream.input.<name>.grace_time` grants final offset commits a cancellation grace window after processing finishes. It does not extend the callback processing deadline. If that commit window expires, Gosoline logs possible redelivery and completes cancellation-related shutdown normally.

Shutdown commits preserve a gap-free processed prefix within each partition. The first record of an admitted work unit may still reach processing after Stop to avoid an offset gap. Records handed to processing remain commit-eligible even when their callback fails or its context is canceled. A callback that ignores cancellation can still block completion; size kernel and deployment shutdown timeouts for processing plus final commits.

## Producer configuration[​](#producer-configuration "Direct link to Producer configuration")

Encoding and compression are producer settings; broker batching and request settings belong to the output:

```
stream:

  producer:

    audit:

      output: audit

      encoding: application/json

      compression: application/zstd



  output:

    audit:

      type: kafka

      connection: default

      topic_id: audit

      linger_timeout: 0s

      request_timeout: 10s

      max_batch_size: 10000

      max_batch_bytes: 1000012
```

| Output setting    | Default   | franz-go behavior                                                                                      |
| ----------------- | --------- | ------------------------------------------------------------------------------------------------------ |
| `linger_timeout`  | `0s`      | Time to wait for additional records before sending a batch                                             |
| `request_timeout` | `10s`     | Timeout for a Kafka produce request; not an overall deadline for the application write                 |
| `max_batch_size`  | `10000`   | Maximum buffered record count (`MaxBufferedRecords`), not a fixed number of records per broker request |
| `max_batch_bytes` | `1000012` | Maximum per-partition producer batch bytes (`ProducerBatchMaxBytes`), including record/batch overhead  |

franz-go splits records into partition-specific batches; one `producer.Write` call need not become one broker request. Match batch bytes to the broker/topic size limits and allow for overhead. The producer daemon also derives its size limits from these output settings; see [Producer Daemon usage](#producer-daemon-usage).

### Write completion and failures[​](#write-completion-and-failures "Direct link to Write completion and failures")

Without the producer daemon, Kafka output uses `ProduceSync`: `WriteOne` and `Write` wait for the records' produce results and return an error if any result fails. Pass a bounded context for an application-level deadline; `request_timeout` alone does not bound the complete operation, including buffering and retries.

A batch can partially succeed. Gosoline returns the first produce error, not per-record results, while metrics count successful and failed records separately. An error is not evidence that the whole batch was rejected. Retrying the entire batch can duplicate records that succeeded; use idempotent business identifiers or downstream deduplication where retries are possible. The streaming output API does not expose transactional writes.

With the producer daemon, the application write confirms admission to its in-process buffer. Remote produce results arrive later in background workers and failures are logged rather than returned to the original caller. See [producer-daemon backpressure and errors](/docs/pr-5/how-to/streaming-applications/use-the-producer-daemon/.md#backpressure-and-errors).

## Compression[​](#compression "Direct link to Compression")

Compression is handled entirely by the franz-go kafka library.

You still need to set the compression type in your producer config as usual but gosoline itself will not compress your messages and let the kafka library handle it.

In addition to the already natively supported compression type (`application/gzip`), you also have the option to select one of the following when writing to kafka:

* `application/snappy`
* `application/lz4`
* `application/zstd`

By default, no compression will be used.

## Producer Daemon usage[​](#producer-daemon-usage "Direct link to Producer Daemon usage")

When enabling the producer daemon for a Kafka producer all producer daemon settings related to aggregation and batching will be ignored.

Aggregation is not supported with Kafka as an aggregated message would not match the schema in case you are using the schema registry. Because of the way gosoline initializes the producer daemon it is also not possible to use aggregation without schema registry (as by the time the producer is aware of the schema registry usage the producer daemon will have already been initialized).

Batching will still be done by the producer daemon, but it will use the `max_batch_size` and `max_batch_bytes` settings specified on the `KafkaOutputConfiguration`. This is to have a single source of truth for the settings and because the Kafka library has an internal process to batch messages and would possibly break up our batches if we ignore these settings.

## Operations[​](#operations "Direct link to Operations")

### Metrics[​](#metrics "Direct link to Metrics")

Enable Gosoline's metric collection and a writer to export Kafka metrics. The following are framework metric names; the exported naming/representation depends on the configured writer:

| Area                               | Metrics                                                                                                                                         |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Consumer                           | `RecordsConsumed`, `RecordsConsumedFailed`, `PollCount`, `PollDuration`, `ProcessDuration`, `CommitDuration`, `RebalanceCount`, `SleepDuration` |
| Producer                           | `RecordsSent`, `RecordsSentFailed`, `ProduceBatchSize`, `ProduceDuration`                                                                       |
| Broker connectivity and throttling | `BrokerConnects`, `BrokerConnectsFailed`, `BrokerThrottleCount`, `BrokerThrottleTime`                                                           |
| Wire batches                       | `ProduceBatchRecords`, `ProduceBatchBytes`, `ProduceBatchBytesCompressed`, `FetchBatchRecords`, `FetchBatchBytes`, `FetchBatchBytesCompressed`  |

Client/topic metrics use `ClientType`, `Client`, and `Topic` dimensions. Wire-batch metrics include both topic totals and partition-level values with `Partition`; broker metrics use `Broker`. `Client` is the configured input/output name. Consumed counts and committed offsets do not establish business success; correlate failure counts, retry logs, and commit errors. Monitor application consumer metrics alongside Kafka-specific metrics for decoding and callback failures.

### Topic creation and metadata[​](#topic-creation-and-metadata "Direct link to Topic creation and metadata")

Kafka inputs and outputs register resource lifecycle managers. When `resource_lifecycles.create.enabled` is enabled, they check for the resolved topic and create it if absent with **one partition, replication factor one, and no topic-specific configuration**. Existing topics are left unchanged. Provision production topics separately when different partitioning, replication, or retention settings are needed.

The lifecycle admin client currently receives broker addresses only; it does not inherit the connection's TLS/SASL settings. For secured brokers, provision topics externally and disable lifecycle creation rather than expecting it to use the authenticated reader/writer connection.

When lifecycle registration is enabled (`resource_lifecycles.register.enabled`, default `true`), application metadata includes `kafka.consumers` and `kafka.producers`. Entries contain the configured component `name`, resolved `topic`, and `bootstrap_servers`. This metadata describes registered resources, not live consumer lag or delivery success.

## Schema Registry[​](#schema-registry "Direct link to Schema Registry")

You can optionally use the Kafka schema registry.

Check this guide for more details on the schema registry usage: [How to use the Kafka Schema Registry](/docs/pr-5/how-to/kafka/use-schema-registry/.md)
