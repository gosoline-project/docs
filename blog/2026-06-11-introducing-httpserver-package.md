---
slug: introducing-httpserver-package
title: "Introducing the Standalone HTTP Server Package"
authors: [jaka]
tags: [gosoline, httpserver, framework, migration]
---

The HTTP server has always been one of the most commonly used parts of gosoline. Most services need a way to expose APIs, register routes, bind requests, return structured responses, and plug into the application lifecycle.

{/* truncate */}

As part of the ongoing modularization of gosoline, the HTTP server is now available as a standalone package:

https://github.com/gosoline-project/httpserver

The new package keeps the familiar Gin-based server lifecycle and middleware model, but the application-facing API has been cleaned up. Instead of building a definition tree and returning it to the framework, applications now receive a router and register routes on it directly.

That sounds like a small change, but it removes a lot of boilerplate and makes route registration easier to read.

## What Improved

The standalone package is not just a copy of the old `github.com/justtrackio/gosoline/pkg/httpserver` package in a new repository. It is a chance to improve the API around the way we build HTTP services today.

The most important changes are:

- Less boilerplate when creating routers and registering routes
- One handler service can handle multiple related routes
- Request binding is more direct and easier to follow
- Handler signatures are typed instead of relying on `any` values inside a generic request wrapper

The result is code that is shorter, clearer, and easier to migrate route by route.

## Router Setup Before and After

The old package commonly used a definer function that allocated and returned `*httpserver.Definitions`. Applications also often ended up creating one handler per route, even when those routes belonged to the same area of the API.

```go
func main() {
    application.RunHttpDefaultServer(DefineRouter)
}

func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger) (*httpserver.Definitions, error) {
    definitions := &httpserver.Definitions{}
    api := definitions.Group("/api/reports")

    overviewHandler, err := NewReportOverviewHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create report overview handler: %w", err)
    }

    runHandler, err := NewRunReportHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create run report handler: %w", err)
    }

    exportHandler, err := NewExportReportHandler(ctx, config, logger)
    if err != nil {
        return nil, fmt.Errorf("can not create export report handler: %w", err)
    }

    api.Use(authMiddleware)
    api.GET("/overview", httpserver.CreateHandler(overviewHandler))
    api.POST("/run", httpserver.CreateJsonHandler(runHandler))
    api.GET("/:id/export", httpserver.CreateUriHandler(exportHandler))

    return definitions, nil
}

```

The repeated setup is the important part. The router has to construct each handler separately, handle each construction error separately, and often repeat dependency initialization for related endpoints.

With the new package, the server creates the router and passes it into your router factory. Using `httpserver.With`, one cohesive handler service can own the whole route group:

```go
func main() {
    httpserver.RunDefaultServer(DefineRouter)
}

func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
    router.Group("/api/reports").HandleWith(httpserver.With(NewReportHandler, func(r *httpserver.Router, h *ReportHandler) {
        r.Use(authMiddleware)
        r.GET("/overview", httpserver.BindN(h.Overview))
        r.POST("/run", httpserver.Bind(h.Run))
        r.GET("/:id/export", httpserver.Bind(h.Export))
    }))

    return nil
}
```

The router is no longer something your application has to allocate and return. Your code mutates the provided router and returns an error only if setup failed.

The handler setup also becomes smaller. `ReportHandler` is created once, shared dependencies are initialized once, and each endpoint maps to a method on the same service. The route group now reads as one unit instead of a sequence of unrelated handler allocations.

This also makes explicit application wiring simpler. If an application already uses `application.Run` with explicit modules, the HTTP server can be registered as a module factory:

```go
func main() {
    application.Run(
        application.WithModuleFactory("http", httpserver.NewServer("default", DefineRouter)),
    )
}
```

## One Handler for a Route Group

The `httpserver.With` pattern is useful because many APIs are organized around capabilities, not individual routes. A reports API might expose an overview endpoint, a run endpoint, and an export endpoint. Those are different HTTP operations, but they usually share the same service dependencies and belong to the same handler type.

The handler methods stay focused on individual endpoints:

```go
func (h *ReportHandler) Overview(ctx context.Context) (httpserver.Response, error) {
    overview, err := h.reports.GetOverview(ctx)
    if err != nil {
        return nil, err
    }

    return httpserver.NewJsonResponse(overview), nil
}

func (h *ReportHandler) Run(ctx context.Context, input *RunReportInput) (httpserver.Response, error) {
    result, err := h.reports.Run(ctx, input.ReportType, input.Parameters)
    if err != nil {
        return nil, err
    }

    return httpserver.NewJsonResponse(result), nil
}

func (h *ReportHandler) Export(ctx context.Context, input *ExportReportInput) (httpserver.Response, error) {
    file, err := h.exporter.Export(ctx, input.Id, input.Format)
    if err != nil {
        return nil, err
    }

    return httpserver.NewResponse(
        httpserver.WithBody(file.Body),
        httpserver.WithHeader("Content-Type", file.ContentType),
    ), nil
}
```

This consolidation is not required for every route. Separate handlers are still fine when routes have genuinely different dependencies or lifecycle requirements. But for most cohesive route groups, one handler service with multiple endpoint methods is the cleaner shape.

It also makes later changes easier. Adding a new report endpoint usually means adding one method and one route registration inside the same group, not introducing another handler type, another factory, and another block of setup code.

## Typed Request Binding

Request binding is another area where the new package is easier to work with.

With the old package, handlers usually implemented `GetInput`, then unpacked the bound value from `request.Body`:

```go
type RunReportInput struct {
    ReportType string         `json:"reportType" binding:"required"`
    Parameters map[string]any `json:"parameters"`
}

func (h *runReportHandler) GetInput() any {
    return &RunReportInput{}
}

func (h *runReportHandler) Handle(ctx context.Context, request *httpserver.Request) (*httpserver.Response, error) {
    input := request.Body.(*RunReportInput)

    result, err := h.reports.Run(ctx, input.ReportType, input.Parameters)
    if err != nil {
        return nil, err
    }

    return httpserver.NewJsonResponse(result), nil
}
```

With the new package, the input is part of the handler method signature:

```go
type RunReportInput struct {
    ReportType string         `json:"reportType" binding:"required"`
    Parameters map[string]any `json:"parameters"`
}

func (h *ReportHandler) Run(ctx context.Context, input *RunReportInput) (httpserver.Response, error) {
    result, err := h.reports.Run(ctx, input.ReportType, input.Parameters)
    if err != nil {
        return nil, err
    }

    return httpserver.NewJsonResponse(result), nil
}
```

Registering the route with `httpserver.Bind(h.Run)` tells the package to bind the request into `RunReportInput` before calling the handler.

This makes the data flow visible from the function signature. There is no generic request body to type assert, and handler methods describe exactly what they need.

## A Cleaner Direction for HTTP Services

The standalone `httpserver` package is part of the broader direction for gosoline: smaller packages, clearer APIs, and fewer dependencies for applications that only need a focused part of the framework.

Existing applications do not need to migrate immediately. The old package remains available for now, but new applications should prefer the standalone module.

For the full API overview, see the [HTTP server package reference](/reference/package-httpserver).  
For step-by-step migration details, see the [HTTP server package migration guide](/migrations/httpserver-package).
