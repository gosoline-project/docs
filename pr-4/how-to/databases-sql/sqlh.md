# sqlh - SQL HTTP Handlers

The `sqlh` package provides typed CRUD handlers for SQLR entities. It uses the current `httpserver`, `sqlc`, and `sqlr` APIs. Each CRUD operation runs in a database transaction and returns a typed value.

CRUD definitions use typed callbacks and transaction-aware repositories instead of transformer interfaces and request-scoped Gin transaction bindings.

## Getting Started[​](#getting-started "Direct link to Getting Started")

Add the SQLH version that provides the typed CRUD API to your Go module:

```
go get github.com/gosoline-project/sqlh@v0.7.2-0.20260914134957-1ff7ba06abce
```

Use the released SQLH version that provides this API when one is available.

The typed API uses `github.com/gosoline-project/httpserver`, `github.com/gosoline-project/sqlc`, and `github.com/gosoline-project/sqlr`. Use compatible versions of all three packages. The examples use the `sqlc` `v0.4.0` and `sqlr` `v0.9.0` APIs.

Then import the package:

```
import "github.com/gosoline-project/sqlh"
```

Configure the HTTP server and the SQL client before the handler starts:

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

## CRUD Model[​](#crud-model "Direct link to CRUD Model")

SQLH separates four concerns:

| Concern        | Callback                                                                    |
| -------------- | --------------------------------------------------------------------------- |
| Create mapping | `CreateInput` maps an HTTP input to a new entity.                           |
| Update mapping | `UpdateInput` applies a complete input to a loaded entity.                  |
| Patch mapping  | `PatchInputFromEntity` creates the complete input used by JSON Merge Patch. |
| Output mapping | `Output` maps an entity to a typed response value.                          |

The update input must expose the route identity. Embed `sqlh.InputById[K]` in the input type. The embedded `Id` field uses the `id` URI parameter and does not bind from JSON.

```
type AuthorUpdateInput struct {

    sqlh.InputById[int64]

    Name string `json:"name" binding:"required"`

}
```

`InputById` also carries server-owned force filters. Use it when the read, update, patch, or delete lookup must respect request-specific authorization restrictions.

### Define an Entity and Callbacks[​](#define-an-entity-and-callbacks "Direct link to Define an Entity and Callbacks")

The following example uses one definition for create, update, patch, and output mapping:

main.go

```
type Author struct {

  sqlr.Entity[int64]

  Name  string `db:"name"`

  Email string `db:"email"`

}
```

main.go

```
type AuthorCreateInput struct {

  Name  string `json:"name" binding:"required"`

  Email string `json:"email" binding:"required,email"`

}



type AuthorUpdateInput struct {

  sqlh.InputById[int64]

  Name string `json:"name" binding:"required"`

}



type AuthorOutput struct {

  Id        int64     `json:"id"`

  Name      string    `json:"name"`

  Email     string    `json:"email"`

  CreatedAt time.Time `json:"created_at"`

  UpdatedAt time.Time `json:"updated_at"`

}
```

`NewCrudDefinition` accepts the four required callbacks. `SimpleCrudDefinition` wraps the definition in the application factory shape:

main.go

```
type AuthorMapper struct{}



func (t *AuthorMapper) TransformCreateInput(_ context.Context, input *AuthorCreateInput) (*Author, error) {

  return &Author{

    Name:  input.Name,

    Email: input.Email,

  }, nil

}



func (t *AuthorMapper) TransformUpdateInput(_ context.Context, entity *Author, input *AuthorUpdateInput) (*Author, error) {

  entity.Name = input.Name



  return entity, nil

}



func (t *AuthorMapper) TransformPatchInputFromEntity(_ context.Context, entity *Author) (*AuthorUpdateInput, error) {

  return &AuthorUpdateInput{

    InputById: sqlh.InputById[int64]{Id: entity.Id},

    Name:      entity.Name,

  }, nil

}



func (t *AuthorMapper) TransformOutput(_ context.Context, entity *Author) (AuthorOutput, error) {

  return AuthorOutput{

    Id:        entity.Id,

    Name:      entity.Name,

    Email:     entity.Email,

    CreatedAt: entity.CreatedAt,

    UpdatedAt: entity.UpdatedAt,

  }, nil

}
```

The output callback returns a normal Go value. The HTTP server selects the response representation. It does not require `httpserver.Response`.

For a dedicated response DTO, map the entity in the output callback:

user\_crud.go

```
type (

  UserCreateInput struct {

    Name string `json:"name"`

  }

  UserUpdateInput struct {

    sqlh.InputById[int]

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
type UserMapper struct{}



func (t *UserMapper) TransformCreateInput(_ context.Context, input *UserCreateInput) (*User, error) {

  return &User{

    Name: input.Name,

  }, nil

}



func (t *UserMapper) TransformUpdateInput(_ context.Context, user *User, input *UserUpdateInput) (*User, error) {

  user.Name = input.Name



  return user, nil

}



func (t *UserMapper) TransformPatchInputFromEntity(_ context.Context, user *User) (*UserUpdateInput, error) {

  return &UserUpdateInput{

    InputById: sqlh.InputById[int]{Id: user.Id},

    Name:      user.Name,

  }, nil

}



func (t *UserMapper) TransformOutput(_ context.Context, user *User) (UserOutput, error) {

  return UserOutput{

    Id:        user.Id,

    Name:      user.Name,

    CreatedAt: user.CreatedAt,

    UpdatedAt: user.UpdatedAt,

  }, nil

}
```

### Register Standard Routes[​](#register-standard-routes "Direct link to Register Standard Routes")

Use `WithCrudHandlers` when the standard route paths and delete response meet your API contract:

main.go

```
func NewAuthorCrud() httpserver.RegisterFactoryFunc {

  transformer := &AuthorMapper{}

  definition := sqlh.NewCrudDefinition(

    transformer.TransformCreateInput,

    transformer.TransformUpdateInput,

    transformer.TransformPatchInputFromEntity,

    transformer.TransformOutput,

  )



  return sqlh.WithCrudHandlers(1, "author", sqlh.SimpleCrudDefinition(definition))

}
```

The generated routes are:

| Method   | Path                          | Operation                    |
| -------- | ----------------------------- | ---------------------------- |
| `POST`   | `/v{version}/{entity}`        | Create                       |
| `GET`    | `/v{version}/{entity}/:id`    | Read                         |
| `PUT`    | `/v{version}/{entity}/:id`    | Complete update              |
| `PATCH`  | `/v{version}/{entity}/:id`    | JSON Merge Patch             |
| `DELETE` | `/v{version}/{entity}/:id`    | Delete with `204 No Content` |
| `POST`   | `/v{version}/{plural-entity}` | List and count               |

For example, version `1` and entity name `author` produce `/v1/author` and `/v1/authors`.

The default delete route uses `DeleteNoContent`. Use manual route registration with `handler.Delete` when the API must return the deleted entity:

```
router.HandleWith(httpserver.With(resource.NewHandler, func(router *httpserver.Router, handler *resource.CrudHandler) {

    router.DELETE("/v1/resources/:id", httpserver.Bind(handler.Delete, httpserver.NoBodyBinding{}))

}))
```

`handler.Delete` requires `CrudDefinition.DeleteOutput`. You can assign the normal output mapper when both responses have the same shape:

```
definition.DeleteOutput = definition.Output
```

### Register Custom Paths[​](#register-custom-paths "Direct link to Register Custom Paths")

Use `NewCrudHandler` when an existing API has different paths, singular and plural names, or response behavior:

```
func DefineRouter(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {

    router.HandleWith(httpserver.With(resource.NewHandler, func(router *httpserver.Router, handler *resource.CrudHandler) {

        router.POST("/v1/resource", httpserver.Bind(handler.Create))

        router.POST("/v1/resources", httpserver.Bind(handler.List))

        router.GET("/v1/resources/:id", httpserver.Bind(handler.Read, httpserver.NoBodyBinding{}))

        router.PUT("/v1/resources/:id", httpserver.Bind(handler.Update))

        router.PATCH("/v1/resources/:id", httpserver.Bind(handler.Patch))

        router.DELETE("/v1/resources/:id", httpserver.Bind(handler.Delete, httpserver.NoBodyBinding{}))

    }))



    return nil

}
```

This pattern preserves existing public routes while SQLH owns the transaction and CRUD flow.

## Transactions[​](#transactions "Direct link to Transactions")

SQLH creates one transaction for every default CRUD operation. It loads the entity, applies the mapper, synchronizes selected associations, maps the output, and commits only after the operation succeeds. Any operation, mapper, validation, repository, or output error rolls back the transaction. The HTTP server renders the result after the commit.

The SQLH repository must use the same SQL client that starts the transaction. This requirement is important when SQLR uses prepared statements. The default handler factory creates both objects from the configured client.

The old `sqlh.WithTx`, `BindTx`, `BindTxN`, `BindTxR`, and `BindTxNR` APIs are removed. Use a SQLH CRUD operation for entity CRUD. For another typed operation, use `TxRunner` and adapt it to the handler signature:

main.go

```
type PostHandler struct {

  runner *sqlh.TxRunner

}



func NewPostHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*PostHandler, error) {

  runner, err := sqlh.NewTxRunner(ctx, config, logger, "default")

  if err != nil {

    return nil, err

  }



  return &PostHandler{runner: runner}, nil

}



func (h *PostHandler) CreatePost(ctx context.Context, input *PostCreateInput) (PostOutput, error) {

  return h.runner.RunValue(ctx, input, func(ctx context.Context, tx sqlr.TTx, input *PostCreateInput) (PostOutput, error) {

    post := &Post{

      Title:  input.Title,

      Body:   input.Body,

      Status: "draft",

    }



    if _, err := tx.Q().Into("posts").Records(post).Exec(tx); err != nil {

      return PostOutput{}, fmt.Errorf("failed to create post: %w", err)

    }



    return PostOutput{

      Id:     post.Id,

      Title:  post.Title,

      Body:   post.Body,

      Status: post.Status,

    }, nil

  })

}
```

main.go

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

`TxRunner.Run` rolls back when the operation returns an error or panics. `RunValue` returns the value only after commit. `InTransaction` converts a transaction-aware operation to the function shape accepted by `httpserver.Bind`.

## Input and Output Types[​](#input-and-output-types "Direct link to Input and Output Types")

Use separate input types for create and update requests. This keeps database fields and server-owned values out of the HTTP request:

```
type AuthorCreateInput struct {

    Name  string `json:"name" binding:"required"`

    Email string `json:"email" binding:"required,email"`

}



type AuthorUpdateInput struct {

    sqlh.InputById[int64]

    Name string `json:"name" binding:"required"`

}



type AuthorOutput struct {

    Id        int64     `json:"id"`

    Name      string    `json:"name"`

    Email     string    `json:"email"`

    CreatedAt time.Time `json:"created_at"`

    UpdatedAt time.Time `json:"updated_at"`

}
```

Use `binding` tags for request validation. The `Output` callback can return an entity, a response DTO, or another value that the HTTP server can render.

The old `Transformer` and `JsonResultsTransformer` interfaces are removed. Replace their methods as follows:

| Old method                                  | New callback                          |
| ------------------------------------------- | ------------------------------------- |
| `TransformCreateInput`                      | `CrudDefinition.CreateInput`          |
| `TransformUpdateInput`                      | `CrudDefinition.UpdateInput`          |
| `TransformOutput` or `RenderEntityResponse` | `CrudDefinition.Output`               |
| No old equivalent                           | `CrudDefinition.PatchInputFromEntity` |

Use `NewCrudDefinition` for the standard callback set. Use a `CrudDefinition` value directly when you need custom list, identity, delete, or operation callbacks.

## JSON Merge Patch[​](#json-merge-patch "Direct link to JSON Merge Patch")

The default `PATCH` operation follows RFC 7396 JSON Merge Patch. It uses the same update mapper as `PUT`:

1. SQLH loads the entity with the update scope and update preloads.
2. `PatchInputFromEntity` maps the entity to a complete update input.
3. SQLH merges the request document into that input.
4. `UpdateInput` validates and applies the merged input.
5. SQLH synchronizes only associations selected by the original patch document.
6. SQLH maps the result and commits the transaction.

`PatchInputFromEntity` must populate every writable field that an omitted request field must preserve. An omitted scalar remains unchanged. An explicit JSON `null` clears a field when the input type permits it. An array replaces the complete array. A `null` or empty array clears a direct association.

```
func (t *AuthorMapper) TransformPatchInputFromEntity(_ context.Context, author *Author) (*AuthorUpdateInput, error) {

    return &AuthorUpdateInput{

        InputById: sqlh.InputById[int64]{Id: author.Id},

        Name:      author.Name,

    }, nil

}
```

SQLH derives direct association paths from update-input JSON tags and SQLR relation names. A relation must have `sqlh:"sync:update"` before SQLH can synchronize it during a patch. Use `PatchAssociations` when a JSON path does not match the relation name.

Use `PatchAssociationTriggers` when a scalar request field changes an association indirectly:

```
definition.PatchAssociationTriggers = map[string]string{

    "state": "Labels",

}
```

The key is a path in the original patch document. The value is an SQLR relation path. A trigger selects the relation for persistence. It does not change merge-patch null behavior or mutate the relation.

A custom `PatchOperation` replaces the complete default patch pipeline. Use it only when the default complete-input mapping and association selection cannot represent the API contract.

## List Inputs, Filters, and Counts[​](#list-inputs-filters-and-counts "Direct link to List Inputs, Filters, and Counts")

The standard list input contains a JSON SQLC filter, page limit, page offset, and server-owned force filters:

```
input := sqlh.ListInput{

    Filter: sqlc.JsonFilter{

        Type:   "eq",

        Column: "status",

        Value:  "active",

    },

    Page: sqlh.ListPage{

        Limit:  50,

        Offset: 100,

    },

}
```

The generated list endpoint accepts a body such as:

```
{

  "filter": {

    "type": "eq",

    "column": "status",

    "value": "active"

  },

  "page": {

    "limit": 50,

    "offset": 100

  }

}
```

SQLH returns a `ListOutput` value:

```
{

  "results": [],

  "total": 0

}
```

The default list operation applies the same user filter, force filters, and delete scope to the query and count. It applies ordering, grouping, and pagination only to the row query. The count ignores those row-only modifiers.

### Force Filters[​](#force-filters "Direct link to Force Filters")

Force filters are server-owned restrictions. They are not bound from HTTP input and the caller cannot remove them. Embed `sqlh.ForceFilters` through `sqlh.ListInput` or `sqlh.InputById`, then add filters after authentication:

```
input.AddForceFilter(func(qb *sqlr.QueryBuilderSelect) {

    qb.Where(sqlc.Col("account_id").Eq(accountID))

})
```

SQLH applies force filters to list, count, identity, update, and delete lookups. A force filter must only add restrictive `WHERE` conditions.

### Custom List Inputs[​](#custom-list-inputs "Direct link to Custom List Inputs")

Embed `sqlh.ListInput` when the service needs extra filter syntax:

```
type ListInput struct {

    sqlh.ListInput



    legacyQuery bool

}
```

A custom list input must implement `ListInputSource`:

* `ApplyFilters` adds user filters and joins.
* `ApplyQueryModifiers` adds grouping and ordering.
* `ApplyPagination` adds the page limit and offset.
* `ValidatePagination` validates the input.
* `GetForceFilters` exposes the embedded, server-owned filters.

Use the same filter adapter for query and count. Custom query callbacks must apply the query plan and scope, then query modifiers, then pagination. Custom count callbacks must apply the query plan and scope, but not query modifiers or pagination.

## SQLR and SQLH Relation Tags[​](#sqlr-and-sqlh-relation-tags "Direct link to SQLR and SQLH Relation Tags")

SQLR tags define the relationship. SQLH tags add CRUD-specific preload and synchronization phases:

```
type Project struct {

    sqlr.Entity[uint]



    Owner  *Owner  `db:"-" sqlr:"belongsTo:owner_id;preload" sqlh:"preload:create,read,query,update"`

    Rules  []*Rule  `db:"-" sqlr:"foreignKey:project_id;preload;sync:create,update" sqlh:"preload:create,read,query,update;sync:create,update"`

    Labels []*Label `db:"-" sqlr:"foreignKey:project_id;preload;sync:create,update" sqlh:"preload:create,read,query,update;sync:create,update"`

}
```

Supported SQLH directives are:

| Directive | Valid phases                        | Effect                                                                                                          |
| --------- | ----------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `preload` | `create`, `read`, `query`, `update` | Loads the relation in the matching operation. Create and update preloads apply to the post-write entity reload. |
| `sync`    | `create`, `update`, `delete`        | Selects the relation for association persistence or cleanup in the matching operation.                          |

The relation path uses Go field names, such as `Labels` or `Posts.Comments`. SQLH traverses nested relations. A `sqlh` tag on a scalar or embedded field is invalid. The relation must exist in the SQLR schema.

`sync:update` is required when PUT or PATCH must persist a relation. Without it, the mapper can change the in-memory graph, but SQLR does not reconcile the related rows.

## Custom Identity and Delete Behavior[​](#custom-identity-and-delete-behavior "Direct link to Custom Identity and Delete Behavior")

The default identity lookup uses the SQL primary key. Set `CrudDefinition.Identity` when the public API uses another key, such as a public UUID:

```
definition.Identity = func(

    ctx context.Context,

    tx sqlr.TTx,

    repository sqlr.RepositoryTx[uint, Project],

    publicID string,

    scope sqlh.QueryScope,

    builder func(*sqlr.QueryBuilderSelect),

) (*Project, error) {

    var scopeErr error

    entities, err := repository.Query(tx, func(qb *sqlr.QueryBuilderSelect) {

        qb.Where(sqlc.Col("projects", "public_id").Eq(publicID))

        if scope != nil {

            scopeErr = scope(qb)

        }

        if scopeErr == nil && builder != nil {

            builder(qb)

        }

    })

    if scopeErr != nil {

        return nil, scopeErr

    }

    if err != nil {

        return nil, err

    }

    if len(entities) == 0 {

        return nil, fmt.Errorf("project %s: %w", publicID, sqlr.ErrNotFound)

    }



    return &entities[0], nil

}
```

Always apply the supplied `scope`. It contains force filters, delete visibility, and other restrictions that protect read, update, patch, and delete operations.

SQLH uses physical SQLR deletion by default. Use `DeleteScope` and `Delete` for soft deletion:

```
definition.DeleteScope = func(qb *sqlr.QueryBuilderSelect) {

    qb.Where(sqlc.Col("deleted_at").IsNull())

}



definition.Delete = func(

    ctx context.Context,

    tx sqlr.TTx,

    repository sqlr.RepositoryTx[uint, Project],

    entity *Project,

) error {

    now := time.Now()

    entity.DeletedAt = &now

    entity.UpdatedAt = now



    updated, err := repository.Update(tx, entity, func(qb *sqlr.QueryBuilderUpdate) {

        qb.OmitAssociation("Rules", "Labels")

    })

    if err != nil {

        return err

    }

    *entity = *updated



    return nil

}



definition.DeleteOutput = definition.Output
```

`DeleteScope` restricts default read, list, count, update, and delete operations. `Delete` receives the entity after the scoped identity lookup. SQLH does not infer soft deletion from a field name.

## Custom Repository Settings[​](#custom-repository-settings "Direct link to Custom Repository Settings")

SQLH uses a transaction-aware `sqlr.RepositoryTx` by default. Use options when the handler needs another SQL client, repository settings, or a wrapper:

* `sqlh.WithClientName[K, E](name)` selects a configured SQL client.
* `sqlh.WithRepositorySettings[K, E](settings)` sets SQLR settings.
* `sqlh.WithRepositoryTxFactory[K, E](factory)` replaces repository construction.

The custom factory receives the same `sqlc.Client` that SQLH uses to start transactions:

main.go

```
type ReportingAuthorRepository struct {

  delegate sqlr.RepositoryTx[int64, Author]

}



func NewReportingAuthorRepository(client sqlc.Client, settings sqlr.Settings) (sqlr.RepositoryTx[int64, Author], error) {

  repo, err := sqlr.NewRepositoryTxWithSettings[int64, Author](client, settings)

  if err != nil {

    return nil, err

  }



  return &ReportingAuthorRepository{delegate: repo}, nil

}



func (r *ReportingAuthorRepository) Create(tx sqlr.TTx, entity *Author, opts ...func(qb *sqlr.QueryBuilderCreate)) error {

  return r.delegate.Create(tx, entity, opts...)

}



func (r *ReportingAuthorRepository) Read(tx sqlr.TTx, id int64, opts ...func(qb *sqlr.QueryBuilderRead)) (*Author, error) {

  return r.delegate.Read(tx, id, opts...)

}



func (r *ReportingAuthorRepository) Query(tx sqlr.TTx, opts ...func(qb *sqlr.QueryBuilderSelect)) ([]Author, error) {

  opts = append(opts, func(qb *sqlr.QueryBuilderSelect) {

    qb.OrderBy("created_at DESC")

    qb.Limit(100)

  })



  return r.delegate.Query(tx, opts...)

}



func (r *ReportingAuthorRepository) Count(tx sqlr.TTx, qb *sqlr.QueryBuilderSelect) (int, error) {

  return r.delegate.Count(tx, qb)

}



func (r *ReportingAuthorRepository) Update(tx sqlr.TTx, entity *Author, opts ...func(qb *sqlr.QueryBuilderUpdate)) (*Author, error) {

  return r.delegate.Update(tx, entity, opts...)

}



func (r *ReportingAuthorRepository) Delete(tx sqlr.TTx, id int64, opts ...func(qb *sqlr.QueryBuilderDelete)) error {

  return r.delegate.Delete(tx, id, opts...)

}



func (r *ReportingAuthorRepository) Close() error {

  return r.delegate.Close()

}
```

main.go

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

A custom repository must implement the transaction-aware `sqlr.RepositoryTx` interface. The interface includes `Count` in addition to `Create`, `Read`, `Query`, `Update`, `Delete`, and `Close`.

## Custom Operations[​](#custom-operations "Direct link to Custom Operations")

A `CrudDefinition` can replace a complete operation when the default flow does not match the domain. Operation callbacks receive the active `sqlr.TTx`, the configured `sqlr.RepositoryTx`, and the typed input:

```
type CrudOperation[K sqlr.KeyTypes, E sqlr.Entitier[K], I, O any] func(

    context.Context,

    sqlr.TTx,

    sqlr.RepositoryTx[K, E],

    *I,

) (O, error)
```

The definition supports `CreateOperation`, `ReadOperation`, `UpdateOperation`, `PatchOperation`, `ListOperation`, and `DeleteOperation`. An operation field replaces the complete default operation. `UpdateOperation` does not change the default PATCH pipeline. `PatchOperation` replaces the complete patch pipeline.

Use custom operations for domain rules that need direct control over the transaction. Keep authorization, validation, status transitions, event publication, and response mapping in application callbacks.

## Migration Notes[​](#migration-notes "Direct link to Migration Notes")

The SQLH CRUD redesign removes these APIs:

* `Transformer` and `JsonResultsTransformer`
* `WithTx`
* `BindTx`, `BindTxN`, `BindTxR`, and `BindTxNR`
* legacy Gin transaction binding files and builder-aware transformer interfaces

Use these replacements:

| Old pattern                            | New pattern                                                                   |
| -------------------------------------- | ----------------------------------------------------------------------------- |
| Transformer factory                    | `CrudDefinitionFactory` with `NewCrudDefinition` or a custom `CrudDefinition` |
| `RenderEntityResponse`                 | Typed `CrudDefinition.Output`                                                 |
| `GetInput` and request body assertions | Typed callback input values                                                   |
| Request-scoped `WithTx` middleware     | Default SQLH operation transactions or `TxRunner`                             |
| Repository factory                     | `WithRepositoryTxFactory` returning `sqlr.RepositoryTx`                       |
| Generic query list handler             | `ListInput`, `ListOutput`, `Query`, and `Count`                               |
| Implicit patch association behavior    | `sqlh` relation tags and `PatchAssociationTriggers`                           |

Add parity tests before changing the implementation. Check routes, response bodies, status codes, authorization, association behavior, filters, pagination, and transaction rollback behavior.

## See Also[​](#see-also "Direct link to See Also")

* [sqlc - SQL Client](/docs/pr-4/how-to/databases-sql/sqlc/.md)
* [sqlr - SQL Repository](/docs/pr-4/how-to/databases-sql/sqlr/.md)
* [Migrate an existing service to SQLH](/docs/pr-4/migrations/sqlh-crud/.md)
