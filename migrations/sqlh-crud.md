# Migrating an Existing Service to SQLH CRUD

This guide covers migrations from SQLH v0.7.x and from custom CRUD handlers. Use it to preserve the service's public API and domain behavior.

SQLH replaces repeated CRUD handlers with typed callbacks, SQLR entities, and transaction-aware operations. Do not replace domain rules with generic CRUD behavior.

This guide targets SQLH `v0.8.2-0.20261027132918-7e9d41caf9c9`, SQLR `v0.9.2-0.20261027132740-484c58e94d7b`, SQLC v0.4.1, and httpserver v0.6.4. This SQLH version declares Go 1.27.0 and depends on Gosoline v0.63.5; a compatible newer Gosoline version can remain selected.

## 1. Review the Current API[​](#1-review-the-current-api "Direct link to 1. Review the Current API")

Before changing code, record the current behavior:

* HTTP paths, methods, middleware, and authorization
* Request and response schemas, envelopes, and status codes
* `Content-Type`, `Accept`, and response negotiation behavior
* Validation errors, not-found behavior, and soft-delete rules
* Association create, update, patch, and delete behavior
* List filters, joins, ordering, grouping, pagination, and counts
* Event publication, transaction boundaries, and rollback behavior

Create a parity matrix for every CRUD endpoint. Add parity checks before removing old handlers. Search for old SQLH symbols, custom handlers, repository factories, relation mappers, soft-delete scopes, generated artifacts, and event or outbox paths. For v0.7.1, inspect `Transformer`, `JsonResultsTransformer`, every `Builder*Aware` hook, `WithTx`/`BindTx*`, and each `WithRepositoryFactory` or returned wrapper. Record its side effects, including database provisioning, data copies, and storage cleanup.

## 2. Replace the Persistence Model[​](#2-replace-the-persistence-model "Direct link to 2. Replace the Persistence Model")

Embed `sqlr.Entity[K]` when SQLR should manage IDs and timestamps. Otherwise implement `sqlr.Entitier[K]` and preserve the service's ID and timestamp behavior. Keep table names, column names, primary keys, and nullable fields, and implement `TableName()` when SQLR's inferred name differs.

Use `db` for columns and `db:"-"` for non-column relation fields:

```
type Project struct {

    sqlr.Entity[uint]



    OwnerID *uint  `db:"owner_id"`

    Owner   *Owner `db:"-" sqlr:"belongsTo:owner_id;preload" sqlh:"preload:read,update"`

    Rules   []*Rule `db:"-" sqlr:"foreignKey:project_id;preload;sync:create,update" sqlh:"preload:create,read,query,update;sync:create,update"`

    Labels  []*Label `db:"-" sqlr:"foreignKey:project_id;preload;sync:create,update" sqlh:"preload:create,read,query,update;sync:create,update"`

    Tags    []*Tag  `db:"-" sqlr:"many2many:project_tags;sync:update" sqlh:"preload:read,update;sync:update"`

}
```

Use SQLR tags for relationship shape and schema defaults. `foreignKey:<column>` names the foreign-key column on a HasOne or HasMany child table. `belongsTo:<column>` names the foreign-key column on the current entity and requires a matching `db` field. `many2many:<table>` names the join table.

Use `parentKey:<column>` and `relatedKey:<column>` for nonstandard many-to-many join-column names. Other relevant options are `primaryKey`, `autoCreateTime`, `autoUpdateTime`, `preload`, `sync:create,update,delete`, and `syncMode:many2many`. Separate SQLR options with semicolons. `syncMode:many2many` requires `sync:update` and a many-to-many relation.

Use SQLH relation tags for CRUD-phase builders: `preload:create,read,query,update` and `sync:create,update,delete`. Separate directives with semicolons and phases with commas. Relation paths use exact Go field names such as `Rules` or `Rules.Owner`, not database columns or JSON names. Scalar fields support `sqlh:"field:alias"` or `sqlh:"field:-"` for automatic query naming or exclusion, not preload/sync. Unknown directives, unsupported phases, empty or duplicate `field` directives, and directives on embedded fields fail during handler setup.

`sqlr:"...;preload"` is a schema-level auto-preload. `sqlh:"preload:<phases>"` adds phase-specific builders, and `preload:update` affects the update lookup and update write, not only post-write loading. SQLH parses nested relation tags recursively.

Treat sync tags as policy. SQLR Create and Delete default to all association paths, but the first create or delete sync path changes that operation to selected-path mode. SQLR Update defaults to no association sync, then merges schema defaults and per-operation paths. Omit options take precedence. SQLR `sync:update` can synchronize PUT relations without an SQLH tag, while default PATCH association selection still needs matching SQLH `sync:update` policy.

## 3. Build the SQLR Repository[​](#3-build-the-sqlr-repository "Direct link to 3. Build the SQLR Repository")

Construct the transaction-aware repository from the same SQLC client that SQLH uses:

```
client, err := sqlc.ProvideClient(ctx, config, logger, "default")

if err != nil {

    return nil, err

}



repository, err := sqlr.NewRepositoryTxWithSettings[uint, model.Project](client, sqlr.DefaultSettings())
```

Keep domain queries and background transitions in an application repository. Apply an active-row scope to normal reads when the service uses soft deletion, and provide a separate method for consumers that must read deleted rows. Use explicit transactions for background transitions and `ForUpdate()` when a transition must lock the root row. Before replacing a v0.7.x `WithRepositoryFactory`, inspect its repository wrapper and method side effects. The v0.8 `WithRepositoryTxFactory` supplies a transaction-aware SQLR repository, but it does not migrate those effects. Preserve database provisioning, data copies, and storage cleanup at the service or route boundary when their order or failure behavior matters. SQLH transactions cannot roll back blob-store or other external effects, and MySQL DDL may commit implicitly. Keep legacy Create/Delete handlers outside the SQLH CRUD handler when that order matters. Use `NewCrudHandler` to manually register only the other SQLH routes.

## 4. Define Typed SQLH Callbacks[​](#4-define-typed-sqlh-callbacks "Direct link to 4. Define Typed SQLH Callbacks")

Create an update input that embeds `sqlh.InputById[Id]`. It supplies `GetId`, force-filter storage, and the `uri:"id" json:"-"` route identity. Keep request fields tagged for JSON and Gin validation.

After authorization, use `AddForceFilter` to attach tenant or account restrictions. Each force filter must add a restrictive `WHERE` clause only.

```
type ProjectUpdateInput struct {

    sqlh.InputById[string]

    Name      string       `json:"name" binding:"required"`

    RuleItems []*RuleInput `json:"ruleItems"`

}
```

### Choose the definition constructor[​](#choose-the-definition-constructor "Direct link to Choose the definition constructor")

For the standard `sqlh.ListInput`, call `NewCrudDefinition` and wrap the result:

```
definition := sqlh.NewCrudDefinition(

    operations.createInput,

    operations.updateInput,

    operations.patchInputFromEntity,

    operations.output,

)

definition.Identity = identityByPublicID

definition.DeleteScope = activeScope

definition.Delete = operations.delete

definition.DeleteOutput = definition.Output

definition.PatchAssociationTriggers = map[string]string{"state": "Labels"}



factory := sqlh.SimpleCrudDefinition(definition)
```

`NewCrudDefinition` fixes the list input type to `sqlh.ListInput`. For a custom list input, instantiate `CrudDefinition` with every type parameter; do not assign `NewCrudDefinition` to a custom-LI definition.

```
type ProjectListInput struct {

    sqlh.ListInput

    LegacyFilter string `json:"legacyFilter,omitempty"`

}



definition := sqlh.CrudDefinition[

    uint, Project, string, ProjectCreateInput, ProjectUpdateInput, ProjectListInput, ProjectOutput,

]{

    CreateInput:          operations.createInput,

    UpdateInput:          operations.updateInput,

    PatchInputFromEntity: operations.patchInputFromEntity,

    Output:               operations.output,

}

factory := sqlh.SimpleCrudDefinition(definition)
```

Use a `CrudDefinitionFactory` instead of `SimpleCrudDefinition` when construction needs context, configuration, a logger, or another startup dependency. Keep domain behavior in callbacks: `CreateInput` authorizes and validates the request, maps root and child values, and preserves ID and state rules.

`UpdateInput` authorizes the loaded entity, validates before mutation, maps complete values, resolves child IDs, creates ID-less children, applies state transitions, and preserves timestamp ownership. `PatchInputFromEntity` copies every writable value that an omitted patch field must retain, including child public IDs. `Output` checks read authorization and maps the existing typed response. Map v0.7.x `TransformCreateInput` to `CreateInput`, `TransformUpdateInput` to `UpdateInput`, and `TransformOutput` or `RenderEntityResponse` to `Output`. Map each old `RenderQueryResponse` result with `Output`. Keep a route wrapper when the old list envelope differs from `ListOutput[O]`. SQLH v0.7.1 had no built-in PATCH route or patch mapper. If the old API has PATCH, preserve it with a custom operation or a complete-input mapper. Move each old `Builder*Aware` hook to relation tags, list or identity callbacks, or a full operation. Preserve its query, preload, and synchronization behavior. The mapper callbacks do not receive a transaction or repository. Use a `*Operation` callback for database reads or writes that need SQLH's active transaction. When an old transformer returned an explicit `httpserver.Response`, record its status, headers, and content type. Typed outputs use httpserver negotiation and can return 406 for an unsupported `Accept`. When JSON is the only registered representation and negotiated responses are acceptable, bind typed output directly. Do not add a JSON-only wrapper. Use an explicit `httpserver.Response` or route wrapper when the old response must bypass negotiation.

Default HTTP binding validates the bound type. PUT validates the update input, but PATCH validates `sqlh.PatchInput`, not the merged update input. Validate the complete input at the start of `UpdateInput`, before changing the entity:

```
import (

    "github.com/gin-gonic/gin/binding"

    "github.com/justtrackio/gosoline/pkg/validation"

)



if err := binding.Validator.ValidateStruct(input); err != nil {

    return nil, validation.NewError(err)

}
```

This reuses `binding` tags and registered Gin field or struct validators. `validation.NewError` maps the failure to HTTP 400. A custom `UpdateOperation` must perform this validation itself.

## 5. Preserve JSON Merge Patch Behavior[​](#5-preserve-json-merge-patch-behavior "Direct link to 5. Preserve JSON Merge Patch Behavior")

The default PATCH operation runs in one transaction:

1. Load the scoped entity with update preloads and a root-row lock.
2. Build the complete update input.
3. Merge the original RFC 7396 document into that input.
4. Call `UpdateInput`.
5. Select associations present in the original document.
6. Persist the root and selected associations.
7. Map output and commit.

Do not treat `UpdateInput` as partial. The complete input must contain every value that an omitted field keeps, and arrays replace the complete array.

An explicit null clears a pointer or nullable value. Use a pointer or nullable wrapper for fields that accept null. For a non-nullable scalar, JSON null becomes the field's zero value during merge. Required or domain validation must reject it when the old contract forbids null.

An explicit null or empty selected collection clears that collection. A selected HasOne null clears its owned child. For a selected, directly mapped BelongsTo path, SQLH also clears its nullable foreign key when PATCH sets it to null. For PUT or trigger-driven changes, the mapper must clear the foreign key because SQLR skips a nil target. An omitted relation is not synchronized merely because the complete input contains its current value.

For direct child relations, an existing child ID updates that child and an ID-less child creates one. A missing HasMany child can be deleted when the selected sync path requires it. For many-to-many relations, an existing ID is link-only by default; SQLR verifies the target and reconciles join-table membership without updating that target row.

An ID-less many-to-many target is created. A missing target removes its link without deleting the target row. Use full target-row sync only when the old contract permits mutation of shared targets. Configure it with SQLR `sync:update;syncMode:many2many` and matching SQLH `sync:update`.

SQLH derives a direct association path from update-input `json` tags and entity relation names. Use overrides when the paths differ:

```
definition.PatchAssociations = map[string]string{

    "ruleItems": "Rules",

}

definition.PatchAssociationTriggers = map[string]string{

    "state": "Labels",

}
```

A direct path or trigger must name a relation with SQLH `sync:update`. A trigger only selects the relation when the scalar JSON path is present; it does not compute the relation or change merge and null semantics. Keep selection based on the original patch document.

A custom operation that bypasses the default builder must request the same association path. Use `SyncMany2many` for an operation-level full-sync choice.

Use a custom `PatchOperation` only when the complete-input pipeline cannot represent the old API. It replaces this pipeline inside SQLH's transaction, and `TxRunner` still wraps the operation. Use the supplied transaction and repository, and implement validation, association selection, persistence, and output mapping.

In SQLH v0.8.0, custom PATCH operations read the parsed document from `input.Document`. Do not call `input.Document()`.

## 6. Preserve List Compatibility[​](#6-preserve-list-compatibility "Direct link to 6. Preserve List Compatibility")

Use the `ProjectListInput` type above for SQLH public-field filters, ordering, pagination, and force filters. Embedding `sqlh.ListInput` supplies the required `ListInputSource` methods:

* `GetFilter() sqlh.JsonFilter`
* `GetOrderBy() []sqlh.JsonOrder`
* `ApplyPagination`
* `ValidatePagination`
* `ValidateQueryModifiers() error`
* `GetForceFilters`

A custom `ValidateQueryModifiers` method can return `nil` when no additional validation is required.

Migrate public filters from `column` to `field` and ordering to `{"field":"name","direction":"asc"}` objects in `orderBy`. `sqlh.JsonFilter.Type` remains `sqlc.JsonFilterType`; nested `Fields` use `sqlh.JsonFilter` too. SQLH validates every leaf and sort field before callbacks or database work, including custom list operations. Unknown names, raw identifiers/expressions, and legacy `column` keys are rejected.

With `CrudDefinition.FieldMappings == nil`, eligible directly mapped scalar entity fields use exported Go names with anonymous-field promotion and Go shadowing rules, not output DTO names or database columns. Binding tags (`json`, `xml`, `form`, `yaml`, etc.) are ignored. Use `sqlh:"field:alias"` only when a different automatic public name is needed, or `sqlh:"field:-"` to exclude a field.

**Migration warning:** `json:"-"` no longer excludes a field from automatic query discovery. Add `sqlh:"field:-"` or select the complete public allowlist with `FieldMappings`. JSON renames that differ from Go names by more than case also need a SQLH alias or explicit mapping.

Public names match case-insensitively with Go's Unicode `strings.EqualFold` semantics, without punctuation/style normalization: `createdAt` matches `CreatedAt`, not `created_at`. Go paths and database identifiers stay exact. Active public names that collide under case folding fail handler construction with the conflicting names and Go paths; excluded fields do not participate.

Set `FieldMappings` to a complete public-name-to-Go-path allowlist, for example `map[string]string{"id": "PublicID", "ownerName": "Owner.Name"}`. Non-nil replaces automatic discovery without merging names or aliases; an empty map exposes none. Explicit entries can select a mapped scalar even if its automatic name is aliased or excluded. Explicit relation paths support one direct `BelongsTo` or `HasOne` hop to a scalar; SQLH adds a `LEFT JOIN`. To-many/deeper paths are unsupported, and invalid declarations fail handler construction.

If `LegacyFilter` represents an additional typed field, optional `ApplyFilters` can add its trusted backend predicate. Optional `ApplyQueryModifiers` can add trusted grouping or ordering. These are not native public-query processors; do not delegate to removed `ListInput.ApplyFilters` or `ListInput.ApplyQueryModifiers` methods or forward request-provided SQL. Force filters remain separate trusted server-owned restrictions.

Do not add pagination if the old API did not expose it. A custom `ListInputSource` can retain the intended fields, provide empty public-filter/order getters when absent, and use no-op pagination and pagination validation. Instantiate `CrudDefinition` with every type parameter; `NewCrudDefinition` fixes the input type to `sqlh.ListInput`. Preserve intended unknown-field behavior outside the public filter/order grammar.

Custom `Query` and `Count` callbacks must call `QueryPlan.ApplyBuilder` and then `QueryPlan.ApplyScope`. Return a scope error before repository work. A custom `Query` then applies modifiers and pagination, in that order. A custom `Count` omits row modifiers, ordering, and pagination. The default count shares the builder and scope, including public filters, trusted filters, force filters, and delete scope. It excludes row modifiers, ordering, and pagination. If modifiers change row cardinality, implement a custom count with the corresponding cardinality logic. Keep the same builder and scope for rows and counts.

`CrudHandler.List` and `CrudDefinition.ListOperation` always return `ListOutput[O]` with `Results` and `Total`. If the old API uses another outer envelope, call `handler.List` from a typed wrapper and map those fields, then register the wrapper manually:

```
func listLegacy(ctx context.Context, input *ProjectListInput) (LegacyListOutput, error) {

    result, err := handler.List(ctx, input)

    if err != nil {

        return LegacyListOutput{}, err

    }



    return LegacyListOutput{Items: result.Results, Count: result.Total}, nil

}



router.POST("/v1/projects", httpserver.Bind(listLegacy))
```

Do not make `ListOperation` return another envelope. For a legacy filter dialect, detect it before decoding and reject unsafe mixed native and legacy forms. Translate only known fields and operators to `sqlh.JsonFilter`/`sqlh.JsonOrder` public names so they pass through the same allowlist, rather than exposing database identifiers.

Preserve required joins, grouping, ordering, pagination, and client-dependent `LIMIT 0` behavior. Reject unknown fields, operators, directions, and boolean values. Apply the same scope to row and count queries.

## 7. Keep Existing Routes and Delete Responses[​](#7-keep-existing-routes-and-delete-responses "Direct link to 7. Keep Existing Routes and Delete Responses")

`WithCrudHandlers` registers:

* `POST /v{version}/{entity}`
* `GET /v{version}/{entity}/:id`
* `PUT /v{version}/{entity}/:id`
* `PATCH /v{version}/{entity}/:id`
* `DELETE /v{version}/{entity}/:id`
* `POST /v{version}/{plural-entity}`

`WithCrudHandlers` in SQLH v0.7.1 did not register a PATCH route. SQLH v0.8.0 adds that route. Both versions return 204 for the standard DELETE route. Use `NewCrudHandler` and manual registration when a path, method, middleware, binder, response envelope, or status differs. Do not register PATCH unless it is part of the public API.

Use `DeleteScope` to hide inactive rows and a custom `Delete` callback for soft deletion or other domain rules. The scope applies to default read, list, count, update, PATCH, and delete lookups. When DELETE returns a body, set `definition.DeleteOutput = definition.Output` and manually bind `handler.Delete`. `DeleteNoContent` never calls `DeleteOutput`.

Set `Identity` when a public ID uses a non-primary-key column or when public `Id` and stored key `K` have different Go types. SQLH rejects differing `Id` and `K` types when `Identity` is nil. Use the default lookup only when the types and key semantics match. In a custom lookup, use the supplied repository and transaction. Add the public-ID predicate. For every lookup, apply the supplied scope, return its error before repository work, and apply the supplied builder. Return `sqlr.ErrNotFound` when no row matches. Ordinary GET lookups use the builder for relation preloads, but not for row locking. Default Update, PATCH, and Delete lookups also use it for `FOR UPDATE`. SQLR locks the root and preloaded rows in the same transaction, including nested and many-to-many preloads.

## 8. Preserve Transactions, Background Work, and Events[​](#8-preserve-transactions-background-work-and-events "Direct link to 8. Preserve Transactions, Background Work, and Events")

Build `RepositoryTx` from the same `sqlc.Client` that SQLH uses to begin transactions. Use `WithRepositoryTxFactory` and `sqlr.NewRepositoryTxWithSettings` for custom repository construction. The default SQLH operation maps output before commit. It rolls back its SQL transaction on operation, association, output, and panic errors. Each custom `*Operation` runs inside SQLH's `TxRunner` and receives its active `sqlr.TTx` and configured repository. Use them for transaction-bound database work; the transaction cannot roll back blob-store or other external effects, and MySQL DDL may commit implicitly. An override owns authorization, scope, root locking, association selection, and persistence for its operation. Remove old `WithTx`/`BindTx*` middleware from migrated CRUD routes. SQLH v0.8.0 starts a transaction for each CRUD operation. Do not wrap it in another transaction or put a transaction in Gin context. `UpdateOperation` does not replace the mapper used by default PATCH. `PatchOperation` replaces the default PATCH pipeline inside the existing transaction. `ListOperation` still returns `ListOutput[O]`. Use `TxRunner.InTransaction` for custom typed handlers outside a CRUD definition. Keep background transitions in the application repository with explicit transactions. Use `ForUpdate()` when a transition must serialize with another writer. Preserve existing retry, idempotency, and reader visibility behavior.

Publish events only after the database transaction commits. For CDC, publish from the committed change event and reload the aggregate with the required active or deleted-row visibility. Preserve soft-delete events, hard-delete tombstones, outbox boundaries, and retry behavior.

Preserve the existing SQLC configuration, migration path, and reset behavior. Set MySQL `clientFoundRows=true` under SQLC `parameters` when a valid unchanged update must count as matched. Without it, SQLR can report `ErrNotFound` for an update that matched a row.

## 9. Validate and Regenerate[​](#9-validate-and-regenerate "Direct link to 9. Validate and Regenerate")

Before removing old handlers, verify route and response parity, authorization, validation and 400 errors, JSON Merge Patch omission and null behavior, association preloads and synchronization, native and legacy filters, row/count scope parity, soft-delete visibility, delete responses, not-found behavior, rollback after root or child failures, background locks, and post-commit events.

Regenerate Go types, mocks, OpenAPI files, and other derived artifacts from the preserved source definitions. Do not edit generated files as the primary fix. Remove old handlers and removed SQLH APIs only after parity checks pass.

For the full operational procedure, copy the instructions below. The prompt is self-contained and includes pinned API semantics, tag rules, constructors, operation choices, acceptance checks, and stop conditions.

Copy AI migration instructions

## See Also[​](#see-also "Direct link to See Also")

* [sqlh - SQL HTTP Handlers](/docs/how-to/databases-sql/sqlh/.md)
* [sqlr - SQL Repository](/docs/how-to/databases-sql/sqlr/.md)
* [sqlc - SQL Client](/docs/how-to/databases-sql/sqlc/.md)
