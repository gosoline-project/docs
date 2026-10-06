---
sidebar:
  order: 4
title: Stream Consumer Lifecycle
---

This guide covers the breaking stream changes introduced in Gosoline **v0.66.0** by commit [`04b54462`](https://github.com/justtrackio/gosoline/commit/04b54462dfb9cfa789577d3972d6098cd7d69860), **stream: unify input processing and shutdown**. Follow this guide when upgrading from an earlier release to v0.66.0 or later.

Previously, inputs delivered messages through `Input.Data()`, and the consumer controlled callback concurrency and acknowledgement. Inputs now invoke a processing callback directly and own acknowledgement, offset commits, or checkpoints. The consumer supplies a shared shutdown processing deadline for its primary and retry inputs.

Normal typed and untyped application callbacks retain their `Consume` contract. Most applications need configuration and behavior changes rather than a callback rewrite. Batch consumers, custom inputs, and direct Kafka or Kinesis integrations also have API changes.

## 1. Discover Consumers and Their Inputs

Before editing, map each `stream.consumer.<name>.input` to its `stream.input.<input-name>` configuration. Include retry handlers, model subscribers, custom input factories, tests, and every deployment configuration source: YAML, Helm values, Terraform templates, and environment-variable mappings.

Classify these findings:

| Finding | Migration action |
|---|---|
| Consumer `consume_grace_time` | Rename to consumer `grace_time`, then review the shared deadline. |
| Consumer `runner_count` | Move the desired concurrency to the referenced input, after reviewing existing input settings and retry concurrency. |
| Consumer `acknowledge_grace_time` | Remove; use transport-specific cleanup settings as described below. |
| Batch consumer APIs or `batch_size` | Choose record-at-a-time processing or application-owned batching. |
| Kafka connection `is_read_only` | Select replacement behavior before removing the setting. |
| File input `blocking` | Remove and review completion at end of file. |
| Custom `Input`, `Data`, `Ack`, or `AckBatch` usage | Review the integration API changes before migrating. |

An input configuration name is not necessarily a shared input instance. Normal consumer construction calls `NewConfigurableInput`; `ProvideConfigurableInput` caches instances, and in-memory wiring can also share them. If several consumers reference the same configuration, check both their desired settings and their actual instance wiring. Kafka and SQS inputs cannot simply be run twice on the same instance.

## 2. Migrate Consumer Settings

### Processing Grace Time

Replace `stream.consumer.<name>.consume_grace_time` with `stream.consumer.<name>.grace_time`. Use the old duration as the initial value, but review its new meaning.

The old setting delayed cancellation separately for each callback. The new setting, defaulting to `10s`, is one shared processing window for both primary and retry work. During shutdown, the consumer calls `Stop` on both inputs and then starts the drain timer. Already-fetched work may still be passed to callbacks. When the timer expires, callback contexts are cancelled.

This is cooperative cancellation, not a hard execution cutoff. A callback that ignores cancellation can keep shutdown waiting. A custom input whose `Stop` blocks can delay the timer itself.

### Concurrency

Remove `stream.consumer.<name>.runner_count`. Callback concurrency belongs to the input at `stream.input.<input-name>.runner_count`.

| Input | Meaning of input `runner_count` |
|---|---|
| SQS / SNS | Number of receive loops; each loop processes callbacks inline, one message at a time. |
| Kafka | Maximum callback concurrency across assigned partitions. |
| Kinesis | Maximum callback concurrency across the kinsumer's owned shards. |
| Redis list | Number of callback workers processing popped messages. |
| File | Number of callback workers processing file records. |
| In-memory | Number of callback workers processing buffered messages. |

Configured primary inputs default to one runner. Use the old consumer count as a starting point, not as an automatic behavior-preserving mapping. SQS and SNS could already have an input runner count; resolve any conflicting values rather than overwriting them mechanically.

**Primary and retry callbacks now run independently.** The old consumer runner pool limited their combined work. A primary input with four runners and an SQS retry input with one runner can now execute five callbacks at once. Review connection pools, rate limits, shared state, and callback thread safety against that combined concurrency.

## 3. Migrate SQS and SNS Inputs

SNS delivery uses an SQS-backed input, so these settings and acknowledgement semantics apply to both transports.

| Input setting | Default | Meaning |
|---|---|---|
| `runner_count` | `1` | Receive loops and maximum inline callback concurrency. |
| `grace_time` | `10s` | Delayed cancellation allowance for acknowledgement operations. |
| `acknowledgement_mode` | `individual` | Delete after each successful callback, or batch successful deletions per receive result. |

Move the old consumer `acknowledge_grace_time` value to the referenced SQS or SNS input's `grace_time` as the initial acknowledgement allowance, then remove the old consumer key.

Individual acknowledgement deletes each successful message immediately after its callback. With `acknowledgement_mode: batch`, a receive loop processes the entire receive result before deleting any of its successful messages. This reduces delete requests but extends visibility-timeout exposure for messages processed early in the result. It does not turn record callbacks into batch callbacks.

Input `grace_time` does not extend callback processing and is not a single global post-drain deadline. Each delete operation uses a delayed-cancellation context derived from the input's `Run` context. Calling `Stop` alone does not cancel that context or start an acknowledgement countdown. Account for AWS client request timeouts and retries when budgeting shutdown.

Keep visibility timeout large enough for time spent waiting within a receive result, processing, and deletion. Deletion failures can cause successful work to be delivered again; acknowledgement uses the configured AWS client retries, and the input generally logs failures and continues. Preserve idempotency.

### Old Configuration

```yaml
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

### New Configuration

```yaml
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

## 4. Migrate Kafka Inputs

Kafka gains these input settings:

| Setting | Default | Meaning |
|---|---|---|
| `runner_count` | `1` | Maximum concurrent callback work across assigned partitions. |
| `processing_mode` | `unordered` | Concurrent processing within a topic-partition, or sequential processing with `ordered`. |
| `grace_time` | `10s` | Commit allowance after processing finishes, including during shutdown. |
| `consume_delay` | `0` | Optional delay based on record timestamp; disabled by default. |

Set `processing_mode: ordered` when side effects must be sequential within each topic-partition. Unordered processing permits concurrent callbacks and out-of-order completion within a partition. Ordered mode is not global ordering across partitions, nor does it order work redelivered through a separate retry queue.

Kafka commits records handed to the processing callback even when the stream consumer returns `ack=false`. Transport commits are not a success-only acknowledgement mechanism. Enable `stream.consumer.<name>.retry.enabled` whenever ordinary failed callbacks must be retried; with retries disabled, those failures are committed without Gosoline redelivery.

Retry publication is not atomic with offset commit. Failed publication does not prevent the commit. Non-ignorable `GetModel` errors, nil models, and decoding errors return before retry publication. Define how such records are detected and recovered rather than treating `retry.enabled: true` as a guarantee for every failure.

### Configuration Example

```yaml
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

## 5. Migrate Kinesis Inputs

Kinesis adds input `runner_count` and `processing_mode`, defaulting to `1` and `unordered`. Choose `ordered` when sequential callbacks within each shard are required. Unordered mode permits concurrent processing within a shard; checkpoints advance only through the contiguous handled prefix.

Kinesis checkpoints records handed to the processing callback even when its result is `ack=false`. Enable consumer retries for ordinary callback failures that require redelivery, subject to the same decoding and retry-publication limitations as Kafka. Kinesis stream-envelope unmarshalling errors can also be logged and checkpointed without reaching the consumer callback.

Keep `release_delay`, which defaults to `5s`, as the Kinesis cleanup window. It starts after processing drains and covers final checkpoint persistence, shard release, and client deregistration. Do not replace it with consumer `grace_time`. Remove consumer `acknowledge_grace_time` and review `release_delay` separately.

Successful later records may be replayed if an earlier record leaves a gap during shutdown. Checkpoint failure can also cause redelivery. Preserve idempotency even with ordered processing.

Kinesis also supports optional `consume_delay`. Keep it unchanged unless delay behavior is intentionally required.

## 6. Migrate Redis, File, and In-Memory Inputs

Set callback concurrency through each input's `runner_count`, defaulting to one worker, and remove the consumer runner count. Remove consumer `acknowledge_grace_time`; these inputs do not have an equivalent acknowledgement setting.

File input `blocking` is removed. A file input now finishes at end of file, which allows its consumer to complete and stop the retry input. Remove the key and verify that exhaustion is intended to finish the workload.

Check tests and local development wiring that previously read `Input.Data()` directly. They must now receive messages through the processing callback passed to `Run`.

## 7. Review Retry Configuration and Callback Results

Preserve existing `stream.consumer.<name>.retry` settings, including queue identifiers, delay, maximum attempts, and dead-letter behavior.

For SQS retry handlers, settings remain below `stream.consumer.<name>.retry`:

| Setting | Default | Purpose |
|---|---|---|
| `runner_count` | `1` | Independent retry-input callback concurrency. |
| `ack_grace_time` | `10s` | Delayed cancellation allowance for retry-message deletion. |
| `acknowledgement_mode` | `individual` | Individual deletion or batching per receive result. |
| `grace_time` | `10s`, or configured `kernel.kill_timeout` | Delayed cancellation allowance for publishing retry messages. |

Do not confuse consumer processing `grace_time`, retry-publication `retry.grace_time`, and retry-message deletion `retry.ack_grace_time`.

The acknowledgement boolean remains significant in application callbacks:

| Callback result | Consumer behavior |
|---|---|
| `true, nil` | Successful processing. |
| `false, err` or `false, nil` | Native redelivery where available, otherwise retry publication if enabled. |
| `true, err` | Error is logged, but the consumer does not request retry. |

For Kafka and Kinesis, returning false does not hold the original record's offset or checkpoint back. For SQS retry messages, returning false leaves the retry message available for native redelivery.

## 8. Replace Removed Batch Consumers

Batch callbacks, factories, runners, mocks, and `batch_size` configuration are removed. Discover `RunBatchConsumer`, `RunBatchConsumers`, `RunUntypedBatchConsumer`, `RunUntypedBatchConsumers`, `BatchConsumerCallback`, `UntypedBatchConsumerCallback`, their runnable variants, and batch consumer constructors/factories.

There is no automatic replacement. Choose one path for each consumer:

1. **Record-at-a-time processing:** use normal typed or untyped callbacks and the corresponding non-batch runner or module factory. Confirm the former batch operation is correct and economically acceptable for each record.
2. **Application-owned batching:** own buffering, flush timing, partial failures, shutdown flushing, idempotency, and how each record's callback result corresponds to the batch result.

Do not mechanically convert operations that rely on transactions, aggregate validation, rate limits, or per-item results. Returning success merely because a record entered an in-memory buffer can acknowledge it before durable processing.

Runnable callback `Run` contexts are cancelled only after both inputs have returned, including processing and transport cleanup. A batch flush needed to finish `Consume` must not wait for that cancellation: this can deadlock shutdown.

## 9. Review Aggregate Messages

Aggregate envelopes are separate from the removed batch consumer APIs. Their acknowledgement behavior also changes:

| `aggregate_message_mode` | Old behavior | New callback result for the envelope |
|---|---|---|
| `atMostOnce` (default) | Acknowledged before processing children. | True after processing if any child succeeds; false if all fail or the aggregate is empty. |
| `atLeastOnce` | Acknowledged after iteration regardless of child results. | True only if every child succeeds; an empty aggregate returns true. |

On SQS/SNS, an `atLeastOnce` partial failure can redeliver successful children alongside failed ones. Review duplicate side effects and idempotency. On Kafka/Kinesis, envelope results do not prevent transport commit/checkpoint; review failed-child retry publication separately.

Keep the existing mode until its intended partial-failure behavior is understood. Test aggregate handling independently of any batch-consumer replacement.

## 10. Replace Kafka Read-Only Behavior

Kafka connection `is_read_only` is removed. Previously it suppressed Kafka output writes and permitted consumption without consumer-group commits. After migration, outputs produce normally and inputs use consumer groups with explicit commits.

For every former read-only connection, identify its purpose before deleting the key:

- **Suppress writes:** use explicit environment-specific output wiring or appropriate credentials; production wiring now writes normally.
- **Inspect or replay:** use a dedicated consumer group and explicit offset policy when committing consumption is acceptable.
- **Avoid consumer-group participation or commits:** select a separate inspection mechanism; the stream Kafka input no longer offers this mode.
- **Use read-only credentials:** verify that the selected consumption mode and its group/commit permissions match the principal.

Do not silently enable output writes or change group participation while removing obsolete configuration.

## 11. Review Shutdown and Observability

Review `kernel.kill_timeout` and the deployment termination budget together. Allow for outstanding receives or Redis pops, shared callback drain, retry publication, transport cleanup, and any other modules shutting down. SQS `Stop` does not interrupt an outstanding receive. These operations can overlap, so grace settings are not a universal formula for maximum shutdown duration.

Verify callback dependencies also respect cancellation. Successful `Run` termination does not prove that deletion, commit, or checkpoint persistence succeeded; some shutdown cleanup failures are logged without failing the run.

Update monitoring and metadata clients:

| Change | Migration action |
|---|---|
| Kafka metrics use consumer/topic granularity rather than partition dimensions. | Update queries and alerts using `Partition`. |
| Kafka `WaitDuration` and `CommitFailures` are no longer emitted. | Replace dashboards and alerts relying on those names. |
| Kafka/Kinesis expose `SleepDuration` with consumption delay; it contributes to `ProcessDuration`. | Account for deliberate delay in latency and lag alerts. |
| Kafka `ProcessDuration` also includes commit work. | Review comparisons with previous callback timings. |
| Health tracks active processing independently. | A slow callback can make the consumer unhealthy while others succeed; review health-check thresholds. |
| Consumer metadata no longer includes `runner_count`. | Update metadata consumers and assertions. |

Counters are not evidence of successful durable effects: stream `ProcessedCount` counts processing attempts, retry-put count counts publication attempts, and Kafka `RecordsConsumed` counts fetched records, including work skipped during shutdown. Correlate them with errors, queue depth, commits/checkpoints, and application outcomes.

## 12. Review Message and Integration Compatibility

Consumed Kafka keys are represented by the `KafkaKey` stream message attribute so they can be preserved when a stream message is written back to Kafka. Internal original-record metadata is removed. Review code or tests using `KafkaSourceMessage`, `MetaDataKafkaOriginalMessage`, or `NewKafkaMessageAttrs`.

Custom inputs and direct transport integrations require a separate API review:

| Integration | Change |
|---|---|
| `stream.Input` | `Run` now takes `stream.InputProcess`, a callback receiving context and message and returning an acknowledgement boolean. `Data()` is removed; `Stop` and `IsHealthy` remain. |
| `AcknowledgeableInput` | Interface and consumer-driven `Ack` / `AckBatch` are removed. Inputs own acknowledgement and must wait for outstanding work before returning. |
| Processing drain | `exec.WithDrainContext` / `DrainContextFrom` carry the caller-owned drain signal. Inputs must preserve that shared processing cancellation contract. |
| Direct Kafka consumer | `Run` receives a per-record context-aware callback returning bool; handler/channel plumbing and read-only arguments are removed. Constructors and readers also change. |
| Direct Kinesis consumer | `Run` receives a context-aware raw-record handler returning error; channel-handler and `Done` plumbing are removed. Ordinary handler failures are not a success-only checkpoint barrier. |
| Low-level consumer tests | Base-consumer APIs are removed; `NewUntypedConsumerWithInterfaces` receives dependencies directly, including a retry input. |

Read the exact constructor signatures in the target Gosoline version and regenerate affected mocks. Config defaults are applied through configuration unmarshalling, not automatically to Go settings literals: explicitly initialize runner counts, processing modes, and cleanup durations in direct integrations. In particular, zero Kinesis runners are not a supported configuration and can stall processing or skip handlers depending on the mode.

The application migration instructions below flag these integrations for separate review rather than prescribing a mechanical framework rewrite.

## Migration Checklist

- Map every consumer to its primary and retry inputs across application and deployment configuration.
- Replace consumer `consume_grace_time` with `grace_time`; remove consumer `runner_count` and `acknowledge_grace_time`.
- Choose primary-input and retry-input concurrency together; resolve conflicting or shared settings.
- Choose Kafka/Kinesis ordering explicitly and review retry behavior, decoding failures, and retry-publication failures.
- Configure transport cleanup separately: SQS/SNS acknowledgement allowance, Kafka commit allowance, or Kinesis `release_delay`.
- Review SQS/SNS deletion mode, visibility timeout, redrive policy, and idempotency.
- Resolve each removed batch consumer and review aggregate-envelope semantics independently.
- Replace former Kafka read-only behavior deliberately.
- Remove file `blocking` and verify completion at end of file.
- Review callback cancellation, runnable-callback flushing, kernel timeout, and deployment shutdown budget.
- Update tests, mocks, metadata clients, dashboards, and alerts affected by the lifecycle changes.
- Run application checks and verify transport resume, retries, and application progress after rollout.

## AI Agent Migration Instructions

Copy the prepared instructions to give an AI agent an operational migration task. They include discovery, configuration mappings, behavior decisions, validation, and stop conditions.

[Download the migration instructions](/docs/pr-6/downloads/stream-consumer-lifecycle-migration-agent-instructions.txt). Use the code block’s copy button to copy the full prompt.

```text title="AI migration instructions"
Migrate this application to the callback-based stream lifecycle introduced in Gosoline v0.66.0 by commit 04b54462dfb9cfa789577d3972d6098cd7d69860 (stream: unify input processing and shutdown). Use Gosoline v0.66.0 or later. Preserve intended ordering, retries, side effects, batching, acknowledgement, and deployment behavior.

Scope:

- Migrate application code, tests, and configuration, including checked-in deployment templates.
- Standard typed and untyped Consume callbacks keep their existing signatures.
- Custom inputs and direct Kafka/Kinesis integrations need a separate API review. Discover and report them before attempting a mechanical rewrite.
- Read repository instructions and inspect the current working tree before editing. Preserve unrelated work. Do not commit or deploy unless asked.

Discovery:

1. Find every stream.consumer.<name> and resolve its input to stream.input.<input-name>. Include model subscribers and configuration generated by postprocessors.
2. Find every retry handler and its settings under stream.consumer.<name>.retry. Record queue IDs, after/delay settings, maximum attempts, redrive policy, and dead-letter behavior.
3. Search YAML, Helm values, Terraform templates, environment-variable mappings, test fixtures, and Go settings literals for consume_grace_time, acknowledge_grace_time, consumer runner_count, batch_size, file blocking, and Kafka connection is_read_only.
4. Find removed batch runners RunBatchConsumer, RunBatchConsumers, RunUntypedBatchConsumer, and RunUntypedBatchConsumers, batch callback interfaces and runnable variants, and batch constructors/module factories.
5. Find custom stream.Input implementations, Input.Data calls, AcknowledgeableInput, Ack/AckBatch, base-consumer APIs, low-level consumer constructors, and generated input/transport mocks.
6. Find direct Kafka and Kinesis consumers, readers, handlers, channel adapters, and settings literals.
7. Find aggregate_message_mode, callbacks returning false or true with an error, non-ignorable GetModel errors, decoding failures, and retry-publication failure handling.
8. Find dashboards and alerts using Kafka Partition dimensions, WaitDuration, or CommitFailures, and metadata clients expecting consumer runner_count.
9. Find KafkaSourceMessage, MetaDataKafkaOriginalMessage, NewKafkaMessageAttrs, and code/tests relying on original Kafka-record metadata.
10. Inspect actual input-instance wiring. A shared config name does not always mean a shared instance: normal construction calls NewConfigurableInput; ProvideConfigurableInput caches, and in-memory wiring can share instances. Kafka/SQS instances cannot be run twice.

Classify findings before editing. Produce a per-consumer decision list covering primary/retry concurrency, ordering, failed-work recovery, aggregate behavior, transport cleanup, and shutdown budgets. Resolve ambiguous choices with the user.

Migration order:

1. Confirm the target Gosoline dependency is v0.66.0 or later and contains the lifecycle change.
2. Resolve removed batch consumers and Kafka read-only behavior.
3. Migrate consumer settings and primary-input concurrency.
4. Review independent retry concurrency and transport-specific ordering/acknowledgement.
5. Review aggregate semantics, callback cancellation, and runnable-callback flushing.
6. Update affected application tests, wiring, metadata consumers, and observability configuration.
7. Run repository checks and summarize remaining decisions.
8. If deployment is separately authorized, verify workload, transport, retry, and application progress afterward.

Consumer configuration rules:

- Rename stream.consumer.<name>.consume_grace_time to stream.consumer.<name>.grace_time. Use the old duration initially and review the new shared deadline; default is 10s.
- The processing timer starts after Stop has been called on retry and primary inputs. Already-fetched work may still reach callbacks. Both inputs share this processing deadline.
- Cancellation is cooperative. Callbacks that ignore it can keep shutdown waiting; custom Stop implementations that block can delay the timer.
- Remove stream.consumer.<name>.runner_count. Set the desired callback concurrency on stream.input.<input-name>.runner_count after reviewing existing values. Configured primary inputs default to one runner.
- Do not overwrite an existing SQS/SNS input runner_count blindly: old input and consumer counts controlled different stages.
- Primary and retry inputs now invoke callbacks independently. Old consumer runners limited their combined work; N primary runners plus one SQS retry runner may now mean N+1 callbacks. Review thread safety, pools, and rate limits.
- Remove stream.consumer.<name>.acknowledge_grace_time. Apply the relevant transport cleanup mapping below.

SQS/SNS rules:

- SNS delivery delegates to SQS; apply the same concurrency/deletion review.
- Input runner_count is the number of receive loops. Each loop processes callbacks inline, sequentially within its receive result.
- Use the old consumer acknowledge_grace_time as the starting value for the referenced SQS/SNS input grace_time; default is 10s.
- Input grace_time is a delayed-cancellation allowance for individual acknowledgement operations. It does not extend callback processing or define a single global post-drain cleanup deadline.
- Each delete uses the input Run context with delayed cancellation. Stop alone does not cancel that context or start the acknowledgement countdown. Budget AWS request timeouts/retries and outstanding receives separately.
- Default acknowledgement_mode is individual: delete after each successful callback.
- Choose batch only when deliberate batch deletion is required. It processes a full receive result before deleting any successful messages from that result; it does not batch application callbacks.
- Keep visibility timeout sufficient for waiting in a receive result, callback work, and deletion. Preserve redrive/dead-letter settings and idempotency; deletion failure can redeliver successful work.
- False callback results leave messages undeleted for native redelivery. Native retrying inputs obtain their retry wiring from the input.

Kafka rules:

- Input runner_count defaults to 1 and limits concurrent callback work across assigned partitions.
- processing_mode defaults to unordered; only unordered and ordered are valid. Use ordered for sequential side effects within a topic-partition. This is not cross-partition ordering or ordering through a separate retry queue.
- Input grace_time defaults to 10s and provides commit allowance after processing. Review it explicitly; do not assume the former consumer acknowledgement setting controlled the same operation.
- Kafka commits records handed to the callback even when its result is false. Enable consumer retries whenever ordinary callback failures must be redelivered; retries are otherwise disabled by default for non-native transports.
- Retry publication is not atomic with commit. Publication failure does not stop offset commit. Non-ignorable GetModel errors, nil models, and decoding errors return before retry publication. Define observability and recovery for these cases.
- Preserve consume_delay unless an intentional delay is required. Default is 0; configured values must be nonnegative and strictly below rebalance_timeout.
- Timestamps may be producer-assigned under CreateTime. Future timestamps are delayed at most consume_delay; missing timestamps are not delayed.
- Rebalance blocking covers processing and commits. Review max_poll_records, callback latency, concurrency, delay, commit retries, and rebalance/session budgets together.

Kinesis rules:

- Input runner_count defaults to 1; processing_mode defaults to unordered. Use ordered for sequential side effects within each shard.
- Unordered mode allows concurrent work within a shard; checkpointing advances only through the contiguous handled prefix. Later successful records can replay if earlier work leaves a hole during shutdown.
- Kinesis checkpoints records handed to callbacks even when their result is false. Enable consumer retries for ordinary failed callbacks, with the same decoding/publication limitations as Kafka.
- Stream-envelope unmarshalling failures can be logged and checkpointed without reaching the consumer callback.
- Preserve release_delay as the post-processing checkpoint/release/deregistration window; default is 5s. It is separate from consumer grace_time. Remove old consumer acknowledge_grace_time and review release_delay independently.
- Preserve consume_delay unless deliberately changing delay behavior. Keep idempotency for checkpoint failures and replay.

Redis/file/in-memory rules:

- Move desired callback concurrency to input runner_count; default is one worker.
- Remove consumer acknowledge_grace_time; no equivalent transport acknowledgement setting exists here.
- Remove file blocking. File input finishes at end of file and allows its consumer to finish and stop retry processing. Confirm this completion behavior is intended.
- Update direct Input.Data tests/wiring to use the callback supplied to Run.

Retry and callback rules:

- Preserve retry settings and intended failure handling rather than disabling retry to simplify migration.
- SQS retry settings remain under stream.consumer.<name>.retry: runner_count defaults to 1, ack_grace_time to 10s, acknowledgement_mode to individual.
- retry.grace_time is the delayed-cancellation allowance for retry publication. It defaults from configured kernel.kill_timeout, otherwise 10s. It is distinct from consumer processing grace and retry ack_grace_time.
- Consume returning true with an error logs the error but does not request retry. Returning false requests native redelivery or enabled consumer retry publication, depending on transport.
- Do not claim that retry.enabled guarantees redelivery of all failures. Inspect decode/model failures and failed publication separately.

Removed batch consumer rules:

- Remove batch callbacks, runnable batch variants, factories/runners/mocks, and batch_size only after choosing replacement behavior.
- Record-at-a-time replacement is acceptable only when the former batch operation is correct and economically acceptable for each record. Use the normal typed or untyped callback and corresponding non-batch runner/module factory.
- Application-owned batching must define buffering, flush timing, partial failures, shutdown flush, idempotency, and per-message retry/acknowledgement results.
- Do not return success just because a message entered an in-memory buffer when acknowledgement must follow durable processing.
- Runnable callback Run contexts are cancelled only after both inputs return, including processing/cleanup. A flush required to finish Consume must not wait for that cancellation.

Aggregate-message rules:

- Aggregate envelopes are separate from batch consumer APIs. Test them independently.
- atMostOnce remains the default, but now acknowledges after processing when any child succeeds. All-failed and empty aggregates return false; previously the envelope was acknowledged before processing.
- atLeastOnce now returns true only when every child succeeds; empty aggregates return true. Previously the envelope was acknowledged after iteration regardless of child results.
- SQS/SNS partial failure can replay successful children. Kafka/Kinesis envelope results do not hold back transport commit/checkpoint; inspect failed-child retry publication separately.
- Preserve the configured mode until partial-failure semantics and duplicate side effects are understood.

Kafka read-only removal:

- is_read_only is removed from Kafka connection settings. Outputs now produce normally; inputs use groups and explicit commits.
- Identify whether it suppressed non-production writes, enabled replay/inspection without commits, avoided group participation, or matched read-only credentials.
- Select explicit environment-specific output wiring/credentials for suppressed writes.
- Use a dedicated consumer group and explicit offset policy only when committing consumption is intended. Choose a separate mechanism if non-committing inspection is required.
- Do not simply delete the setting and silently enable writes or group participation.

Integration and compatibility review:

- Input.Run now takes InputProcess (context plus *Message, returning bool); Data is removed. Stop and IsHealthy remain.
- AcknowledgeableInput and consumer-driven Ack/AckBatch are removed. Custom inputs must own acknowledgement and wait for outstanding callback work before returning.
- exec.WithDrainContext / DrainContextFrom convey caller-owned processing cancellation; preserve that shared drain contract.
- Direct Kafka Run uses a context-aware per-record callback returning bool. Constructors/readers change; handler/channel and read-only plumbing is removed.
- Direct Kinesis Run uses a context-aware raw-record handler returning error. Handler/channel/Done plumbing is removed; ordinary handler failure is not a success-only checkpoint barrier.
- Base-consumer APIs are removed. NewUntypedConsumerWithInterfaces takes dependencies directly, including a retry input (normally noop if unused). Regenerate affected mocks from the new interfaces.
- Check target-version signatures rather than guessing argument replacements. Go settings literals do not automatically receive config defaults; initialize runner counts, processing modes, and cleanup durations explicitly. Zero Kinesis runners can stall processing or skip handlers depending on mode.
- Consumed Kafka keys use the KafkaKey stream attribute. Original-record metadata and related helpers are removed; update affected code/tests.

Shutdown and observability:

- Review kernel.kill_timeout and deployment termination grace together. Include receive/pop latency, shared processing drain, retry publication, transport cleanup, and other module shutdown. Operations may overlap; grace settings are not a universal maximum-duration formula.
- Verify callback dependencies respect context cancellation and runnable callback cleanup cannot deadlock processing.
- Kafka metrics use consumer/topic granularity instead of former per-partition dimensions. WaitDuration and CommitFailures are no longer emitted.
- Kafka/Kinesis SleepDuration contributes to ProcessDuration when delay is used; Kafka ProcessDuration also includes commit work.
- Consumer health tracks active processing independently: a slow callback can cause unhealthy status while other callbacks succeed. Review thresholds against expected worst-case duration.
- Consumer metadata no longer contains runner_count.
- ProcessedCount counts attempts; RetryPutCount counts publication attempts; Kafka RecordsConsumed counts fetched records, including work skipped during shutdown. None alone proves durable success. Successful Run completion also does not prove cleanup persistence.

Validation:

- Confirm all obsolete consumer keys, batch APIs/config, file blocking, and Kafka is_read_only have been addressed across every environment and fixture.
- Confirm each consumer has explicit primary/retry concurrency, ordering, recovery, aggregate, and cleanup decisions.
- Run repository-standard formatting/build/tests with required fixture/integration tags. Fix compile failures first; preserve behavior rather than weakening assertions.
- Exercise combined primary/retry work, ordered/unordered processing, callback failure and retry-publication failure, deletion modes, active-work shutdown, file exhaustion, and aggregate partial failure where applicable.
- Report changed files, chosen behavior, checks run and results, and unresolved decisions.

Post-deployment verification (only if deployment is authorized):

- Record rollout time in UTC. Compare a bounded 5-15 minute pre-rollout baseline with a post-rollout window; allow at least five minutes after the last pod starts for group stabilization.
- Check expected readiness, no new restarts, resource usage, and continuing logs/traces where instrumented.
- Check stable Kafka sync/assignment, Kinesis shard consumption, or SQS/SNS receive/deletion activity. Cancellation/rebalance noise is benign only when bounded to handoff and followed by recovery.
- Compare errors, duration, retry-get/put attempts, retry queue depth, source backlog, and application outcomes. Distinguish existing retry backlog from newly failed work.
- Verify subscribers separately. Idle polling does not establish end-to-end recovery without an expected update; prefer a known update consumed and persisted after rollout.
- Report exact UTC windows, workloads, health/restarts, transport resume, retry trends, and observed progress. State sparse traffic or missing telemetry limitations.

Stop conditions:

- Batch semantics are unclear or require unresolved transactions, buffering, partial results, or acknowledgement decisions.
- Referenced input settings conflict across consumers, or an actual shared input instance has unsupported lifecycle/concurrency requirements.
- Kafka/Kinesis failures require redelivery but retry policy or decode/publication recovery is unresolved.
- Ordering requirements conflict with independent retries or concurrent callbacks.
- Former Kafka read-only intent cannot be preserved with an agreed replacement.
- Aggregate partial-failure behavior or runnable-callback shutdown flushing cannot be determined from application evidence.
- Custom input/direct transport APIs require a framework-integration rewrite outside the agreed application scope.
- Required observability replacements or shutdown budgets remain unresolved.

Ask for the missing decision; do not silently change behavior to make the migration compile.
```
