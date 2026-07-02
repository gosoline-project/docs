# sqlh - SQL HTTP Handlers

The `sqlh` package exposes database entities as REST API endpoints. Built on top of [`sqlr`](/docs/how-to/databases-sql/sqlr.md) and [`sqlc`](/docs/how-to/databases-sql/sqlc.md), it provides automatic CRUD handler generation, a transformer pattern for input/output mapping, repository customization hooks, and transaction middleware for wrapping HTTP requests in database transactions.

## Getting Started[​](#getting-started "Direct link to Getting Started")

Add the dependency to your Go module:

```
go get github.com/gosoline-project/sqlh@v0.7.0
```

Then import the package in your Go code:

```
import "github.com/gosoline-project/sqlh"
```

## Configuration[​](#configuration "Direct link to Configuration")

The `sqlh` package requires both an HTTP server and a database client. Configure them under the `httpserver` and `sqlc` keys respectively:

config.dist.yml

```
httpserver:

  default:

    port: 8080

    mode: release



sqlc:

  default:

    driver: mysql

    uri:

      host: 127.0.0.1

      port: 3306

      user: root

      password: gosoline

      database: blog

    migrations:

      enabled: true

      reset: true

      path: migrations
```

The HTTP server configuration is described in the httpserver documentation. The database configuration follows the same format described in the [sqlc documentation](/docs/how-to/databases-sql/sqlc.md#configuration).

## CRUD Handlers[​](#crud-handlers "Direct link to CRUD Handlers")

The `WithCrudHandlers` function generates a complete set of REST endpoints for an entity. It connects an `sqlr` repository with a transformer to handle input/output mapping, and accepts optional configuration for custom repository setup.

### Defining Entities[​](#defining-entities "Direct link to Defining Entities")

Entities use the same `sqlr.Entity` base struct described in the [sqlr documentation](/docs/how-to/databases-sql/sqlr.md#defining-entities):

main.go

```
type Author struct {

  sqlr.Entity[int64]

  Name  string `db:"name"`

  Email string `db:"email"`

}
```

### Input Types[​](#input-types "Direct link to Input Types")

Define separate types for create and update input. These decouple the HTTP API from the database entity:

main.go

```
type AuthorCreateInput struct {

  Name  string `json:"name" binding:"required"`

  Email string `json:"email" binding:"required"`

}



type AuthorUpdateInput struct {

  Name string `json:"name" binding:"required"`

}
```

Using separate input types lets you:

* Validate input with `binding` tags (e.g., `binding:"required"`)
* Exclude internal fields (like `Id` or `CreatedAt`) from create/update payloads

### Implementing the Transformer[​](#implementing-the-transformer "Direct link to Implementing the Transformer")

The `Transformer[K, E, IC, IU]` interface converts between input types and entity types, and controls how responses are rendered. Implement `TransformCreateInput`, `TransformUpdateInput`, `RenderEntityResponse`, and `RenderQueryResponse`:

main.go

```
type AuthorTransformer struct{}



func (t *AuthorTransformer) TransformCreateInput(_ context.Context, input *AuthorCreateInput) (*Author, error) {

  return &Author{

    Name:  input.Name,

    Email: input.Email,

  }, nil

}



func (t *AuthorTransformer) TransformUpdateInput(_ context.Context, entity *Author, input *AuthorUpdateInput) (*Author, error) {

  entity.Name = input.Name



  return entity, nil

}



func (t *AuthorTransformer) RenderEntityResponse(_ context.Context, entity *Author) (httpserver.Response, error) {

  return httpserver.NewJsonResponse(entity), nil

}



func (t *AuthorTransformer) RenderQueryResponse(_ context.Context, entities []Author) (httpserver.Response, error) {

  return httpserver.NewJsonResponse(entities), nil

}
```

The type parameters are:

| Parameter | Description                                   |
| --------- | --------------------------------------------- |
| `K`       | Primary key type (e.g., `int64`)              |
| `E`       | Entity type (e.g., `Author`)                  |
| `IC`      | Create input type (e.g., `AuthorCreateInput`) |
| `IU`      | Update input type (e.g., `AuthorUpdateInput`) |

The interface requires four methods:

| Method                 | Description                                                  |
| ---------------------- | ------------------------------------------------------------ |
| `TransformCreateInput` | Converts a create input DTO into a new entity                |
| `TransformUpdateInput` | Merges an update input DTO into an existing entity           |
| `RenderEntityResponse` | Serialises a single entity into an `httpserver.Response`     |
| `RenderQueryResponse`  | Serialises a slice of entities into an `httpserver.Response` |

In the simple case above, `RenderEntityResponse` and `RenderQueryResponse` return the entity directly as JSON. See [Customizing Output Transformers](#customizing-output-transformers) for approaches when you need a separate output shape.

### Registering CRUD Handlers[​](#registering-crud-handlers "Direct link to Registering CRUD Handlers")

Use `WithCrudHandlers` to generate all endpoints and register them with the HTTP server:

main.go

```
func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

  router.HandleWith(sqlh.WithCrudHandlers(1, "author", sqlh.SimpleTransformer(&AuthorTransformer{})))



  return nil

},
```

The arguments are:

| Argument             | Description                                                             |
| -------------------- | ----------------------------------------------------------------------- |
| `version`            | API version number, used in the URL path (e.g., `1` produces `/v1/...`) |
| `entityName`         | Singular entity name for the URL path (e.g., `"author"`)                |
| `transformerFactory` | Factory that creates the transformer instance                           |
| `options`            | Optional `sqlh.Option` values that customize repository creation        |

By default, CRUD handlers create an `sqlr.Repository` against the `default` SQL client. Pass additional options to change that behavior:

* `sqlh.WithClientName[K, E](name)` selects a different configured SQL client for the default repository factory.
* `sqlh.WithRepositoryFactory[K, E](factory)` replaces repository construction entirely.

### Generated Endpoints[​](#generated-endpoints "Direct link to Generated Endpoints")

`WithCrudHandlers` registers five endpoints:

| Method   | Path                 | Handler        | Description                                       |
| -------- | -------------------- | -------------- | ------------------------------------------------- |
| `POST`   | `/v{n}/{entity}`     | `HandleCreate` | Creates an entity from `IC` input                 |
| `GET`    | `/v{n}/{entity}/:id` | `HandleRead`   | Reads a single entity by ID                       |
| `PUT`    | `/v{n}/{entity}/:id` | `HandleUpdate` | Updates an entity from `IU` input                 |
| `DELETE` | `/v{n}/{entity}/:id` | `HandleDelete` | Deletes an entity by ID; returns `204 No Content` |
| `POST`   | `/v{n}/{entities}`   | `HandleQuery`  | Queries entities with a JSON filter               |

The query endpoint uses the **plural** form of the entity name (e.g., `"author"` becomes `"/v1/authors"`). Pluralization is handled automatically.

### Handler Behavior[​](#handler-behavior "Direct link to Handler Behavior")

All generated CRUD handlers follow the same high-level pattern: bind input, delegate persistence to `sqlr.Repository`, and render the result through the configured transformer.

| Handler        | Flow                                                                                                                                                                          |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `HandleCreate` | Binds `IC`, calls `TransformCreateInput`, persists via `repo.Create()`, then renders the created entity (including any post-create preloads configured on the create builder) |
| `HandleRead`   | Loads a single entity via `repo.Read()` and renders it                                                                                                                        |
| `HandleQuery`  | Converts the JSON filter into an `sqlc.Expression`, calls `repo.Query()`, then renders the result list                                                                        |
| `HandleUpdate` | Loads the existing entity via `repo.Read()`, applies `TransformUpdateInput`, persists via `repo.Update()`, then renders the entity returned by `Update()`                     |
| `HandleDelete` | Deletes the entity via `repo.Delete()` and returns `204 No Content`                                                                                                           |

The update flow is slightly different from the others: `HandleUpdate` renders the rehydrated entity returned by `sqlr.Update()` rather than issuing a separate follow-up `Read()` call. That means post-update preloads configured on the update builder are reflected directly in the HTTP response. See [`sqlr` Update with Post-Update Preloading](/docs/how-to/databases-sql/sqlr.md#update-with-post-update-preloading) for the underlying repository behavior.

### Query with JSON Filter[​](#query-with-json-filter "Direct link to Query with JSON Filter")

The query endpoint accepts a JSON body with a `filter` field that maps to `sqlc.JsonFilter`:

```
{

  "filter": {

    "column": "name",

    "operator": "=",

    "value": "Alice"

  }

}
```

The filter is converted to an `sqlc.Expression` and applied as a WHERE condition on the query. See the [sqlc JSON filter documentation](/docs/how-to/databases-sql/sqlc.md) for the full filter syntax.

### Association Tags[​](#association-tags "Direct link to Association Tags")

CRUD handlers can also configure association loading and syncing directly from entity relationship fields with the `sqlh` struct tag:

```
type Author struct {

    sqlr.Entity[int64]

    ProfileID int64         `db:"profile_id"`

    Profile   Profile       `db:"-" sqlr:"belongsTo:profile_id" sqlh:"preload:create,read,query,update"`

    Tags      []Tag         `db:"-" sqlr:"many2many:author_tags" sqlh:"preload:create,read;sync:create,update,delete"`

}
```

Supported directives are:

| Directive | Phases                              | Effect                                                                                                                                    |
| --------- | ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `preload` | `create`, `read`, `query`, `update` | Adds `Preload()` to the matching CRUD builder; on `create` and `update`, it configures the post-write reload used for the returned entity |
| `sync`    | `create`, `update`, `delete`        | Calls `SyncAssociation()` for the association while persisting or cleaning up the entity                                                  |

`sqlh` tags are additive to the association behavior already defined by `sqlr`. The relationship itself must still be declared via an explicit `sqlr` relationship tag or discovered through [`sqlr` auto-detected relationships](/docs/how-to/databases-sql/sqlr.md#auto-detected-relationships); the `sqlh` tag only adds CRUD-specific preload and sync behavior on top. `db:"-"` remains optional for relationship fields. In particular, `sqlh` preload phases are merged with any preload behavior already configured in `sqlr`, rather than replacing it.

For creates, `sqlh:"preload:create"` is applied to the underlying `sqlr.Create()` builder, so the entity returned by the create handler is reloaded with those relations already hydrated. This follows the same post-create preload behavior described in [`sqlr` Create with Post-Create Preloading](/docs/how-to/databases-sql/sqlr.md#create-with-post-create-preloading).

For updates, `sqlh:"preload:update"` is applied to the underlying `sqlr.Update()` builder, so the entity returned by the update handler is reloaded with those relations already hydrated. This follows the same post-update preload behavior described in [`sqlr` Update with Post-Update Preloading](/docs/how-to/databases-sql/sqlr.md#update-with-post-update-preloading).

For delete operations, `sync:delete` configures the association cleanup passed to the underlying `sqlr.Delete()` call, following the same cleanup semantics described in [`sqlr` Delete with Association Cleanup](/docs/how-to/databases-sql/sqlr.md#delete-with-association-cleanup).

Tags are only valid on association fields recognised by [`sqlr` relationships](/docs/how-to/databases-sql/sqlr.md#relationships); using them on scalar fields causes startup to fail with an error. Tagged associations are traversed recursively, so nested paths such as `Child.Nested` are picked up automatically when the related entity also declares `sqlh` tags.

For the underlying `sqlr` behavior behind these options, see [Read with Association Loading](/docs/how-to/databases-sql/sqlr.md#read-with-association-loading), [Eager Loading with Preload](/docs/how-to/databases-sql/sqlr.md#eager-loading-with-preload), [Create with Selective Association Persistence](/docs/how-to/databases-sql/sqlr.md#create-with-selective-association-persistence), [Create with Post-Create Preloading](/docs/how-to/databases-sql/sqlr.md#create-with-post-create-preloading), [Update with Association Sync](/docs/how-to/databases-sql/sqlr.md#update-with-association-sync), and [Delete with Association Cleanup](/docs/how-to/databases-sql/sqlr.md#delete-with-association-cleanup).

### Wiring into the Application[​](#wiring-into-the-application "Direct link to Wiring into the Application")

Register handlers with the gosoline HTTP server using `router.HandleWith()`:

main.go

```
func main() {

  application.New(

    application.WithConfigBytes(config, "yml"),

    application.WithLoggerHandlersFromConfig,

    application.WithModuleFactory("http", httpserver.NewServer(

      "default",

      func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

        router.HandleWith(sqlh.WithCrudHandlers(1, "author", sqlh.SimpleTransformer(&AuthorTransformer{})))



        return nil

      },

    )),

  ).Run()

}
```

## Customizing Output Transformers[​](#customizing-output-transformers "Direct link to Customizing Output Transformers")

The `Transformer` interface is the primary extension point for controlling how entities are serialised in API responses. Because it is a plain Go interface, a single implementation can be written once and reused across multiple entities — or shared as an internal library across services. This makes it straightforward to enforce consistent response shapes, pagination envelopes, or field-level access control in one place.

### JsonResultsTransformer[​](#jsonresultstransformer "Direct link to JsonResultsTransformer")

`JsonResultsTransformer[K, E, IC, IU]` is a simplified variant of `Transformer` for the common case where you want to map entities to a dedicated output type and return them as JSON, without needing to construct `httpserver.Response` values manually. Implement three methods — `TransformCreateInput`, `TransformUpdateInput`, and a single `TransformOutput` that converts one entity to any JSON-serialisable value — and pass the implementation to `NewJsonResultsTransformer`. The wrapping of single and multi-entity responses into JSON is handled automatically, and optional CRUD builder hooks implemented by the transformer — including create hooks, delete hooks, and the split update read/write hooks — are forwarded too:

user\_crud.go

```
type (

  UserCreateInput struct {

    Name string `json:"name"`

  }

  UserUpdateInput struct {

    Name string `json:"name"`

  }

  User struct {

    sqlr.Entity[int]

    Name string

  }

  UserOutput struct {

    Id        int       `json:"id"`

    Name      string    `json:"name"`

    CreatedAt time.Time `json:"created_at"`

    UpdatedAt time.Time `json:"updated_at"`

  }

)
```

user\_crud.go

```
type UserTransformer struct{}



func (t *UserTransformer) TransformCreateInput(ctx context.Context, input *UserCreateInput) (*User, error) {

  return &User{

    Name: input.Name,

  }, nil

}



func (t *UserTransformer) TransformUpdateInput(ctx context.Context, user *User, input *UserUpdateInput) (*User, error) {

  user.Name = input.Name



  return user, nil

}



func (t *UserTransformer) TransformOutput(ctx context.Context, user *User) (any, error) {

  return UserOutput{

    Id:        user.Id,

    Name:      user.Name,

    CreatedAt: user.CreatedAt,

    UpdatedAt: user.UpdatedAt,

  }, nil

}
```

user\_crud.go

```
func NewUserCrud() httpserver.RegisterFactoryFunc {

  return sqlh.WithCrudHandlers(0, "user", sqlh.NewJsonResultsTransformer(&UserTransformer{}))

}
```

Because `TransformOutput` receives a single entity, the same function is used for both single-entity and list responses — there is no duplication. Use the full `Transformer` interface directly when you need control over HTTP status codes, headers, or non-JSON response bodies.

### Wiring Transformers[​](#wiring-transformers "Direct link to Wiring Transformers")

Use `sqlh.SimpleTransformer()` to wrap an already-constructed transformer into a factory when it has no startup dependencies:

```
sqlh.SimpleTransformer[K, E, IC, IU](&MyTransformer{})
```

For transformers that require configuration or other dependencies at startup, implement `TransformerFactory` directly:

```
type TransformerFactory[K, E, IC, IU] func(ctx context.Context, config cfg.Config, logger log.Logger) (Transformer[K, E, IC, IU], error)
```

This follows the standard gosoline factory pattern and gives the transformer access to the application config and logger during initialisation.

## CRUD Builder Hooks[​](#crud-builder-hooks "Direct link to CRUD Builder Hooks")

Both `Transformer` and `JsonResultsTransformer` implementations can optionally customize generated CRUD queries by implementing one or more builder-aware interfaces:

| Interface                 | Builder                   | Used for                                                                                                               |
| ------------------------- | ------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `BuilderCreateAware`      | `sqlr.QueryBuilderCreate` | Customize create persistence                                                                                           |
| `BuilderReadAware`        | `sqlr.QueryBuilderRead`   | Customize single-entity reads                                                                                          |
| `BuilderQueryAware`       | `sqlr.QueryBuilderSelect` | Customize list/query requests                                                                                          |
| `BuilderDeleteAware`      | `sqlr.QueryBuilderDelete` | Customize delete cleanup                                                                                               |
| `BuilderUpdateReadAware`  | `sqlr.QueryBuilderRead`   | Customize the read of the existing entity before `TransformUpdateInput` runs                                           |
| `BuilderUpdateWriteAware` | `sqlr.QueryBuilderUpdate` | Customize the `Update()` call itself, including association sync and post-update preloads used for the returned entity |

The update flow is intentionally split across two hooks. Use `BuilderUpdateReadAware` when the transformer needs additional relations loaded before it merges the incoming update DTO into the existing entity. Use `BuilderUpdateWriteAware` when you need to change how `sqlr.Update()` executes, including which relations are synchronized and which associations are preloaded onto the rehydrated entity returned in the HTTP response.

## Custom Repository Implementations[​](#custom-repository-implementations "Direct link to Custom Repository Implementations")

By default, `WithCrudHandlers` creates a regular `sqlr.Repository` using the `default` SQL client. When you need different startup behavior, you can customize repository creation with dedicated options instead of changing your handler logic.

### Choosing an Option[​](#choosing-an-option "Direct link to Choosing an Option")

* Use `sqlh.WithClientName[K, E](name)` when the standard `sqlr.NewRepository()` behavior is correct and you only want to point it at a different configured SQL client.
* Use `sqlh.WithRepositoryFactory[K, E](factory)` when you want full control over repository construction, such as wrapping the default repository, adding opinionated query defaults, or returning a completely custom implementation.

The factory is called once during application startup and receives the application `context`, `config`, `logger`, and the selected client name.

### Example Configuration[​](#example-configuration "Direct link to Example Configuration")

This example configures both a `default` and a `reporting` SQL client. The custom repository factory honors the configured client name, so the same wrapper can be reused with either connection:

config.dist.yml

```
httpserver:

  default:

    port: 8080

    mode: release



sqlc:

  default:

    driver: mysql

    uri:

      host: 127.0.0.1

      port: 3306

      user: root

      password: gosoline

      database: blog

    migrations:

      enabled: true

      reset: true

      path: migrations

  reporting:

    driver: mysql

    uri:

      host: 127.0.0.1

      port: 3306

      user: root

      password: gosoline

      database: blog
```

### Wrapping the Default Repository[​](#wrapping-the-default-repository "Direct link to Wrapping the Default Repository")

In most cases, the simplest custom implementation is a small wrapper around `sqlr.NewRepository()`. That lets you preserve the standard CRUD behavior while adding defaults in selected methods:

main.go

```
type ReportingAuthorRepository struct {

  delegate sqlr.Repository[int64, Author]

}



func NewReportingAuthorRepository(ctx context.Context, config cfg.Config, logger log.Logger, name string) (sqlr.Repository[int64, Author], error) {

  repo, err := sqlr.NewRepository[int64, Author](ctx, config, logger, name)

  if err != nil {

    return nil, err

  }



  return &ReportingAuthorRepository{delegate: repo}, nil

}



func (r *ReportingAuthorRepository) Create(ctx context.Context, entity *Author, opts ...func(qb *sqlr.QueryBuilderCreate)) error {

  return r.delegate.Create(ctx, entity, opts...)

}



func (r *ReportingAuthorRepository) Read(ctx context.Context, id int64, opts ...func(qb *sqlr.QueryBuilderRead)) (*Author, error) {

  return r.delegate.Read(ctx, id, opts...)

}



func (r *ReportingAuthorRepository) Query(ctx context.Context, opts ...func(qb *sqlr.QueryBuilderSelect)) ([]Author, error) {

  return r.delegate.Query(ctx, append(opts, func(qb *sqlr.QueryBuilderSelect) {

    qb.OrderBy("created_at DESC")

    qb.Limit(100)

  })...)

}



func (r *ReportingAuthorRepository) Update(ctx context.Context, entity *Author, opts ...func(qb *sqlr.QueryBuilderUpdate)) (*Author, error) {

  return r.delegate.Update(ctx, entity, opts...)

}



func (r *ReportingAuthorRepository) Delete(ctx context.Context, id int64, opts ...func(qb *sqlr.QueryBuilderDelete)) error {

  return r.delegate.Delete(ctx, id, opts...)

}



func (r *ReportingAuthorRepository) Close() error {

  return r.delegate.Close()

}
```

This pattern is useful when you want to add metrics, logging, tenant-aware setup, or query defaults without reimplementing the full `sqlr.Repository` interface.

### Registering CRUD Handlers with a Custom Repository[​](#registering-crud-handlers-with-a-custom-repository "Direct link to Registering CRUD Handlers with a Custom Repository")

Pass both options to `WithCrudHandlers` when you want to select a non-default SQL client and override how the repository is created:

main.go

```
func main() {

  application.New(

    application.WithConfigBytes(config, "yml"),

    application.WithLoggerHandlersFromConfig,

    application.WithModuleFactory("http", httpserver.NewServer(

      "default",

      func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

        router.HandleWith(sqlh.WithCrudHandlers(

          1,

          "author",

          sqlh.SimpleTransformer(&AuthorTransformer{}),

          sqlh.WithClientName[int64, Author]("reporting"),

          sqlh.WithRepositoryFactory[int64, Author](NewReportingAuthorRepository),

        ))



        return nil

      },

    )),

  ).Run()

}
```

The custom repository must still satisfy `sqlr.Repository[K, E]`, so `sqlh` can continue to call `Create`, `Read`, `Query`, `Update`, and `Delete` normally.

## Transaction Middleware[​](#transaction-middleware "Direct link to Transaction Middleware")

The `WithTx` function wraps a group of HTTP routes in a database transaction. Each request automatically begins a transaction before the handler runs, commits on success, and rolls back if any error occurs.

### Setting Up WithTx[​](#setting-up-withtx "Direct link to Setting Up WithTx")

Create a handler struct with a factory function, then use `WithTx` to register routes:

main.go

```
type PostHandler struct {

}



func NewPostHandler() httpserver.HandlerFactory[PostHandler] {

  return func(ctx context.Context, config cfg.Config, logger log.Logger) (*PostHandler, error) {

    return &PostHandler{}, nil

  }

}
```

main.go

```
func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

  router.HandleWith(sqlh.WithTx(NewPostHandler(), func(router *httpserver.Router, handler *PostHandler) {

    router.POST("/v1/authors/:id/posts", sqlh.BindTx(handler.HandleCreatePost))

    router.GET("/v1/posts/:id", sqlh.BindTxN(handler.HandleReadPost))

  }))



  return nil

},
```

Notice that the handler struct does not create or hold a repository. Instead of manually setting up database access, each handler receives an active `sqlc.Tx` directly through the `Bind*` helpers — the transaction is started automatically before the handler runs and committed or rolled back afterward.

`WithTx` takes two arguments:

| Argument         | Description                                                                    |
| ---------------- | ------------------------------------------------------------------------------ |
| `handlerFactory` | Factory that creates the handler struct (with access to `config` and `logger`) |
| `register`       | Function that registers routes on the router, receiving the handler instance   |

The middleware:

1. Begins a transaction via `sqlClient.BeginTx()` before each request
2. Stores the transaction in the gin context
3. Calls the next handler
4. **Commits** if no errors occurred
5. **Rolls back** if any handler in the chain added an error to the gin context

### Transaction Binding Helpers[​](#transaction-binding-helpers "Direct link to Transaction Binding Helpers")

Inside a `WithTx`-wrapped route group, use the `BindTx` family of functions to extract the transaction and bind request input:

#### BindTx — With Input[​](#bindtx--with-input "Direct link to BindTx — With Input")

Use `BindTx` when the handler needs both the transaction and a parsed request body:

main.go

```
func (h *PostHandler) HandleCreatePost(cttx sqlc.Tx, input *PostCreateInput) (httpserver.Response, error) {

  // The transaction is automatically managed — commit on success, rollback on error.

  // Use cttx.Q() to execute queries within the transaction scope.



  post := &Post{

    Title:  input.Title,

    Body:   input.Body,

    Status: "draft",

  }



  _, err := cttx.Q().Into("posts").Records(post).Exec(cttx)

  if err != nil {

    return nil, fmt.Errorf("failed to create post: %w", err)

  }



  return httpserver.NewJsonResponse(post), nil

}
```

`BindTx` automatically:

* Binds the request body to the input type `I`
* Extracts the `sqlc.Tx` from the gin context
* Calls the handler with both
* Writes the response

#### BindTxN — No Input[​](#bindtxn--no-input "Direct link to BindTxN — No Input")

Use `BindTxN` when the handler only needs the transaction (no request body):

main.go

```
func (h *PostHandler) HandleReadPost(cttx sqlc.Tx) (httpserver.Response, error) {

  // BindTxN is used when no request body input is needed.

  // The transaction is still available for database operations.



  var posts []Post

  err := cttx.Q().From("posts").Where(sqlc.Col("status").Eq("draft")).Select(cttx, &posts)

  if err != nil {

    return nil, fmt.Errorf("failed to query posts: %w", err)

  }



  return httpserver.NewJsonResponse(posts), nil

}
```

Since `sqlc.Tx` implements `Querier`, you can pass it to `WithClient()` on query builders to execute queries within the transaction.

#### BindTxR / BindTxNR — With Raw Request[​](#bindtxr--bindtxnr--with-raw-request "Direct link to BindTxR / BindTxNR — With Raw Request")

For handlers that need access to the raw `*http.Request` (e.g., to read headers or query parameters), use the `R` variants:

```
// With input + raw request

sqlh.BindTxR(func(cttx sqlc.Tx, req *http.Request, input *MyInput) (httpserver.Response, error) {

    userAgent := req.Header.Get("User-Agent")

    // ...

})



// No input + raw request

sqlh.BindTxNR(func(cttx sqlc.Tx, req *http.Request) (httpserver.Response, error) {

    // ...

})
```

### Wiring into the Application[​](#wiring-into-the-application-1 "Direct link to Wiring into the Application")

main.go

```
func main() {

  application.New(

    application.WithConfigBytes(config, "yml"),

    application.WithLoggerHandlersFromConfig,

    application.WithModuleFactory("http", httpserver.NewServer(

      "default",

      func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

        router.HandleWith(sqlh.WithTx(NewPostHandler(), func(router *httpserver.Router, handler *PostHandler) {

          router.POST("/v1/authors/:id/posts", sqlh.BindTx(handler.HandleCreatePost))

          router.GET("/v1/posts/:id", sqlh.BindTxN(handler.HandleReadPost))

        }))



        return nil

      },

    )),

  ).Run()

}
```

### Summary of Binding Functions[​](#summary-of-binding-functions "Direct link to Summary of Binding Functions")

| Function   | Input | Raw Request | Signature                                                           |
| ---------- | ----- | ----------- | ------------------------------------------------------------------- |
| `BindTx`   | Yes   | No          | `func(cttx sqlc.Tx, input *I) (Response, error)`                    |
| `BindTxR`  | Yes   | Yes         | `func(cttx sqlc.Tx, req *http.Request, input *I) (Response, error)` |
| `BindTxN`  | No    | No          | `func(cttx sqlc.Tx) (Response, error)`                              |
| `BindTxNR` | No    | Yes         | `func(cttx sqlc.Tx, req *http.Request) (Response, error)`           |
