# Migrating to the Callback-Based Stream Consumer Lifecycle

This guide covers the breaking stream changes introduced in Gosoline **v0.66.0** by commit [`04b54462`](https://github.com/justtrackio/gosoline/commit/04b54462dfb9cfa789577d3972d6098cd7d69860), **stream: unify input processing and shutdown**. Follow this guide when upgrading from an earlier release to v0.66.0 or later.

Previously, inputs delivered messages through `Input.Data()`, and the consumer controlled callback concurrency and acknowledgement. Inputs now invoke a processing callback directly and own acknowledgement, offset commits, or checkpoints. The consumer supplies a shared shutdown processing deadline for its primary and retry inputs.

Normal typed and untyped application callbacks retain their `Consume` contract. Most applications need configuration and behavior changes rather than a callback rewrite. Batch consumers, custom inputs, and direct Kafka or Kinesis integrations also have API changes.

## 1. Discover Consumers and Their Inputs[​](#1-discover-consumers-and-their-inputs "Direct link to 1. Discover Consumers and Their Inputs")

Before editing, map each `stream.consumer.<name>.input` to its `stream.input.<input-name>` configuration. Include retry handlers, model subscribers, custom input factories, tests, and every deployment configuration source: YAML, Helm values, Terraform templates, and environment-variable mappings.

Classify these findings:

| Finding                                            | Migration action                                                                                                     |
| -------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Consumer `consume_grace_time`                      | Rename to consumer `grace_time`, then review the shared deadline.                                                    |
| Consumer `runner_count`                            | Move the desired concurrency to the referenced input, after reviewing existing input settings and retry concurrency. |
| Consumer `acknowledge_grace_time`                  | Remove; use transport-specific cleanup settings as described below.                                                  |
| Batch consumer APIs or `batch_size`                | Choose record-at-a-time processing or application-owned batching.                                                    |
| Kafka connection `is_read_only`                    | Select replacement behavior before removing the setting.                                                             |
| File input `blocking`                              | Remove and review completion at end of file.                                                                         |
| Custom `Input`, `Data`, `Ack`, or `AckBatch` usage | Review the integration API changes before migrating.                                                                 |

An input configuration name is not necessarily a shared input instance. Normal consumer construction calls `NewConfigurableInput`; `ProvideConfigurableInput` caches instances, and in-memory wiring can also share them. If several consumers reference the same configuration, check both their desired settings and their actual instance wiring. Kafka and SQS inputs cannot simply be run twice on the same instance.

## 2. Migrate Consumer Settings[​](#2-migrate-consumer-settings "Direct link to 2. Migrate Consumer Settings")

### Processing Grace Time[​](#processing-grace-time "Direct link to Processing Grace Time")

Replace `stream.consumer.<name>.consume_grace_time` with `stream.consumer.<name>.grace_time`. Use the old duration as the initial value, but review its new meaning.

The old setting delayed cancellation separately for each callback. The new setting, defaulting to `10s`, is one shared processing window for both primary and retry work. During shutdown, the consumer calls `Stop` on both inputs and then starts the drain timer. Already-fetched work may still be passed to callbacks. When the timer expires, callback contexts are cancelled.

This is cooperative cancellation, not a hard execution cutoff. A callback that ignores cancellation can keep shutdown waiting. A custom input whose `Stop` blocks can delay the timer itself.

### Concurrency[​](#concurrency "Direct link to Concurrency")

Remove `stream.consumer.<name>.runner_count`. Callback concurrency belongs to the input at `stream.input.<input-name>.runner_count`.

| Input      | Meaning of input `runner_count`                                                       |
| ---------- | ------------------------------------------------------------------------------------- |
| SQS / SNS  | Number of receive loops; each loop processes callbacks inline, one message at a time. |
| Kafka      | Maximum callback concurrency across assigned partitions.                              |
| Kinesis    | Maximum callback concurrency across the kinsumer's owned shards.                      |
| Redis list | Number of callback workers processing popped messages.                                |
| File       | Number of callback workers processing file records.                                   |
| In-memory  | Number of callback workers processing buffered messages.                              |

Configured primary inputs default to one runner. Use the old consumer count as a starting point, not as an automatic behavior-preserving mapping. SQS and SNS could already have an input runner count; resolve any conflicting values rather than overwriting them mechanically.

**Primary and retry callbacks now run independently.** The old consumer runner pool limited their combined work. A primary input with four runners and an SQS retry input with one runner can now execute five callbacks at once. Review connection pools, rate limits, shared state, and callback thread safety against that combined concurrency.

## 3. Migrate SQS and SNS Inputs[​](#3-migrate-sqs-and-sns-inputs "Direct link to 3. Migrate SQS and SNS Inputs")

SNS delivery uses an SQS-backed input, so these settings and acknowledgement semantics apply to both transports.

| Input setting          | Default      | Meaning                                                                                  |
| ---------------------- | ------------ | ---------------------------------------------------------------------------------------- |
| `runner_count`         | `1`          | Receive loops and maximum inline callback concurrency.                                   |
| `grace_time`           | `10s`        | Delayed cancellation allowance for acknowledgement operations.                           |
| `acknowledgement_mode` | `individual` | Delete after each successful callback, or batch successful deletions per receive result. |

Move the old consumer `acknowledge_grace_time` value to the referenced SQS or SNS input's `grace_time` as the initial acknowledgement allowance, then remove the old consumer key.

Individual acknowledgement deletes each successful message immediately after its callback. With `acknowledgement_mode: batch`, a receive loop processes the entire receive result before deleting any of its successful messages. This reduces delete requests but extends visibility-timeout exposure for messages processed early in the result. It does not turn record callbacks into batch callbacks.

Input `grace_time` does not extend callback processing and is not a single global post-drain deadline. Each delete operation uses a delayed-cancellation context derived from the input's `Run` context. Calling `Stop` alone does not cancel that context or start an acknowledgement countdown. Account for AWS client request timeouts and retries when budgeting shutdown.

Keep visibility timeout large enough for time spent waiting within a receive result, processing, and deletion. Deletion failures can cause successful work to be delivered again; acknowledgement uses the configured AWS client retries, and the input generally logs failures and continues. Preserve idempotency.

### Old Configuration[​](#old-configuration "Direct link to Old Configuration")

```
stream:

  consumer:

    orders:

      input: order-events

      runner_count: 4

      consume_grace_time: 30s

      acknowledge_grace_time: 10s

  input:

    order-events:

      type: sqs

      queue_id: order-events

      runner_count: 1

      visibility_timeout: 120
```

### New Configuration[​](#new-configuration "Direct link to New Configuration")

```
stream:

  consumer:

    orders:

      input: order-events

      grace_time: 30s

  input:

    order-events:

      type: sqs

      queue_id: order-events

      runner_count: 4

      grace_time: 10s

      acknowledgement_mode: individual

      visibility_timeout: 120
```

This example deliberately chooses four inline receive workers. Confirm that this concurrency and visibility timeout fit your workload; the old input and consumer counts represented separate stages.

SQS and SNS inputs provide native redelivery for messages whose callback returns `ack=false`: those messages are not deleted. Preserve redrive policy and dead-letter queue configuration.

## 4. Migrate Kafka Inputs[​](#4-migrate-kafka-inputs "Direct link to 4. Migrate Kafka Inputs")

Kafka gains these input settings:

| Setting           | Default     | Meaning                                                                                  |
| ----------------- | ----------- | ---------------------------------------------------------------------------------------- |
| `runner_count`    | `1`         | Maximum concurrent callback work across assigned partitions.                             |
| `processing_mode` | `unordered` | Concurrent processing within a topic-partition, or sequential processing with `ordered`. |
| `grace_time`      | `10s`       | Commit allowance after processing finishes, including during shutdown.                   |
| `consume_delay`   | `0`         | Optional delay based on record timestamp; disabled by default.                           |

Set `processing_mode: ordered` when side effects must be sequential within each topic-partition. Unordered processing permits concurrent callbacks and out-of-order completion within a partition. Ordered mode is not global ordering across partitions, nor does it order work redelivered through a separate retry queue.

Kafka commits records handed to the processing callback even when the stream consumer returns `ack=false`. Transport commits are not a success-only acknowledgement mechanism. Enable `stream.consumer.<name>.retry.enabled` whenever ordinary failed callbacks must be retried; with retries disabled, those failures are committed without Gosoline redelivery.

Retry publication is not atomic with offset commit. Failed publication does not prevent the commit. Non-ignorable `GetModel` errors, nil models, and decoding errors return before retry publication. Define how such records are detected and recovered rather than treating `retry.enabled: true` as a guarantee for every failure.

### Configuration Example[​](#configuration-example "Direct link to Configuration Example")

```
stream:

  consumer:

    orders:

      input: order-events

      grace_time: 30s

      retry:

        enabled: true

        type: sqs

        runner_count: 1

  input:

    order-events:

      type: kafka

      connection: default

      topic_id: order-events

      group_id: orders

      runner_count: 4

      processing_mode: ordered

      grace_time: 10s
```

The example chooses per-partition ordering and separate SQS retries. Review whether retrying a failed record later is compatible with your ordering requirements. Preserve all other connection, topic, offset, and retry settings during migration.

Remove the old consumer `acknowledge_grace_time`; explicitly review input `grace_time` for commits rather than assuming the old setting controlled the same operation.

Do not add `consume_delay` just for migration. When used, it must be nonnegative and strictly less than `rebalance_timeout`. Record timestamps may come from the producer's clock under Kafka's `CreateTime` policy. Future timestamps are delayed by at most the configured delay; records without timestamps are not delayed.

Rebalance blocking now includes callback processing and commits. Even a valid `consume_delay` does not guarantee the rebalance budget is sufficient. Review `max_poll_records`, callback latency, concurrency, commit retries, and rebalance/session timeouts together.

## 5. Migrate Kinesis Inputs[​](#5-migrate-kinesis-inputs "Direct link to 5. Migrate Kinesis Inputs")

Kinesis adds input `runner_count` and `processing_mode`, defaulting to `1` and `unordered`. Choose `ordered` when sequential callbacks within each shard are required. Unordered mode permits concurrent processing within a shard; checkpoints advance only through the contiguous handled prefix.

Kinesis checkpoints records handed to the processing callback even when its result is `ack=false`. Enable consumer retries for ordinary callback failures that require redelivery, subject to the same decoding and retry-publication limitations as Kafka. Kinesis stream-envelope unmarshalling errors can also be logged and checkpointed without reaching the consumer callback.

Keep `release_delay`, which defaults to `5s`, as the Kinesis cleanup window. It starts after processing drains and covers final checkpoint persistence, shard release, and client deregistration. Do not replace it with consumer `grace_time`. Remove consumer `acknowledge_grace_time` and review `release_delay` separately.

Successful later records may be replayed if an earlier record leaves a gap during shutdown. Checkpoint failure can also cause redelivery. Preserve idempotency even with ordered processing.

Kinesis also supports optional `consume_delay`. Keep it unchanged unless delay behavior is intentionally required.

## 6. Migrate Redis, File, and In-Memory Inputs[​](#6-migrate-redis-file-and-in-memory-inputs "Direct link to 6. Migrate Redis, File, and In-Memory Inputs")

Set callback concurrency through each input's `runner_count`, defaulting to one worker, and remove the consumer runner count. Remove consumer `acknowledge_grace_time`; these inputs do not have an equivalent acknowledgement setting.

File input `blocking` is removed. A file input now finishes at end of file, which allows its consumer to complete and stop the retry input. Remove the key and verify that exhaustion is intended to finish the workload.

Check tests and local development wiring that previously read `Input.Data()` directly. They must now receive messages through the processing callback passed to `Run`.

## 7. Review Retry Configuration and Callback Results[​](#7-review-retry-configuration-and-callback-results "Direct link to 7. Review Retry Configuration and Callback Results")

Preserve existing `stream.consumer.<name>.retry` settings, including queue identifiers, delay, maximum attempts, and dead-letter behavior.

For SQS retry handlers, settings remain below `stream.consumer.<name>.retry`:

| Setting                | Default                                    | Purpose                                                       |
| ---------------------- | ------------------------------------------ | ------------------------------------------------------------- |
| `runner_count`         | `1`                                        | Independent retry-input callback concurrency.                 |
| `ack_grace_time`       | `10s`                                      | Delayed cancellation allowance for retry-message deletion.    |
| `acknowledgement_mode` | `individual`                               | Individual deletion or batching per receive result.           |
| `grace_time`           | `10s`, or configured `kernel.kill_timeout` | Delayed cancellation allowance for publishing retry messages. |

Do not confuse consumer processing `grace_time`, retry-publication `retry.grace_time`, and retry-message deletion `retry.ack_grace_time`.

The acknowledgement boolean remains significant in application callbacks:

| Callback result              | Consumer behavior                                                          |
| ---------------------------- | -------------------------------------------------------------------------- |
| `true, nil`                  | Successful processing.                                                     |
| `false, err` or `false, nil` | Native redelivery where available, otherwise retry publication if enabled. |
| `true, err`                  | Error is logged, but the consumer does not request retry.                  |

For Kafka and Kinesis, returning false does not hold the original record's offset or checkpoint back. For SQS retry messages, returning false leaves the retry message available for native redelivery.

## 8. Replace Removed Batch Consumers[​](#8-replace-removed-batch-consumers "Direct link to 8. Replace Removed Batch Consumers")

Batch callbacks, factories, runners, mocks, and `batch_size` configuration are removed. Discover `RunBatchConsumer`, `RunBatchConsumers`, `RunUntypedBatchConsumer`, `RunUntypedBatchConsumers`, `BatchConsumerCallback`, `UntypedBatchConsumerCallback`, their runnable variants, and batch consumer constructors/factories.

There is no automatic replacement. Choose one path for each consumer:

1. **Record-at-a-time processing:** use normal typed or untyped callbacks and the corresponding non-batch runner or module factory. Confirm the former batch operation is correct and economically acceptable for each record.
2. **Application-owned batching:** own buffering, flush timing, partial failures, shutdown flushing, idempotency, and how each record's callback result corresponds to the batch result.

Do not mechanically convert operations that rely on transactions, aggregate validation, rate limits, or per-item results. Returning success merely because a record entered an in-memory buffer can acknowledge it before durable processing.

Runnable callback `Run` contexts are cancelled only after both inputs have returned, including processing and transport cleanup. A batch flush needed to finish `Consume` must not wait for that cancellation: this can deadlock shutdown.

## 9. Review Aggregate Messages[​](#9-review-aggregate-messages "Direct link to 9. Review Aggregate Messages")

Aggregate envelopes are separate from the removed batch consumer APIs. Their acknowledgement behavior also changes:

| `aggregate_message_mode` | Old behavior                                              | New callback result for the envelope                                                      |
| ------------------------ | --------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| `atMostOnce` (default)   | Acknowledged before processing children.                  | True after processing if any child succeeds; false if all fail or the aggregate is empty. |
| `atLeastOnce`            | Acknowledged after iteration regardless of child results. | True only if every child succeeds; an empty aggregate returns true.                       |

On SQS/SNS, an `atLeastOnce` partial failure can redeliver successful children alongside failed ones. Review duplicate side effects and idempotency. On Kafka/Kinesis, envelope results do not prevent transport commit/checkpoint; review failed-child retry publication separately.

Keep the existing mode until its intended partial-failure behavior is understood. Test aggregate handling independently of any batch-consumer replacement.

## 10. Replace Kafka Read-Only Behavior[​](#10-replace-kafka-read-only-behavior "Direct link to 10. Replace Kafka Read-Only Behavior")

Kafka connection `is_read_only` is removed. Previously it suppressed Kafka output writes and permitted consumption without consumer-group commits. After migration, outputs produce normally and inputs use consumer groups with explicit commits.

For every former read-only connection, identify its purpose before deleting the key:

* **Suppress writes:** use explicit environment-specific output wiring or appropriate credentials; production wiring now writes normally.
* **Inspect or replay:** use a dedicated consumer group and explicit offset policy when committing consumption is acceptable.
* **Avoid consumer-group participation or commits:** select a separate inspection mechanism; the stream Kafka input no longer offers this mode.
* **Use read-only credentials:** verify that the selected consumption mode and its group/commit permissions match the principal.

Do not silently enable output writes or change group participation while removing obsolete configuration.

## 11. Review Shutdown and Observability[​](#11-review-shutdown-and-observability "Direct link to 11. Review Shutdown and Observability")

Review `kernel.kill_timeout` and the deployment termination budget together. Allow for outstanding receives or Redis pops, shared callback drain, retry publication, transport cleanup, and any other modules shutting down. SQS `Stop` does not interrupt an outstanding receive. These operations can overlap, so grace settings are not a universal formula for maximum shutdown duration.

Verify callback dependencies also respect cancellation. Successful `Run` termination does not prove that deletion, commit, or checkpoint persistence succeeded; some shutdown cleanup failures are logged without failing the run.

Update monitoring and metadata clients:

| Change                                                                                            | Migration action                                                                                      |
| ------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| Kafka metrics use consumer/topic granularity rather than partition dimensions.                    | Update queries and alerts using `Partition`.                                                          |
| Kafka `WaitDuration` and `CommitFailures` are no longer emitted.                                  | Replace dashboards and alerts relying on those names.                                                 |
| Kafka/Kinesis expose `SleepDuration` with consumption delay; it contributes to `ProcessDuration`. | Account for deliberate delay in latency and lag alerts.                                               |
| Kafka `ProcessDuration` also includes commit work.                                                | Review comparisons with previous callback timings.                                                    |
| Health tracks active processing independently.                                                    | A slow callback can make the consumer unhealthy while others succeed; review health-check thresholds. |
| Consumer metadata no longer includes `runner_count`.                                              | Update metadata consumers and assertions.                                                             |

Counters are not evidence of successful durable effects: stream `ProcessedCount` counts processing attempts, retry-put count counts publication attempts, and Kafka `RecordsConsumed` counts fetched records, including work skipped during shutdown. Correlate them with errors, queue depth, commits/checkpoints, and application outcomes.

## 12. Review Message and Integration Compatibility[​](#12-review-message-and-integration-compatibility "Direct link to 12. Review Message and Integration Compatibility")

Consumed Kafka keys are represented by the `KafkaKey` stream message attribute so they can be preserved when a stream message is written back to Kafka. Internal original-record metadata is removed. Review code or tests using `KafkaSourceMessage`, `MetaDataKafkaOriginalMessage`, or `NewKafkaMessageAttrs`.

Custom inputs and direct transport integrations require a separate API review:

| Integration              | Change                                                                                                                                                                                   |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `stream.Input`           | `Run` now takes `stream.InputProcess`, a callback receiving context and message and returning an acknowledgement boolean. `Data()` is removed; `Stop` and `IsHealthy` remain.            |
| `AcknowledgeableInput`   | Interface and consumer-driven `Ack` / `AckBatch` are removed. Inputs own acknowledgement and must wait for outstanding work before returning.                                            |
| Processing drain         | `exec.WithDrainContext` / `DrainContextFrom` carry the caller-owned drain signal. Inputs must preserve that shared processing cancellation contract.                                     |
| Direct Kafka consumer    | `Run` receives a per-record context-aware callback returning bool; handler/channel plumbing and read-only arguments are removed. Constructors and readers also change.                   |
| Direct Kinesis consumer  | `Run` receives a context-aware raw-record handler returning error; channel-handler and `Done` plumbing are removed. Ordinary handler failures are not a success-only checkpoint barrier. |
| Low-level consumer tests | Base-consumer APIs are removed; `NewUntypedConsumerWithInterfaces` receives dependencies directly, including a retry input.                                                              |

Read the exact constructor signatures in the target Gosoline version and regenerate affected mocks. Config defaults are applied through configuration unmarshalling, not automatically to Go settings literals: explicitly initialize runner counts, processing modes, and cleanup durations in direct integrations. In particular, zero Kinesis runners are not a supported configuration and can stall processing or skip handlers depending on the mode.

The application migration instructions below flag these integrations for separate review rather than prescribing a mechanical framework rewrite.

## Migration Checklist[​](#migration-checklist "Direct link to Migration Checklist")

* Map every consumer to its primary and retry inputs across application and deployment configuration.
* Replace consumer `consume_grace_time` with `grace_time`; remove consumer `runner_count` and `acknowledge_grace_time`.
* Choose primary-input and retry-input concurrency together; resolve conflicting or shared settings.
* Choose Kafka/Kinesis ordering explicitly and review retry behavior, decoding failures, and retry-publication failures.
* Configure transport cleanup separately: SQS/SNS acknowledgement allowance, Kafka commit allowance, or Kinesis `release_delay`.
* Review SQS/SNS deletion mode, visibility timeout, redrive policy, and idempotency.
* Resolve each removed batch consumer and review aggregate-envelope semantics independently.
* Replace former Kafka read-only behavior deliberately.
* Remove file `blocking` and verify completion at end of file.
* Review callback cancellation, runnable-callback flushing, kernel timeout, and deployment shutdown budget.
* Update tests, mocks, metadata clients, dashboards, and alerts affected by the lifecycle changes.
* Run application checks and verify transport resume, retries, and application progress after rollout.

## AI Agent Migration Instructions[​](#ai-agent-migration-instructions "Direct link to AI Agent Migration Instructions")

Copy the prepared instructions to give an AI agent an operational migration task. They include discovery, configuration mappings, behavior decisions, validation, and stop conditions.

Copy AI migration instructions
