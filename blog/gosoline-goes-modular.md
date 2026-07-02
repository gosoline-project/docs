# gosoline goes modular: smaller packages, better APIs

June 11, 2026 ·

<!-- -->

4 min read

[![Jan Kamieth](https://avatars.githubusercontent.com/u/783502?s=400\&v=4)](https://github.com/j4k4)

[Jan Kamieth](https://github.com/j4k4)

gosoline has grown into a mature application framework that powers production backend systems. It provides the building blocks teams need to run services reliably: application lifecycle management, configuration handling, logging, metrics, tracing, and integrations for many common infrastructure components.

<!-- -->

That broad scope has served us well. It helped teams build services quickly without having to assemble every piece from scratch.

But it also comes with a cost.

Over time, gosoline became one large library that covers a wide range of use cases. Not every application needs all of them. A small worker might not need an HTTP server. A service might not use SQL. Another application might not touch most AWS integrations at all.

Still, depending on the main gosoline module can pull in a large dependency graph, which affects local development, CI build times, and dependency management.

At the same time, some of the earliest gosoline packages are showing their age. They were built when the project and its usage patterns were still evolving. Today, we have a much clearer understanding of how those APIs should look and how they can be easier to use.

That is why we are starting to split gosoline into smaller, focused packages.

This is a long-running effort, not a big-bang migration. The transition will take multiple months, and we do not expect it to be completed this year. It may not even be completed next year.

## A New Home for Focused Packages[​](#a-new-home-for-focused-packages "Direct link to A New Home for Focused Packages")

The new packages live under the gosoline project organization on GitHub:

<https://github.com/gosoline-project>

The documentation has also moved and is now available here:

<https://gosoline-project.github.io/docs/>

This gives each package more room to evolve independently. Packages can have cleaner APIs, smaller dependency sets, and documentation focused on the specific problem they solve.

The first extracted packages focus on HTTP services and SQL support:

* [`httpserver`](https://github.com/gosoline-project/httpserver): for building HTTP services
* [`sqlc`](https://github.com/gosoline-project/sqlc): low-level SQL client support, replacing the old `db` package
* [`sqlr`](https://github.com/gosoline-project/sqlr): SQL repositories, replacing the old `db-repo` package
* [`sqlh`](https://github.com/gosoline-project/sqlh): SQL HTTP server integration for building CRUD HTTP APIs

You can find the HTTP server documentation here:

<https://gosoline-project.github.io/docs/how-to/http-server/build-an-http-service>

The SQL package documentation is available here:

<https://gosoline-project.github.io/docs/how-to/databases-sql/>

## Why This Change Matters[​](#why-this-change-matters "Direct link to Why This Change Matters")

The goal is not just to move code around. The goal is to make gosoline easier to adopt, easier to build, and easier to maintain.

Smaller packages mean applications only depend on what they actually use. That helps reduce unnecessary dependencies and keeps builds leaner.

It also gives us the opportunity to improve APIs where the old packages no longer match how we build services today. Some of these new packages are intentionally not backward compatible because they are designed as a clean step forward, not as a thin wrapper around the old APIs.

## What Happens to the Existing Packages?[​](#what-happens-to-the-existing-packages "Direct link to What Happens to the Existing Packages?")

The existing packages remain in the main gosoline library for now. We know that existing applications rely on them, and we do not want to force migrations immediately.

However, usage of the old packages is deprecated. Over time, they will be removed from the main library as applications migrate to the new packages.

This gives teams a clear path forward while keeping existing services stable. There is no need to migrate everything at once. Existing applications can move package by package as the new modules become available and as teams have time to adopt them.

## The Long-Term Direction[​](#the-long-term-direction "Direct link to The Long-Term Direction")

Long term, the main gosoline library should focus on the core pieces needed to build applications: the kernel, configuration, logging, observability, and other framework-level primitives that are not tied to one specific use case.

Use-case-specific functionality, such as HTTP servers, SQL support, or cloud service integrations, can then live in dedicated packages with their own lifecycle and documentation.

This is an important step for gosoline. It keeps the framework useful for large production systems while making it lighter, clearer, and easier to evolve.

**Tags:**

* [gosoline](/docs/blog/tags/gosoline)
* [modularization](/docs/blog/tags/modularization)
* [framework](/docs/blog/tags/framework)

[Edit this page](https://github.com/gosoline-project/docs/tree/main/blog/2026-06-11-gosoline-goes-modular.md)
