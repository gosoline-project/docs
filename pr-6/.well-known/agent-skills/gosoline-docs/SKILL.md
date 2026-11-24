---
name: gosoline-docs
description: "Read the Gosoline Docs instead of answering from memory. Documentation and articles for the Gosoline Go application framework. Use when a task involves Gosoline Docs: how it works, how to set it up or configure it, or its API."
---

# Gosoline Docs

Documentation and articles for the Gosoline Go application framework.

This skill points you at the Gosoline Docs, at https://gosoline-project.github.io/docs/pr-6. When a task involves Gosoline Docs, read the relevant page before relying on what you remember: the docs are the source of truth, and they change.

## Reading the docs

- Any page as Markdown: add `.md` to its URL (the home page is https://gosoline-project.github.io/docs/pr-6/index.md). Every page link below already points at the Markdown.
- [llms.txt](https://gosoline-project.github.io/docs/pr-6/llms.txt): every page with a one-line summary.
- [llms-full.txt](https://gosoline-project.github.io/docs/pr-6/llms-full.txt): every page in one file, for when you need all of it.
- [JSON API](https://gosoline-project.github.io/docs/pr-6/api/docs/pages.json): the page index, with each page as JSON and Markdown.

## Pages

- [Overview](https://gosoline-project.github.io/docs/pr-6/index.md)

### Fundamentals

- [Fundamentals](https://gosoline-project.github.io/docs/pr-6/fundamentals.md): Learn the fundamental concepts and core principles of gosoline:
- [The Application](https://gosoline-project.github.io/docs/pr-6/fundamentals/application.md)
- [Naming Patterns](https://gosoline-project.github.io/docs/pr-6/fundamentals/naming-patterns.md)

### Getting started

- [Getting started](https://gosoline-project.github.io/docs/pr-6/getting-started.md): Here, you'll find tutorials to help you get started quickly with gosoline:
- [Create a consumer](https://gosoline-project.github.io/docs/pr-6/getting-started/create-a-consumer.md)
- [Create an application](https://gosoline-project.github.io/docs/pr-6/getting-started/create-an-application.md)

#### Testing

- [Testing](https://gosoline-project.github.io/docs/pr-6/getting-started/testing.md): Testing is an important responsibility for developing working code. With these tutorials, you'll write your first tests with gosoline:
- [Test your consumer](https://gosoline-project.github.io/docs/pr-6/getting-started/testing/test-your-consumer.md)

### How-to guides

- [How-to guides](https://gosoline-project.github.io/docs/pr-6/how-to.md): Here, you'll learn how to do common tasks with gosoline:
- [Load configurations](https://gosoline-project.github.io/docs/pr-6/how-to/load-configs.md)
- [Build a CLI tool](https://gosoline-project.github.io/docs/pr-6/how-to/use-cli.md)
- [Implement health checks](https://gosoline-project.github.io/docs/pr-6/how-to/write-health-checks.md)

#### HTTP Server

- [HTTP Server](https://gosoline-project.github.io/docs/pr-6/how-to/http-server.md): Learn how to build HTTP services with the gosoline httpserver package.
- [Build an HTTP service](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/build-an-http-service.md)
- [Stream with Server-Sent Events](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/stream-with-sse.md)
- [Serve a frontend](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/serve-a-frontend.md)
- [Configure your server](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/configure-your-server.md)
- [Add middleware](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/add-middleware.md)
- [Authenticate requests](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/authentication.md)
- [Concurrency and connection pressure](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/concurrency-and-connection-pressure.md)
- [Real-world example](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/real-world-example.md)
- [Chaos middleware](https://gosoline-project.github.io/docs/pr-6/how-to/http-server/chaos-middleware.md)

#### Logging

- [Use loggers](https://gosoline-project.github.io/docs/pr-6/how-to/logging/use-loggers.md)
- [Sampling & fingers-crossed](https://gosoline-project.github.io/docs/pr-6/how-to/logging/sampling-and-fingers-crossed.md)
- [Use context with logs](https://gosoline-project.github.io/docs/pr-6/how-to/logging/log-context.md)
- [Implement a log handler](https://gosoline-project.github.io/docs/pr-6/how-to/logging/implement-a-log-handler.md)

#### Relational databases

- [Working with Relational Databases](https://gosoline-project.github.io/docs/pr-6/how-to/databases-sql.md)
- [sqlc - SQL Client](https://gosoline-project.github.io/docs/pr-6/how-to/databases-sql/sqlc.md)
- [sqlr - SQL Repository](https://gosoline-project.github.io/docs/pr-6/how-to/databases-sql/sqlr.md)
- [sqlh - SQL HTTP Handlers](https://gosoline-project.github.io/docs/pr-6/how-to/databases-sql/sqlh.md)

#### Streaming applications

- [Streaming applications](https://gosoline-project.github.io/docs/pr-6/how-to/streaming-applications.md): Learn how to build streaming applications with gosoline inputs, outputs, producers, consumers, and producer daemons.
- [Understand inputs and outputs](https://gosoline-project.github.io/docs/pr-6/how-to/streaming-applications/understand-inputs-and-outputs.md)
- [Create a producer](https://gosoline-project.github.io/docs/pr-6/how-to/streaming-applications/create-a-producer.md)
- [Create a consumer](https://gosoline-project.github.io/docs/pr-6/how-to/streaming-applications/create-a-consumer.md)
- [Use the producer daemon](https://gosoline-project.github.io/docs/pr-6/how-to/streaming-applications/use-the-producer-daemon.md)

#### Kafka

- [Notes on general Kafka usage](https://gosoline-project.github.io/docs/pr-6/how-to/kafka/general.md)
- [How to use the Kafka Schema Registry](https://gosoline-project.github.io/docs/pr-6/how-to/kafka/use-schema-registry.md)

#### Sampling

- [Use sampling](https://gosoline-project.github.io/docs/pr-6/how-to/sampling/use-sampling.md)

### Reference

- [Reference](https://gosoline-project.github.io/docs/pr-6/reference.md): Here, you can look up implementation details about gosoline:
- [Naming patterns](https://gosoline-project.github.io/docs/pr-6/reference/naming-patterns.md)
- [Package cfg](https://gosoline-project.github.io/docs/pr-6/reference/package-cfg.md)
- [Package cloud/aws](https://gosoline-project.github.io/docs/pr-6/reference/package-cloud_aws.md)
- [Package dbx](https://gosoline-project.github.io/docs/pr-6/reference/package-dbx.md)
- [Gosoline Fixture Loader](https://gosoline-project.github.io/docs/pr-6/reference/package-fixtures.md)
- [Package httpserver](https://gosoline-project.github.io/docs/pr-6/reference/package-httpserver.md)
- [Package log](https://gosoline-project.github.io/docs/pr-6/reference/package-log.md)
- [Package test](https://gosoline-project.github.io/docs/pr-6/reference/package-test.md)

### Migrations

- [Migrations](https://gosoline-project.github.io/docs/pr-6/migrations.md): Guides for migrating between gosoline versions:
- [Migrating to Identity and Naming Patterns](https://gosoline-project.github.io/docs/pr-6/migrations/app-identity-and-naming-patterns.md)
- [Migrating to the Standalone HTTP Server Package](https://gosoline-project.github.io/docs/pr-6/migrations/httpserver-package.md)
- [Migrating an Existing Service to SQLH CRUD](https://gosoline-project.github.io/docs/pr-6/migrations/sqlh-crud.md)
- [Migrating to the Callback-Based Stream Consumer Lifecycle](https://gosoline-project.github.io/docs/pr-6/migrations/stream-consumer-lifecycle.md)

### Blog

- [Gosoline Blog](https://gosoline-project.github.io/docs/pr-6/blog.md): Articles and announcements from the gosoline project.
- [Clearer Stream Processing, Safer Shutdown](https://gosoline-project.github.io/docs/pr-6/blog/gosoline-v0-66-0-stream-processing.md): Stream consumers spend most of their time doing something straightforward: receive a message, run application logic, and record that the message was handled.
- [Introducing the Standalone HTTP Server Package](https://gosoline-project.github.io/docs/pr-6/blog/introducing-httpserver-package.md): The HTTP server has always been one of the most commonly used parts of gosoline.
- [gosoline goes modular: smaller packages, better APIs](https://gosoline-project.github.io/docs/pr-6/blog/gosoline-goes-modular.md): gosoline has grown into a mature application framework that powers production backend systems.
- [Identity & Naming Patterns: Flexible, Configuration-Driven Resource Naming](https://gosoline-project.github.io/docs/pr-6/blog/identity-and-naming-patterns.md): Managing cloud resources across multiple environments, teams, and services is hard enough without fighting your own naming conventions.
- [Log Sampling in Go: Less Noise, More Debuggability](https://gosoline-project.github.io/docs/pr-6/blog/sampling-fingers-crossed-logging.md): High-traffic services have a logging problem: the more successful traffic you handle, the more you pay to store and query logs that rarely matter.
