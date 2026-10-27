# Overview

Gosoline is a Golang-based application framework specialized for building microservices in the cloud. It provides tools for handling most of the common challenges like configuration, logging, structured code execution, handling http requests, asynchronous message processing, writing integration tests and much more.

## Primary use cases[​](#primary-use-cases "Direct link to Primary use cases")

There are three primary uses cases for gosoline:

HTTP Server

Build REST web services with HTTP handling, caching, OAuth, and much more.

[Get started](/docs/pr-10/category/http-server/.md)

Message Queues

Process asynchronous messages from Kafka, Redis, or any other queuing or streaming system.

[Get started](/docs/pr-10/getting-started/create-a-consumer/.md)

Kernel Module

Implement a kernel module with which you can do anything, using gosoline's logging, configuration, and other solutions.

[Get started](/docs/pr-10/getting-started/create-an-application/.md)

<br />

Want to skip the guided walkthroughs and get straight to the information? Check out our [API reference docs](/docs/pr-10/category/reference/.md).

## Running the Go examples[​](#running-the-go-examples "Direct link to Running the Go examples")

The executable examples share the repository's root Go module and require **Go 1.27.3 or newer**. Use the dependency versions in `go.mod` and `go.sum`; copying an older example's imports into another module does not carry these security fixes with it.

Run `go test -count=1 -tags=fixtures ./...` to compile all snippets and test the examples. The `sqlh/result_transformer` directory is a compile-only fragment, not a standalone program. The SQLR fixture integration example additionally needs Docker and the `integration fixtures` build tags. The HTTP build-service example can be run directly from `docs/how-to/http-server/src/build-service` with `go run .`.

The dependency update keeps the existing example APIs. AWS S3 uses the SDK's current request-checksum and response-validation defaults; custom S3-compatible endpoints must support them. The legacy PostgreSQL and Avro dependencies still have advisories without compatible released fixes. Gosoline v0.68.3 needs the OpenTelemetry log v0.20 API and the stable SDK/exporters v1.44 resource schema. The security fixes in log v0.21 and stable v1.45 cannot be selected until a compatible Gosoline release exists: v0.21 removes log APIs, while SDK v1.45 changes the default resource schema and prevents enabled telemetry from starting with this Gosoline version. These examples are therefore not a vulnerability-free dependency baseline.
