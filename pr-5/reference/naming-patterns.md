# Naming patterns

Gosoline uses naming patterns to generate consistent names for resources (SQS queues, SNS topics, DynamoDB tables, Kafka topics, Redis keys, metric namespaces, tracing service names, ...).

Naming patterns are plain strings with placeholders (macros). They are configured via config keys (e.g. in `config.dist.yml`) and can be overridden via environment variables.

## How env vars map to config keys[​](#how-env-vars-map-to-config-keys "Direct link to How env vars map to config keys")

Environment variable keys are derived from config keys:

* Uppercase
* Replace `.` and `-` with `_` (via `cfg.DefaultEnvKeyReplacer`)
* Optionally prepend an env var prefix (e.g. `GOSO_`) if you configure one

Example (no prefix):

```
export CLOUD_AWS_SQS_CLIENTS_DEFAULT_NAMING_QUEUE_PATTERN="{app.namespace}-{queueId}"
```

Example (with prefix):

```
export GOSO_CLOUD_AWS_SQS_CLIENTS_DEFAULT_NAMING_QUEUE_PATTERN="{app.namespace}-{queueId}"
```

To use a prefix in code, configure it when building your app config:

```
// application option

application.WithConfigEnvKeyPrefix("goso")



// or cfg option

cfg.WithEnvKeyPrefix("goso")
```

## Global placeholders[​](#global-placeholders "Direct link to Global placeholders")

These placeholders come from `cfg.Identity` and work in all `Identity.Format()` based patterns:

| Placeholder        | Config source    | Meaning                                         |
| ------------------ | ---------------- | ----------------------------------------------- |
| `{app.env}`        | `app.env`        | Environment                                     |
| `{app.name}`       | `app.name`       | Application name                                |
| `{app.tags.<key>}` | `app.tags.<key>` | Any tag value (dynamic)                         |
| `{app.namespace}`  | `app.namespace`  | A reusable namespace pattern (see next section) |

### `app.namespace`[​](#appnamespace "Direct link to appnamespace")

`app.namespace` itself is a pattern, typically defined with dot-separated placeholders.

```
app:

  namespace: "{app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}"
```

When `{app.namespace}` is used in a naming pattern, the service-specific delimiter replaces the dots.

## ModelId domain pattern (canonical IDs)[​](#modelid-domain-pattern-canonical-ids "Direct link to ModelId domain pattern (canonical IDs)")

The canonical `mdl.ModelId` string form is configured via:

* **Config key:** `app.model_id.domain_pattern`

Rules:

* Only `{app.env}`, `{app.name}`, and `{app.tags.<key>}` placeholders are allowed
* Patterns may freely mix placeholders with static text and use any delimiter between placeholders
* The model name is appended automatically as the last dot-separated segment

info

When parsing a canonical model ID string back into a `ModelId`, each placeholder matches non-dot characters (`[^.]+`), and the model name is everything after the final dot. This means dots in the pattern separate the final model name segment from the domain.

Examples:

```
app:

  model_id:

    # dot-separated placeholders

    domain_pattern: "{app.tags.project}.{app.env}"

    # Result: myProject.production.myModel



    # mixed delimiters

    domain_pattern: "{app.tags.project}-{app.env}"

    # Result: myProject-production.myModel



    # static text prefix

    domain_pattern: "prefix-{app.env}"

    # Result: prefix-production.myModel
```

## Available naming patterns[​](#available-naming-patterns "Direct link to Available naming patterns")

### App identity[​](#app-identity "Direct link to App identity")

* `app.env` (string)
* `app.name` (string)
* `app.tags.<key>` (map)
* `app.namespace` (pattern string)

`app.namespace` has no delimiter of its own; it is expanded using the delimiter of the pattern that references it.

### AWS SQS queues[​](#aws-sqs-queues "Direct link to AWS SQS queues")

* **Config key:** `cloud.aws.sqs.clients.<client_name>.naming.queue_pattern`
* **Default:** `{app.namespace}-{queueId}`
* **Delimiter config key:** `cloud.aws.sqs.clients.<client_name>.naming.queue_delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{queueId}`

Env var (example for `default` client, with prefix `GOSO_`):

```
export GOSO_CLOUD_AWS_SQS_CLIENTS_DEFAULT_NAMING_QUEUE_PATTERN="{app.namespace}-{queueId}"

export GOSO_CLOUD_AWS_SQS_CLIENTS_DEFAULT_NAMING_QUEUE_DELIMITER="-"
```

### AWS SNS topics[​](#aws-sns-topics "Direct link to AWS SNS topics")

* **Config key:** `cloud.aws.sns.clients.<client_name>.naming.topic_pattern`
* **Default:** `{app.namespace}-{topicId}`
* **Delimiter config key:** `cloud.aws.sns.clients.<client_name>.naming.topic_delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{topicId}`

### AWS Kinesis[​](#aws-kinesis "Direct link to AWS Kinesis")

Kinesis has three independent patterns:

**Streams**

* **Config key:** `cloud.aws.kinesis.clients.<client_name>.naming.stream_pattern`
* **Default:** `{app.namespace}-{streamName}`
* **Delimiter config key:** `cloud.aws.kinesis.clients.<client_name>.naming.stream_delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{streamName}`

**Kinsumer metadata table (DynamoDB)**

* **Config key:** `cloud.aws.kinesis.clients.<client_name>.naming.metadata_table_pattern`
* **Default:** `{app.namespace}-kinsumer-metadata`
* **Delimiter config key:** `cloud.aws.kinesis.clients.<client_name>.naming.metadata_table_delimiter`
* **Default delimiter:** `-`

**Kinsumer metadata namespace (record prefix inside the metadata table)**

* **Config key:** `cloud.aws.kinesis.clients.<client_name>.naming.metadata_namespace_pattern`
* **Default:** `{app.namespace}-{app.name}`
* **Delimiter config key:** `cloud.aws.kinesis.clients.<client_name>.naming.metadata_namespace_delimiter`
* **Default delimiter:** `-`

### AWS S3 buckets[​](#aws-s3-buckets "Direct link to AWS S3 buckets")

* **Config key:** `cloud.aws.s3.clients.<client_name>.naming.bucket_pattern`
* **Default:** `{app.namespace}`
* **Delimiter config key:** `cloud.aws.s3.clients.<client_name>.naming.delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{bucketId}`

### DynamoDB tables (`pkg/ddb`)[​](#dynamodb-tables-pkgddb "Direct link to dynamodb-tables-pkgddb")

* **Config key:** `cloud.aws.dynamodb.clients.<client_name>.naming.table_pattern`
* **Default:** `{app.namespace}-{name}`
* **Delimiter config key:** `cloud.aws.dynamodb.clients.<client_name>.naming.table_delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{name}` (from `mdl.ModelId.Name`)

### Kafka[​](#kafka "Direct link to Kafka")

**Topics**

* **Config key:** `kafka.naming.topic_pattern`
* **Default:** `{app.namespace}-{topicId}`
* **Delimiter config key:** `kafka.naming.topic_delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{topicId}`

**Consumer groups**

* **Config key:** `kafka.naming.group_pattern`
* **Default:** `{app.namespace}-{app.name}-{groupId}`
* **Delimiter config key:** `kafka.naming.group_delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{groupId}`

### Redis client naming (`pkg/redis`)[​](#redis-client-naming-pkgredis "Direct link to redis-client-naming-pkgredis")

**Address pattern** (used when `redis.<name>.address` is empty)

* **Config key:** `redis.<client_name>.naming.address_pattern`
* **Default:** `{name}.{app.tags.group}.redis.{app.env}.{app.tags.family}`
* **Delimiter config key:** `redis.<client_name>.naming.address_delimiter`
* **Default delimiter:** `.`
* **Extra placeholders:** `{name}` (the redis client name)

**Key pattern** (prefixing keys)

* **Config key:** `redis.<client_name>.naming.key_pattern`
* **Default:** `{key}`
* **Delimiter config key:** `redis.<client_name>.naming.key_delimiter`
* **Default delimiter:** `-`
* **Extra placeholders:** `{key}`

### KvStore Redis key pattern (`pkg/kvstore`)[​](#kvstore-redis-key-pattern-pkgkvstore "Direct link to kvstore-redis-key-pattern-pkgkvstore")

KvStore expands `{store}` first and then passes the pattern to the underlying Redis naming.

* **Config key (store-specific):** `kvstore.<name>.redis.key_pattern`
* **Config key (global default):** `kvstore.default.redis.key_pattern`
* **Default:** `{app.namespace}-kvstore-{store}-{key}`
* **Extra placeholders:** `{store}`, `{key}`

### Metrics[​](#metrics "Direct link to Metrics")

Metric emission has its own V2 schema and writer naming rules; it is separate from the resource naming patterns below. See the [Metric schema reference](/docs/pr-5/reference/metric-schema/.md) for the current inventory, units, dimensions, help text, and backend-rendered names.

**CloudWatch namespace**

* **Config key:** `metric.writer_settings.cloudwatch.naming.namespace_pattern`
* **Default:** `{app.namespace}-{app.name}`
* **Delimiter config key:** `metric.writer_settings.cloudwatch.naming.namespace_delimiter`
* **Default delimiter:** `/`

This setting remains supported and formats the CloudWatch service namespace; it does not format each metric name.

**Prometheus metric prefix**

The Prometheus writer uses the fixed `gosoline` prefix. The application identity belongs on the scrape target through service discovery/target labels, not in metric names. The former `metric.writer_settings.prometheus.naming.namespace_pattern` and `metric.writer_settings.prometheus.naming.namespace_delimiter` settings no longer control names.

:::info Historical note Earlier editions of this reference described a configurable Prometheus prefix and a metric-calculator leader-election table. Those are not current metric configuration: the prefix is now fixed, and the calculator configuration was removed. The Kinsumer autoscale leader-election table below remains an active resource-naming pattern. :::

### Tracing[​](#tracing "Direct link to Tracing")

**Tracing service name**

* **Config key:** `tracing.naming.pattern`
* **Default:** `{app.namespace}-{app.name}`
* **Delimiter config key:** `tracing.naming.delimiter`
* **Default delimiter:** `-`

**AWS X-Ray daemon SRV lookup name** (used when `tracing.xray.addr_type: srv` and `tracing.xray.add_value` is empty)

* **Config key:** `tracing.xray.srv_naming.pattern`
* **Default:** `xray.{app.namespace}`
* **Delimiter config key:** `tracing.xray.srv_naming.delimiter`
* **Default delimiter:** `.`

### DDB leader election table (`pkg/conc/ddb`)[​](#ddb-leader-election-table-pkgconcddb "Direct link to ddb-leader-election-table-pkgconcddb")

* **Config key:** `conc.leader_election.<name>.naming.table_pattern`
* **Default:** `{app.tags.project}-{app.env}-{app.tags.family}-leader-elections`
* **Delimiter config key:** `conc.leader_election.<name>.naming.table_delimiter`
* **Default delimiter:** `-`

### Kinsumer autoscale leader election table[​](#kinsumer-autoscale-leader-election-table "Direct link to Kinsumer autoscale leader election table")

Kinsumer autoscale provides a table pattern which is used to configure the underlying DDB leader election module.

* **Config key:** `kinsumer.autoscale.dynamodb.naming.pattern`
* **Default:** `{app.env}-kinsumer-autoscale-leaders`

If you need to adjust the delimiter, configure the leader election settings (e.g. `conc.leader_election.<name>.naming.table_delimiter`).
