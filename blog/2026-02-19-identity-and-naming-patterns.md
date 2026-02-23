---
slug: identity-and-naming-patterns
title: "Identity & Naming Patterns: Flexible, Configuration-Driven Resource Naming"
authors: [jaka]
tags: [identity, naming, configuration, breaking-changes]
---

Managing cloud resources across multiple environments, teams, and services is hard enough without fighting your own naming conventions. Gosoline's old `AppId` model locked you into a fixed `project/family/group` hierarchy — useful in practice, but inflexible by design. This release replaces it with a tag-based `Identity` system and a unified naming pattern engine that gives you full control over how every resource is named, from a single config file.

{/* truncate */}

## What changed at a glance

- The fixed `AppId` fields (`project`, `family`, `group`) are replaced by a flexible `app.tags` map — define any keys your organization needs.
- A new **unified naming pattern engine** controls the names of SQS queues, SNS topics, DynamoDB tables, Kinesis streams, Kafka topics, Redis keys, CloudWatch namespaces, Prometheus prefixes, and tracing service names — all from config.
- `{app.namespace}` lets you define your naming hierarchy once and reference it everywhere.
- `application.Default()` has been slimmed down from ~22 options to 7, making every capability an explicit opt-in.
- `pkg/es/` (Elasticsearch) and `pkg/parquet/` have been removed.

---

## The new Identity model

### Before and after

The old flat top-level config keys are replaced with a nested `app:` block:

```yaml
# Before
env: production
app_project: myproject
app_family: platform
app_group: core
app_name: my-service

# After
app:
  env: production
  name: my-service
  namespace: "{app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}"
  tags:
    project: myproject
    family: platform
    group: core
```

If you set your config via environment variables, the keys change too:

| Old variable | New variable |
| :--- | :--- |
| `APP_PROJECT` | `APP_TAGS_PROJECT` |
| `APP_FAMILY` | `APP_TAGS_FAMILY` |
| `APP_GROUP` | `APP_TAGS_GROUP` |
| `APP_NAME` | `APP_NAME` (unchanged) |
| `ENV` | `APP_ENV` |

### Dynamic tags

Tags are no longer limited to `project/family/group`. You can define any keys that make sense for your organization:

```yaml
app:
  tags:
    project: myproject
    team: backend
    region: eu-west-1
    cost_center: engineering
```

Tags are **only required if your naming patterns reference them** — a minimal setup with `{app.env}-{queueId}` needs no tags at all.

---

## Naming patterns: define once, use everywhere

Every resource gosoline manages has a configurable naming pattern. Patterns are plain strings with `{placeholder}` macros resolved from your Identity at startup.

### The namespace pattern

`app.namespace` is the cornerstone of the new naming system. Define your hierarchy once, reference it everywhere:

```yaml
app:
  namespace: "{app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}"
```

When `{app.namespace}` appears in a resource pattern, the dots are replaced by that service's delimiter — `-` for most AWS services, `/` for CloudWatch, `_` for Prometheus. This means the same namespace definition produces correctly formatted names for every service automatically.

### Concrete example: SQS queues

```yaml
app:
  name: order-service
  env: production
  namespace: "{app.tags.project}.{app.env}.{app.tags.group}"
  tags:
    project: logistics
    group: platform

cloud:
  aws:
    sqs:
      clients:
        default:
          naming:
            queue_pattern: "{app.namespace}-{queueId}"
            queue_delimiter: "-"
```

| Queue ID in code | Resolved queue name |
| :--- | :--- |
| `orders` | `logistics-production-platform-orders` |
| `shipments` | `logistics-production-platform-shipments` |

Switching to `env: dev` automatically updates all queue names — no queue-specific config changes needed.

### Strict placeholder validation

Unknown placeholders in naming patterns now return an error at startup. A typo like `{app.tag.project}` (missing `s`) or a leftover legacy `{project}` is caught immediately rather than silently producing wrong resource names in production.

### Full pattern reference

For a complete list of all configurable patterns across every service, see the [Naming Patterns reference](/reference/naming-patterns) and [Naming Patterns fundamentals](/fundamentals/naming-patterns).

---

## `application.Default()` slimmed down

Previously, `application.Default()` bundled ~22 options — health checks, metrics, tracing, profiling, Sentry, task runner, producer daemon, and more — whether your application needed them or not. A simple CLI tool or a lightweight worker ended up opting out of most of them, or unknowingly carrying capabilities it didn't use.

`application.Default()` now ships only the essentials: config loading infrastructure and logger wiring. Everything else — health checks, metrics, tracing, profiling, stream infrastructure — is an explicit opt-in. You compose exactly the application you need:

```go
application.Run(
    // load config from file
    application.WithConfigFile("./config.dist.yml", "yml"),

    // only the capabilities this service actually uses
    application.WithHttpHealthCheck,
    application.WithMetrics,
    application.WithTracing,

    application.WithModuleFactory("my-module", NewMyModule),
)
```

A service that only needs config loading and logging needs nothing beyond the new defaults. A full-featured service adds exactly the options it requires — no hidden defaults to opt out of.

This pairs naturally with the new naming system: since tags are optional and `app.env` defaults to `dev` and `app.name` defaults to `gosoline`, a minimal application needs zero configuration to start. Getting something running no longer requires wiring up a full config file first.

---

## Other breaking changes

### Go API

- `cfg.AppId` → `cfg.Identity`; `cfg.GetAppIdFromConfig()` → `cfg.GetAppIdentity()`
- `WithLoggerGroupTag` → `WithLoggerApplicationTag("group")`; `WithLoggerApplicationTag` → `WithLoggerApplicationName`

### Config key renames

Pattern keys have been renamed from the generic `naming.pattern` to descriptive, service-specific names: `naming.queue_pattern`, `naming.topic_pattern`, `naming.stream_pattern`, `naming.table_pattern`, etc. Each has an accompanying `naming.*_delimiter` setting.

### ModelId refactoring

`mdl.ModelId` no longer has explicit hierarchy fields (`.Project`, `.Family`, `.Group`). Like `cfg.Identity`, it now uses a `.Tags` map:

```go
// Before
modelId := mdl.ModelId{
    Project: "my-project",
    Family:  "my-family",
    Name:    "my-model",
}

// After
modelId := mdl.ModelId{
    Name: "my-model",
    Tags: map[string]string{
        "project": "my-project",
        "family":  "my-family",
    },
}
```

The string representation of a `ModelId` — used for message routing attributes and `mdlsub` publishers — is now configurable via `app.model_id.domain_pattern`. The model name is always appended automatically as the last dot-separated segment:

```yaml
app:
  model_id:
    domain_pattern: "{app.tags.project}.{app.env}"
# Result: myproject.production.myModel
```

A few things to be aware of:

- The `{modelId}` placeholder is **not** used in this pattern — the model name is appended automatically.
- If `domain_pattern` is not configured, calling `modelId.String()` will return an error.
- When parsing a canonical model ID string back, each placeholder matches non-dot characters, and the model name is everything after the final dot. This means using non-dot delimiters (e.g. `{app.tags.project}-{app.env}`) is valid, but the last `.` in the full string always marks the boundary before the model name.

See the [Naming Patterns reference](/reference/naming-patterns#modelid-domain-pattern-canonical-ids) for the full rules and examples.

### Removed packages

- **`pkg/es/`** — Elasticsearch client integration
- **`pkg/parquet/`** — Parquet file read/write support

---

## Migration quick start

For full step-by-step instructions covering stream input/output config, `mdlsub` publishers, Go code changes, and per-service pattern updates, see the [migration guide](/migrations/app-identity-and-naming-patterns).
