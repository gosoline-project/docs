# Kinesis RecordWriter

`RecordWriter` in `github.com/justtrackio/gosoline/pkg/cloud/aws/kinesis` writes records to one configured Kinesis stream using `PutRecords`. Construct it with `NewRecordWriter(ctx, config, logger, settings)` and a `*RecordWriterSettings`; use `PutRecord` for one record or `PutRecords` for a batch. The writer resolves the full stream name and Kinesis client from configuration, splits requests at the Kinesis batch limit, and retries records rejected in a response.

See [AWS package configuration](/docs/pr-5/reference/package-cloud_aws/.md) for shared AWS client settings.

## Metrics[​](#metrics "Direct link to Metrics")

Names below are Gosoline-authored canonical names. The OTEL writer exports them with a `gosoline.` prefix; the table shows the units authored by the emitter. See the [metric schema](/docs/pr-5/reference/metric-schema/.md) for writer conversion details.

| Metric                            | Instrument | Authored unit | Dimensions                                                                                                                                                                                                                                      | Description                                                                                                                                                                                                                     |
| --------------------------------- | ---------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cloud.aws.kinesis.sent.messages` | Counter    | count         | `stream.name` (full stream name), `error.type={{default}}` for records accepted by Kinesis, and conditionally `error.type=<count> <ErrorCode> errors` for record-level failures (comma-separated reason groups when multiple error codes occur) | Accepted and rejected records reported by one successful `PutRecords` response attempt. The default series counts accepted records; each conditional error value counts rejected records for its reported Kinesis error reason. |
| `cloud.aws.kinesis.batch.records` | Histogram  | count         | `stream.name` (full stream name)                                                                                                                                                                                                                | Number of records in each `PutRecords` request attempt. Retries are separate observations and contain only records still rejected by the previous attempt.                                                                      |

### Measurement notes[​](#measurement-notes "Direct link to Measurement notes")

For each successful `PutRecords` response, `sent.messages` records the accepted records under `error.type={{default}}` and rejected records under the response's error-reason value. A record rejected on one attempt and accepted on a later retry is counted in both attempts, so series aggregated across retries are not counts of unique input records. `batch.records` likewise measures each request attempt, including retry subsets, rather than only the original batch size.

A top-level `PutRecords` API error returns before metric emission, so that failed API call produces neither a `sent.messages` nor a `batch.records` observation. `PutRecord` delegates to `PutRecords` as a one-record request.

## Related pages[​](#related-pages "Direct link to Related pages")

* [Kinesis shard reader](/docs/pr-5/reference/package-cloud_aws_kinesis-reader/.md)
* [Kinsumer](/docs/pr-5/reference/package-kinsumer/.md)
* [AWS package configuration](/docs/pr-5/reference/package-cloud_aws/.md)
* [Metric schema](/docs/pr-5/reference/metric-schema/.md)
