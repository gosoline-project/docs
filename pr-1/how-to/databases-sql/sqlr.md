# sqlr - SQL Repository

The `sqlr` package provides a generic, type-safe repository layer built on top of [`sqlc`](/docs/pr-1/how-to/databases-sql/sqlc.md). It offers CRUD operations, relationship management, eager loading via joins and preloads, and transaction support — all using Go generics for compile-time type safety.

## Getting Started[​](#getting-started "Direct link to Getting Started")

Add the dependency to your Go module:

```
go get github.com/gosoline-project/sqlr@v0.8.0
```

Then import the package in your Go code:

```
import "github.com/gosoline-project/sqlr"
```

## Configuration[​](#configuration "Direct link to Configuration")

The `sqlr` package uses `sqlc` under the hood for database connections. Configure your database using the same `sqlc` configuration key described in the [sqlc documentation](/docs/pr-1/how-to/databases-sql/sqlc.md#configuration):

config.dist.yml

```
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

      path: migrations
```

## Defining Entities[​](#defining-entities "Direct link to Defining Entities")

Entities are Go structs that map to database tables. Embed `sqlr.Entity[K]` to get a primary key (`Id`), and timestamp fields (`CreatedAt`, `UpdatedAt`) that `sqlr` manages automatically by default:

main.go

```
type Author struct {

  sqlr.Entity[int64]

  Name  string

  Email string

  Posts []Post // HasMany

}



type Post struct {

  sqlr.Entity[int64]

  AuthorID int64

  Title    string

  Body     string

  Status   string

  Author   Author `sqlr:"belongsTo:author_id"` // BelongsTo

  Tags     []Tag  `sqlr:"many2many:"`          // ManyToMany

}



type Tag struct {

  sqlr.Entity[int64]

  Name string

}
```

### Entity Base Struct[​](#entity-base-struct "Direct link to Entity Base Struct")

The `sqlr.Entity[K]` base struct provides:

| Field       | Tag                                     | Description                                       |
| ----------- | --------------------------------------- | ------------------------------------------------- |
| `Id`        | `db:"id" sqlr:"primaryKey"`             | Primary key, auto-increment for integer types     |
| `CreatedAt` | `db:"created_at" sqlr:"autoCreateTime"` | Set automatically on insert by default            |
| `UpdatedAt` | `db:"updated_at" sqlr:"autoUpdateTime"` | Set automatically on insert and update by default |

All entities must implement the `Entitier[K]` interface (satisfied automatically by embedding `Entity[K]`):

```
type Entitier[K KeyTypes] interface {

    GetId() K

    GetUpdatedAt() time.Time

    GetCreatedAt() time.Time

}
```

### Struct Tags[​](#struct-tags "Direct link to Struct Tags")

Column names are controlled via the `db` struct tag, while `sqlr`-specific metadata is controlled via the `sqlr` struct tag. The `db` tag is **optional** for public fields — when omitted, the field name is converted to a column name using [`SchemaNameTransformer`](#schemanametransformer) (default: snake\_case). Public struct and slice-of-struct fields without a `db` or `sqlr` tag can also be auto-detected as relationships when `sqlr` finds matching entity and foreign-key evidence (see [Auto-detected Relationships](#auto-detected-relationships)). To explicitly exclude a public field from all mapping, use `db:"-"` by itself.

Within the `sqlr` tag, options are separated with semicolons. Grouped sync options may combine any of `create`, `update`, and `delete`, for example `sync:create,update` or `sync:update,delete`.

| Tag                                                     | Description                                                                                                                                                 |
| ------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `db:"column_name"`                                      | Maps the field to a database column                                                                                                                         |
| `db:"-"`                                                | Explicitly excludes a public field from column and relationship mapping                                                                                     |
| `sqlr:"primaryKey"`                                     | Marks the field as the primary key                                                                                                                          |
| `sqlr:"autoCreateTime"`                                 | Auto-sets the field to `time.Now()` on insert by default                                                                                                    |
| `sqlr:"autoUpdateTime"`                                 | Auto-sets the field to `time.Now()` on insert and update by default                                                                                         |
| `sqlr:"foreignKey:col"`                                 | Defines a HasOne or HasMany relationship (see [Relationships](#relationships))                                                                              |
| `sqlr:"belongsTo:col"`                                  | Defines a BelongsTo relationship                                                                                                                            |
| `sqlr:"many2many:table"`                                | Defines a ManyToMany relationship                                                                                                                           |
| `sqlr:"many2many:table;parentKey:col"`                  | Overrides the join table column that references the parent entity's PK                                                                                      |
| `sqlr:"many2many:table;relatedKey:col"`                 | Overrides the join table column that references the related entity's PK                                                                                     |
| `sqlr:"...;preload"`                                    | Automatically preloads the relationship on every `Read()` and `Query()` call, and on `Create()` and `Update()` whenever `sqlr` performs a post-write reload |
| `sqlr:"...;sync:create"`                                | Synchronizes the relationship by default during `Create()`                                                                                                  |
| `sqlr:"...;sync:update"`                                | Synchronizes the relationship by default during `Update()`                                                                                                  |
| `sqlr:"...;sync:delete"`                                | Cleans up the relationship by default during `Delete()`                                                                                                     |
| `sqlr:"...;sync:create,update,delete"`                  | Applies the relationship by default during any listed operations                                                                                            |
| `sqlr:"many2many:table;sync:update;syncMode:many2many"` | Enables full related-row synchronization by default for a many-to-many relation during `Update()`                                                           |

### Excluding Fields[​](#excluding-fields "Direct link to Excluding Fields")

To exclude a public field from all mapping (no column, no relationship), use `db:"-"`:

```
type Author struct {

    sqlr.Entity[int64]

    Name     string `db:"name"`

    Internal string `db:"-"` // excluded from column and relationship mapping

}
```

### Table Name Derivation[​](#table-name-derivation "Direct link to Table Name Derivation")

Table names are automatically derived from the struct type name by applying [`SchemaNameTransformer`](#schemanametransformer) (default: snake\_case conversion) and then pluralizing:

| Struct Name | Table Name   |
| ----------- | ------------ |
| `Author`    | `authors`    |
| `Post`      | `posts`      |
| `PostTag`   | `posts_tags` |

To override the default table name, implement the `TableNamer` interface:

```
type TableNamer interface {

    TableName() string

}



func (a Author) TableName() string {

    return "my_authors"

}
```

### SchemaNameTransformer[​](#schemanametransformer "Direct link to SchemaNameTransformer")

`SchemaNameTransformer` is a package-level variable that controls how Go field names and type names are converted to database identifiers. It is used for:

* Deriving column names for untagged public fields
* Deriving table names from struct type names (before pluralization)
* Deriving foreign key column names for auto-detected relationships
* Deriving join table column names for ManyToMany relationships when no `parentKey:`/`relatedKey:` override is set

The default transformer is `toSnakeCase` (PascalCase/camelCase → snake\_case). You can replace it at program startup before any repository is created:

```
import (

    "strings"

    "github.com/gosoline-project/sqlr"

)



func init() {

    // Use lowercase field names instead of snake_case

    sqlr.SchemaNameTransformer = strings.ToLower

}
```

The transformer must be set **before** any repository or schema parsing occurs, as schemas are cached after first use.

### Supported Key Types[​](#supported-key-types "Direct link to Supported Key Types")

Primary keys can be any of these types (or their pointer variants):

```
bool | string | int | int64 | uint | uint64 | float32 | float64
```

Integer primary key types (`int`, `int64`, `uint`, `uint64`) are automatically treated as auto-increment by default — they are excluded from INSERT statements and their value is set from `LastInsertId()` after creation. If you disable auto updates for `Create()`, `sqlr` instead inserts the primary key value already present on the entity.

## Creating the Repository[​](#creating-the-repository "Direct link to Creating the Repository")

Create a repository using `sqlr.NewRepository[K, E]()` within a gosoline application. The type parameters specify the primary key type and the entity type:

main.go

```
authorRepo, err := sqlr.NewRepository[int64, Author](ctx, config, logger, "default")

if err != nil {

  return nil, fmt.Errorf("failed to create author repository: %w", err)

}
```

The last argument (`"default"` above) is the **sqlc client name**. It must match a key under the `sqlc` block in your configuration file — the repository uses that named client for all database operations. You can have multiple repositories pointing to different clients (e.g., `"default"`, `"analytics"`, `"readonly"`) by passing different names.

The `Repository` interface provides:

| Method                         | Description                                                                                                                                                                                                                                                                       |
| ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Create(ctx, entity, opts...)` | Inserts the entity and any eligible populated association fields; optional create builder methods can limit, omit, or disable association persistence for a call, and can request a post-create reload with preloads                                                              |
| `Read(ctx, id, opts...)`       | Loads one entity by primary key, with optional `QueryBuilderRead` relation loading                                                                                                                                                                                                |
| `Query(ctx, opts...)`          | Loads entities matching query conditions                                                                                                                                                                                                                                          |
| `Update(ctx, entity, opts...)` | Updates only the base entity row by default; associations remain untouched unless explicitly synchronized per call or via `sync:update` tags, many-to-many updates stay link-only unless `SyncMany2many()` is used, and `Preload()` can request a post-update reload of relations |
| `Delete(ctx, id, opts...)`     | Deletes the entity by primary key; by default also cleans up owned associations, and delete builder methods can narrow or disable that cleanup                                                                                                                                    |
| `Close()`                      | Releases resources (prepared statements, etc.)                                                                                                                                                                                                                                    |

## Relationships[​](#relationships "Direct link to Relationships")

Relationships can be declared in two ways: via **auto-detection** of eligible untagged public fields (convention over configuration) or via **explicit `sqlr` struct tags** (full control over every name). Relationship fields may optionally also use `db:"-"`, but `db:"-"` is not required when relationship metadata is already present in `sqlr`.

### Auto-detected Relationships[​](#auto-detected-relationships "Direct link to Auto-detected Relationships")

When `sqlr` parses a public field without a `db` tag, it first checks whether the field qualifies for relationship auto-detection:

* Only public struct fields and slices of structs are candidates.
* The related type must declare a primary key (directly or through an embedded field).
* A public non-slice struct field can auto-detect as `BelongsTo` only when the inferred FK column `SchemaNameTransformer(fieldName) + "_id"` exists on the parent entity.
* A public slice-of-struct field can auto-detect as `HasMany` only when the inferred FK column `SchemaNameTransformer(parentTypeName) + "_id"` exists on the related entity.
* If these checks fail, the field is treated like any other untagged field and mapped using `SchemaNameTransformer`.
* `HasOne` and `ManyToMany` are never auto-detected and always require explicit tags.

### HasOne[​](#hasone "Direct link to HasOne")

The foreign key lives on the **related** table and the field type is a single struct. HasOne cannot be auto-detected — untagged non-slice struct fields can only auto-detect as `BelongsTo`, never `HasOne`. An explicit `foreignKey:` tag is always required.

#### Explicit Tag[​](#explicit-tag "Direct link to Explicit Tag")

```
type Author struct {

    sqlr.Entity[int64]

    Name    string  `db:"name"`

    Profile Profile `sqlr:"foreignKey:author_id"`

}



type Profile struct {

    sqlr.Entity[int64]

    AuthorID int64  `db:"author_id"`

    Bio      string `db:"bio"`

}
```

Table schema

```
CREATE TABLE authors (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    name       VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

);



CREATE TABLE profiles (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    author_id  BIGINT NOT NULL,

    bio        TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (author_id) REFERENCES authors(id)

);
```

### HasMany[​](#hasmany "Direct link to HasMany")

The foreign key lives on the **related** table and the field type is a slice.

#### Auto-detected[​](#auto-detected "Direct link to Auto-detected")

A public slice-of-struct field with no `db` tag is automatically treated as `HasMany` only when the related type declares a primary key and the inferred foreign key column exists on the related table. The foreign key column name is derived as `SchemaNameTransformer(parentTypeName) + "_id"`.

```
type Author struct {

    sqlr.Entity[int64]

    Name  string

    Posts []Post // FK "author_id" derived on posts table

}



type Post struct {

    sqlr.Entity[int64]

    AuthorID int64

    Title    string

}
```

#### Explicit Tag[​](#explicit-tag-1 "Direct link to Explicit Tag")

```
type Author struct {

    sqlr.Entity[int64]

    Name  string `db:"name"`

    Posts []Post `sqlr:"foreignKey:author_id"`

}



type Post struct {

    sqlr.Entity[int64]

    AuthorID int64  `db:"author_id"`

    Title    string `db:"title"`

}
```

Table schema

```
CREATE TABLE authors (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    name       VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

);



CREATE TABLE posts (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    author_id  BIGINT NOT NULL,

    title      VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (author_id) REFERENCES authors(id)

);
```

### BelongsTo[​](#belongsto "Direct link to BelongsTo")

The foreign key lives on the **current** entity's table.

#### Auto-detected[​](#auto-detected-1 "Direct link to Auto-detected")

A public non-slice struct field with no `db` tag is automatically treated as `BelongsTo` only when the related type declares a primary key and the inferred foreign key column exists on the current table. The foreign key column name is derived as `SchemaNameTransformer(fieldName) + "_id"`.

```
type Post struct {

    sqlr.Entity[int64]

    AuthorID int64

    Title    string

    Author   Author // FK "author_id" derived on this table

}



type Author struct {

    sqlr.Entity[int64]

    Name string

}
```

#### Explicit Tag[​](#explicit-tag-2 "Direct link to Explicit Tag")

```
type Post struct {

    sqlr.Entity[int64]

    AuthorID int64  `db:"author_id"`

    Title    string `db:"title"`

    Author   Author `sqlr:"belongsTo:author_id"`

}



type Author struct {

    sqlr.Entity[int64]

    Name string `db:"name"`

}
```

Table schema

The database schema is the same as HasMany — BelongsTo is the inverse perspective of the same foreign key column:

```
CREATE TABLE authors (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    name       VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

);



CREATE TABLE posts (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    author_id  BIGINT NOT NULL,

    title      VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (author_id) REFERENCES authors(id)

);
```

### ManyToMany[​](#manytomany "Direct link to ManyToMany")

A join table is required and `many2many:` always needs an explicit `sqlr` tag — ManyToMany relationships cannot be auto-detected.

The join table must have columns referencing the primary key of each side. By default these column names are derived as `SchemaNameTransformer(EntityType) + "_id"`.

#### Auto-derived Join Table Name[​](#auto-derived-join-table-name "Direct link to Auto-derived Join Table Name")

Leave the `many2many:` value empty to have the join table name derived automatically from both entity table names, sorted alphabetically and joined with an underscore:

```
type Post struct {

    sqlr.Entity[int64]

    Title string

    Tags  []Tag `sqlr:"many2many:"` // join table "posts_tags" auto-derived

}



type Tag struct {

    sqlr.Entity[int64]

    Name string

}
```

#### Explicit Tag[​](#explicit-tag-3 "Direct link to Explicit Tag")

```
type Post struct {

    sqlr.Entity[int64]

    Title string `db:"title"`

    Tags  []Tag  `sqlr:"many2many:posts_tags"`

}



type Tag struct {

    sqlr.Entity[int64]

    Name string `db:"name"`

}
```

Table schema

Both the auto-derived and explicit variants map to the same schema. The join table column names are derived as `SchemaNameTransformer(EntityType) + "_id"`:

```
CREATE TABLE posts (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    title      VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

);



CREATE TABLE tags (

    id         BIGINT AUTO_INCREMENT PRIMARY KEY,

    name       VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

);



CREATE TABLE posts_tags (

    post_id BIGINT NOT NULL,

    tag_id  BIGINT NOT NULL,

    PRIMARY KEY (post_id, tag_id),

    FOREIGN KEY (post_id) REFERENCES posts(id),

    FOREIGN KEY (tag_id)  REFERENCES tags(id)

);
```

#### Overriding Join Table Column Names[​](#overriding-join-table-column-names "Direct link to Overriding Join Table Column Names")

Use `parentKey:` and `relatedKey:` to override the join table column names when they differ from the default convention:

```
type Article struct {

    sqlr.Entity[int64]

    Title  string  `db:"title"`

    Labels []Label `sqlr:"many2many:article_labels;parentKey:art_id;relatedKey:lbl_id"`

}



type Label struct {

    sqlr.Entity[int64]

    Name string `db:"name"`

}
```

Table schema

The corresponding join table uses abbreviated column names that differ from the default `SchemaNameTransformer` convention (`article_id` / `label_id`):

```
CREATE TABLE article_labels (

    art_id BIGINT NOT NULL,

    lbl_id BIGINT NOT NULL,

    PRIMARY KEY (art_id, lbl_id),

    FOREIGN KEY (art_id) REFERENCES articles(id),

    FOREIGN KEY (lbl_id) REFERENCES labels(id)

);
```

| Option           | Description                                                        |
| ---------------- | ------------------------------------------------------------------ |
| `parentKey:col`  | Join table column referencing the **parent** entity's primary key  |
| `relatedKey:col` | Join table column referencing the **related** entity's primary key |

## Association Path Semantics[​](#association-path-semantics "Direct link to Association Path Semantics")

`sqlr` uses dot-separated association paths such as `Posts.Comments` anywhere an API needs to identify relationships on an entity graph. That includes association persistence on `Create()`, association synchronization on `Update()`, association cleanup on `Delete()`, and relation loading through `Preload()` on `Read()`, `Query()`, and the post-write reloads available on `Create()` and `Update()`.

Paths are based on the entity schema's Go struct relationship field names, not on table names, column names, foreign-key column names, or `db` tag values.

* Use `Posts`, `Profile`, `Author`, or `Tags` when those are the relationship field names on your structs.
* Use `Posts.Comments` to follow the `Posts` field on the root entity and then the `Comments` field on `Post`.
* Do not use names such as `posts`, `author_id`, or database table names in relation paths.

```
qb.Preload("Posts.Comments")

qb.LeftJoin("Profile")

qb.SyncAssociation("Posts.Comments")

qb.OmitAssociation("Profile")
```

* Paths use dot notation for nested relations, such as `Posts.Comments`.
* `Preload()` on `Create()`, `Read()`, `Query()`, and `Update()` post-reloads accepts direct and nested relation paths.
* Join methods such as `LeftJoin()` and `InnerJoin()` accept direct relation field names only; nested paths and many-to-many loading use `Preload()`.
* For `Create()` and `Update()`, selecting a nested path implicitly traverses the required ancestors, so `Posts.Comments` includes both `Posts` and `Comments`.
* Schema-level defaults from `sync:create`, `sync:update`, and `sync:delete` follow the same path resolution and can target nested paths even when an intermediate parent relation is itself untagged.
* Omitting a parent path such as `Posts` also omits all of its descendants.
* `Delete()` uses the selected path to reach the owned branch and then recursively cleans up that selected branch according to delete semantics.
* `SyncMany2many(paths...)` only accepts paths whose terminal relation is many-to-many.
* Invalid association paths return an error.

## CRUD Operations[​](#crud-operations "Direct link to CRUD Operations")

### Create[​](#create "Direct link to Create")

Pass a pointer to an entity. By default, auto-increment IDs and timestamp fields are set automatically:

main.go

```
func (s *BlogService) createAuthor(ctx context.Context, name, email string) (*Author, error) {

  author := &Author{

    Name:  name,

    Email: email,

  }



  if err := s.authorRepo.Create(ctx, author); err != nil {

    return nil, fmt.Errorf("failed to create author: %w", err)

  }



  // author.Id is automatically set for auto-increment primary keys

  // author.CreatedAt and author.UpdatedAt are set via autoCreateTime/autoUpdateTime



  return author, nil

}
```

#### Preserve preset IDs and timestamps[​](#preserve-preset-ids-and-timestamps "Direct link to Preserve preset IDs and timestamps")

Use `DisableAutoUpdates()` when you want `Create()` to persist the primary key and timestamp values already present on the entity instead of letting `sqlr` generate them:

```
err := repo.Create(ctx, &author, func(qb *sqlr.QueryBuilderCreate) {

    qb.DisableAutoUpdates()

})
```

* `Create()` includes the entity's current primary key and timestamp values in the `INSERT`.
* This also applies across the selected association graph when creating related entities.
* If a created entity has a zero primary key while auto updates are disabled, `Create()` returns `ErrAutoUpdatesRequirePresetPrimaryKey`.

### Create with Associations[​](#create-with-associations "Direct link to Create with Associations")

By default, populate relationship fields on the entity before calling `Create()` and `sqlr` will automatically persist them in the correct order within a single transaction:

main.go

```
func (s *BlogService) createPost(ctx context.Context, authorId int64, title, body, status string, tags ...Tag) (*Post, error) {

  post := &Post{

    AuthorID: authorId,

    Title:    title,

    Body:     body,

    Status:   status,

    Tags:     tags,

  }



  // When Tags is populated, sqlr automatically inserts the tag rows and

  // the posts_tags join table entries within a single transaction.

  if err := s.postRepo.Create(ctx, post); err != nil {

    return nil, fmt.Errorf("failed to create post: %w", err)

  }



  return post, nil

}
```

The association save order is:

1. **BelongsTo** — related entity inserted first so the parent's FK column is set before the parent row is written.
2. **Parent entity** — the base row is inserted.
3. **HasOne / HasMany** — related entities are inserted with their FK pointing to the parent PK.
4. **ManyToMany** — related entities with zero PKs are inserted, then join table rows are created for all of them.

Entities with a non-zero primary key are treated as already-persisted: they are skipped for insertion, but the FK or join table row is still created.

When `DisableAutoUpdates()` is enabled for `Create()`, `sqlr` also preserves preset IDs and timestamps across associations. Foreign keys are reconciled against the associated primary key: empty FK fields are filled automatically, while conflicting preset FK values return an error instead of being overwritten.

If no relationship path in the schema is tagged with `sync:create`, `Create()` persists all eligible populated relationships by default. As soon as at least one path is tagged with `sync:create`, the default behavior narrows to the tagged paths; nested tagged paths are still honored even when an intermediate parent relation is untagged, and `SyncAssociation()` / `OmitAssociation()` still let you refine that selection per call. `OmitAllAssociations()` disables association persistence for the call entirely, even if explicit sync paths or schema-level `sync:create` defaults would otherwise select relations.

When `Create()` persists associations and the schema defines `preload` tags, `sqlr` reloads the created entity before returning so those auto-preloaded relations are hydrated.

#### Create with Selective Association Persistence[​](#create-with-selective-association-persistence "Direct link to Create with Selective Association Persistence")

Use `QueryBuilderCreate` to limit or omit association paths for a single `Create()` call:

```
err := repo.Create(ctx, &author, func(qb *sqlr.QueryBuilderCreate) {

    qb.SyncAssociation("Posts.Comments")

    qb.OmitAssociation("Profile")

})
```

* `SyncAssociation(paths...)` persists only the listed association paths.
* `OmitAssociation(paths...)` skips the listed paths and all of their descendants.
* `OmitAllAssociations()` skips all association persistence and inserts only the root row.
* `DisableAutoUpdates()` can be combined with these options when you want to preserve preset IDs and timestamps during the same `Create()` call.
* Invalid association paths return an error.
* If no association remains eligible after applying the options, `Create()` behaves like a plain parent insert.

#### Create with Post-Create Preloading[​](#create-with-post-create-preloading "Direct link to Create with Post-Create Preloading")

Use `QueryBuilderCreate` when you want `Create()` to reload the entity by primary key after persistence and hydrate relations before returning:

```
err := repo.Create(ctx, &post, func(qb *sqlr.QueryBuilderCreate) {

    qb.Preload("Author")

    qb.Preload("Tags", sqlr.Condition("name != ?", ""))

})
```

* `Preload()` on `Create()` uses the same association-path syntax as `Read()`, `Query()`, and `Update()`.
* Conditions and nested paths are supported.
* `sqlr` performs the insert first, then reloads the entity and applies the requested preloads.
* When explicit create preloads overlap with schema auto-preloads, the explicit create preloads take precedence.
* Invalid preload paths return an error.

### Read[​](#read "Direct link to Read")

Load a single entity by its primary key:

main.go

```
func (s *BlogService) readAuthor(ctx context.Context, id int64) (*Author, error) {

  author, err := s.authorRepo.Read(ctx, id)

  if err != nil {

    return nil, fmt.Errorf("failed to read author: %w", err)

  }



  return author, nil

}
```

### Read with Association Loading[​](#read-with-association-loading "Direct link to Read with Association Loading")

Use `QueryBuilderRead` to control association loading for a single `Read()` call:

```
author, err := repo.Read(ctx, id, func(qb *sqlr.QueryBuilderRead) {

    qb.Preload("Posts.Comments")

    qb.LeftJoin("Profile")

})
```

* `Read()` always loads the base entity by primary key; `QueryBuilderRead` only controls eager loading for that call.
* Use `Preload()` and join methods such as `LeftJoin()` and `InnerJoin()` to load related entities together with the base row.
* Schema auto-preloads defined with the `preload` tag option are still applied and merged with any explicit preloads you add.
* See [Eager Loading with Preload](#eager-loading-with-preload) and [Eager Loading with Joins](#eager-loading-with-joins) below for supported relation types, nested paths, conditions, and limitations.

### Update[​](#update "Direct link to Update")

Modify the entity and pass it to `Update()`. By default, only the base row is updated and the `UpdatedAt` timestamp is refreshed automatically. Any populated association fields on the entity are ignored unless association sync is explicitly enabled for that call or activated by `sync:update` tags on relationship fields:

main.go

```
func (s *BlogService) updatePostStatus(ctx context.Context, post *Post, status string) (*Post, error) {

  post.Status = status



  updated, err := s.postRepo.Update(ctx, post)

  if err != nil {

    return nil, fmt.Errorf("failed to update post: %w", err)

  }



  // updated.UpdatedAt is automatically refreshed via autoUpdateTime



  return updated, nil

}
```

#### Preserve preset timestamps during update[​](#preserve-preset-timestamps-during-update "Direct link to Preserve preset timestamps during update")

Use `DisableAutoUpdates()` when you want `Update()` to persist the values already present on the entity instead of assigning fresh `autoUpdateTime` timestamps:

```
updated, err := repo.Update(ctx, &author, func(qb *sqlr.QueryBuilderUpdate) {

    qb.DisableAutoUpdates()

})
```

* `Update()` writes the entity's current column values as-is, including preset timestamp fields.
* When association sync is also enabled, the same behavior applies across the synchronized association graph.

### Update with Association Sync[​](#update-with-association-sync "Direct link to Update with Association Sync")

If you want `Update()` to also insert, update, delete, or unlink related records, opt in per call with `QueryBuilderUpdate` or declare schema-level defaults with `sync:update` tags. Existing many-to-many related rows are a special case: by default, `Update()` only reconciles join-table membership for them and does not update their columns unless you explicitly opt that path into full entity synchronization. When association sync is enabled and the schema defines `preload` tags, `sqlr` reloads the updated entity before returning so those auto-preloaded relations are hydrated.

Use relationship tags when you want synchronization to happen by default without repeating query-builder calls:

```
type Author struct {

    sqlr.Entity[int64]

    Name  string `db:"name"`

    Posts []Post `sqlr:"foreignKey:author_id;sync:create,update"`

}



type Article struct {

    sqlr.Entity[int64]

    Title string `db:"title"`

    Tags  []Tag  `sqlr:"many2many:article_tags;sync:update;syncMode:many2many"`

}
```

* `sync:update` enables default association sync for the tagged path during `Update()`.
* Nested `sync:update` paths behave like explicit path selection: `sqlr` still traverses required ancestors even when an intermediate parent relation is untagged.
* `syncMode:many2many` is only valid on many-to-many relations and must be combined with `sync:update`.

Enable full association synchronization with `SyncAllAssociations()`:

```
updated, err := repo.Update(ctx, &author, func(qb *sqlr.QueryBuilderUpdate) {

    qb.SyncAllAssociations()

})
```

Sync only selected association paths with `SyncAssociation(paths...)`:

```
updated, err := repo.Update(ctx, &author, func(qb *sqlr.QueryBuilderUpdate) {

    qb.SyncAssociation("Profile", "Posts.Comments")

})
```

Opt selected many-to-many paths into full related-row synchronization with `SyncMany2many(paths...)`:

```
updated, err := repo.Update(ctx, &article, func(qb *sqlr.QueryBuilderUpdate) {

    qb.SyncAllAssociations()

    qb.SyncMany2many("Tags")

})
```

Combine `SyncAllAssociations()` with `OmitAssociation(paths...)` to exclude specific paths from a full sync:

```
updated, err := repo.Update(ctx, &author, func(qb *sqlr.QueryBuilderUpdate) {

    qb.SyncAllAssociations()

    qb.OmitAssociation("Posts.Tags")

})
```

When association sync is enabled, `sqlr` synchronizes the selected association graph in a single transaction. With `SyncAllAssociations()`, that means the full entity graph:

1. **BelongsTo** - related entities are inserted or updated first so the parent's FK column is current before the parent row is written.
2. **Parent entity** - the base row is updated.
3. **HasOne** - the related entity is inserted or updated with its FK pointing to the parent; previously stored replacement rows are deleted.
4. **HasMany** - listed related entities are inserted or updated with their FK pointing to the parent; previously stored children missing from the slice are deleted.
5. **ManyToMany** - related entities with zero PKs are inserted; existing related rows are link-synchronized by default, so missing join-table links are removed and new ones are inserted without updating the related rows themselves unless that path was opted into `SyncMany2many()`.

If `DisableAutoUpdates()` is also enabled, `sqlr` preserves preset IDs and timestamps throughout the synchronized graph. Foreign keys are reconciled against the associated primary key: empty FK fields are filled automatically, while conflicting preset FK values return an error and roll back the transaction.

See [Association Path Semantics](#association-path-semantics) for the shared path syntax, naming rules, and traversal behavior used by `Update()`.

Association sync only acts on relationships that are selected for synchronization and explicitly present on the entity passed to `Update()`:

* A zero-value struct relation is treated as untouched.
* A `nil` slice relation is treated as untouched.
* An empty slice means "synchronize this relation to no rows" and removes existing HasMany or ManyToMany links.

If the parent row does not exist, `Update()` with association synchronization enabled returns `ErrNotFound` and rolls back the transaction. `ErrNotFound` is also returned when default many-to-many synchronization tries to link an existing related row by primary key but that related row does not exist.

#### Update with Post-Update Preloading[​](#update-with-post-update-preloading "Direct link to Update with Post-Update Preloading")

Use `QueryBuilderUpdate` when you want `Update()` to reload the entity by primary key after persistence and hydrate relations before returning:

main.go

```
func (s *BlogService) updatePostStatusWithRelations(ctx context.Context, post *Post, status string) (*Post, error) {

  post.Status = status



  updated, err := s.postRepo.Update(ctx, post, func(qb *sqlr.QueryBuilderUpdate) {

    qb.Preload("Author")

    qb.Preload("Tags", sqlr.Condition(sqlc.Col("name").NotEq("")))

  })

  if err != nil {

    return nil, fmt.Errorf("failed to update post with relations: %w", err)

  }



  // updated.Author and updated.Tags are hydrated by the post-update reload.



  return updated, nil

}
```

* `Preload()` on `Update()` uses the same association-path syntax as `Read()` and `Query()`.
* Conditions and nested paths are supported.
* `sqlr` performs the update first, then reloads the entity and applies the requested preloads.
* Invalid preload paths return an error.
* When explicit update preloads overlap with schema auto-preloads, the explicit update preloads take precedence.

### Delete[​](#delete "Direct link to Delete")

Remove an entity by its primary key:

main.go

```
func (s *BlogService) deleteTag(ctx context.Context, id int64) error {

  if err := s.tagRepo.Delete(ctx, id); err != nil {

    return fmt.Errorf("failed to delete tag: %w", err)

  }



  return nil

}
```

### Delete with Association Cleanup[​](#delete-with-association-cleanup "Direct link to Delete with Association Cleanup")

By default, `Delete()` cleans up owned associations before deleting the root row:

* `HasOne` and `HasMany` relations are recursively deleted.
* `ManyToMany` relations remove join-table rows only.
* `BelongsTo` relations are left untouched.

If no relationship path in the schema is tagged with `sync:delete`, `Delete()` cascades all owned relations by default. As soon as at least one path is tagged with `sync:delete`, the default cleanup narrows to the tagged paths. Nested tagged paths are still honored even when an intermediate parent relation is untagged; `Delete()` traverses the required ancestor branch to reach the selected owned subtree and then recursively cleans up that subtree.

Use `QueryBuilderDelete` to refine association cleanup for a single `Delete()` call:

```
err := repo.Delete(ctx, id, func(qb *sqlr.QueryBuilderDelete) {

    qb.SyncAssociation("Posts.Comments")

    qb.OmitAssociation("Profile")

})
```

* `SyncAssociation(paths...)` restricts cleanup to the listed owned association paths.
* `OmitAssociation(paths...)` skips the listed paths and all of their descendants.
* `OmitAllAssociations()` disables owned-association cleanup entirely and deletes only the root row.
* For `Delete()`, a nested selected path such as `Posts.Comments` traverses `Posts` to reach `Comments`, then recursively cleans up the selected owned branch.
* Invalid association paths return an error.

## Query Operations[​](#query-operations "Direct link to Query Operations")

Use `Query()` with a `QueryBuilderSelect` to filter, sort, and paginate results:

main.go

```
func (s *BlogService) queryPublishedPosts(ctx context.Context) ([]Post, error) {

  posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {

    qb.Where(sqlc.Col("status").Eq("published")).

      OrderBy("created_at DESC").

      Limit(10).

      Offset(0)

  })

  if err != nil {

    return nil, fmt.Errorf("failed to query posts: %w", err)

  }



  return posts, nil

}
```

The `QueryBuilderSelect` supports:

| Method                         | Description                                       |
| ------------------------------ | ------------------------------------------------- |
| `Where(condition, params...)`  | Adds a WHERE condition (multiple calls are ANDed) |
| `OrderBy(cols...)`             | Sets the ORDER BY clause                          |
| `Limit(n)`                     | Limits the number of results                      |
| `Offset(n)`                    | Skips the first n results                         |
| `GroupBy(cols...)`             | Sets the GROUP BY clause                          |
| `Having(condition, params...)` | Adds a HAVING condition                           |

WHERE conditions use the same `sqlc.Col()` expression API from the [`sqlc` package](/docs/pr-1/how-to/databases-sql/sqlc.md):

```
qb.Where(sqlc.Col("status").Eq("published"))

qb.Where(sqlc.Col("age").Gt(18))

qb.Where(sqlc.And(sqlc.Col("a").Eq(1), sqlc.Col("b").Eq(2)))
```

## Eager Loading with Preload[​](#eager-loading-with-preload "Direct link to Eager Loading with Preload")

Use `Preload()` to load related entities in separate queries. On `Read()` and `Query()`, preloads run as part of the lookup; on `Create()` and `Update()`, preloads run in a follow-up reload after the write succeeds. Preloads support all relationship types: HasOne, HasMany, BelongsTo, and ManyToMany.

main.go

```
func (s *BlogService) queryPostsWithAuthor(ctx context.Context) ([]Post, error) {

  posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {

    qb.Where(sqlc.Col("status").Eq("published")).

      Preload("Author").

      Preload("Tags")

  })

  if err != nil {

    return nil, fmt.Errorf("failed to query posts: %w", err)

  }



  // Each post now has post.Author and post.Tags populated



  return posts, nil

}
```

### Nested Preloads[​](#nested-preloads "Direct link to Nested Preloads")

Load nested relationships using dot-separated paths. Conditions on nested paths apply to the leaf relation only:

```
qb.Preload("Posts.Comments")

qb.Preload("Posts.Comments", sqlr.Condition("body != ?", ""))
```

### Auto-Preload[​](#auto-preload "Direct link to Auto-Preload")

Add the `preload` tag option to a relationship to automatically load it on every `Read()` and `Query()` call, and on `Create()` and `Update()` whenever `sqlr` performs a post-write reload — without requiring an explicit `Preload()` call:

```
type PostWithPreloads struct {

    sqlr.Entity[int64]

    AuthorID int64  `db:"author_id"`

    Title    string `db:"title"`

    Author   Author `sqlr:"belongsTo:author_id;preload"`

    Tags     []Tag  `sqlr:"many2many:posts_tags;preload"`

}
```

Auto-preloads are recursively discovered across nested relationships and merged with any explicit preloads, including `QueryBuilderCreate.Preload()` and `QueryBuilderUpdate.Preload()` (explicit preloads take precedence when both are present).

During `Create()`, `sqlr` performs that post-create reload when you request `QueryBuilderCreate.Preload()` explicitly or when `Create()` synchronizes associations and the schema defines auto-preloads.

### Preload Conditions[​](#preload-conditions "Direct link to Preload Conditions")

Pass conditions to `Preload()` to filter which related entities are loaded:

main.go

```
func (s *BlogService) queryPostsPreloadWithCondition(ctx context.Context) ([]Post, error) {

  posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {

    qb.Preload("Author").Preload("Tags", sqlr.Condition(sqlc.Col("name").NotEq("")))

  })

  if err != nil {

    return nil, fmt.Errorf("failed to query posts: %w", err)

  }



  return posts, nil

}
```

## Eager Loading with Joins[​](#eager-loading-with-joins "Direct link to Eager Loading with Joins")

Use `LeftJoin()` or `InnerJoin()` to load related entities via SQL JOINs. Joins support direct HasOne, HasMany, and BelongsTo relationships (ManyToMany and nested paths require `Preload`).

### Joins on Query[​](#joins-on-query "Direct link to Joins on Query")

main.go

```
func (s *BlogService) queryPostsWithAuthorJoin(ctx context.Context) ([]Post, error) {

  posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {

    qb.Where(sqlc.Col("status").Eq("published")).

      LeftJoin("Author")

  })

  if err != nil {

    return nil, fmt.Errorf("failed to query posts: %w", err)

  }



  // Each post now has post.Author populated via a SQL JOIN



  return posts, nil

}
```

### Joins on Read[​](#joins-on-read "Direct link to Joins on Read")

Joins can also be used with `Read()` to load relations alongside a single entity lookup:

main.go

```
func (s *BlogService) readPostWithAuthor(ctx context.Context, postId int64) (*Post, error) {

  post, err := s.postRepo.Read(ctx, postId, func(qb *sqlr.QueryBuilderRead) {

    qb.LeftJoin("Author")

  })

  if err != nil {

    return nil, fmt.Errorf("failed to read post with author: %w", err)

  }



  return post, nil

}
```

### Join Conditions[​](#join-conditions "Direct link to Join Conditions")

Pass conditions to restrict the joined rows:

main.go

```
func (s *BlogService) queryPostsJoinWithCondition(ctx context.Context) ([]Post, error) {

  posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {

    qb.LeftJoin("Author", sqlr.Condition(sqlc.Col("name").Eq("Alice")))

  })

  if err != nil {

    return nil, fmt.Errorf("failed to query posts: %w", err)

  }



  return posts, nil

}
```

## Error Handling[​](#error-handling "Direct link to Error Handling")

The `sqlr.ErrNotFound` sentinel error is returned when `Read()` cannot find the requested entity, when plain `Delete()` cannot delete the requested row, when association-aware `Delete()` cannot load the root entity selected for cleanup, when `Update()` with association synchronization enabled cannot find the parent row, and when default many-to-many synchronization references a related entity by primary key that does not exist:

```
author, err := repo.Read(ctx, id)

if errors.Is(err, sqlr.ErrNotFound) {

    // entity does not exist

}
```

When `DisableAutoUpdates()` is used with `Create()`, `sqlr.ErrAutoUpdatesRequirePresetPrimaryKey` is returned if the entity being inserted does not already have a primary key value:

```
err := repo.Create(ctx, &author, func(qb *sqlr.QueryBuilderCreate) {

    qb.DisableAutoUpdates()

})

if errors.Is(err, sqlr.ErrAutoUpdatesRequirePresetPrimaryKey) {

    // provide a primary key before creating the entity

}
```

## Fixtures and Test Data[​](#fixtures-and-test-data "Direct link to Fixtures and Test Data")

You can also use `sqlr` to load typed database fixtures through the gosoline `fixtures` package instead of maintaining raw SQL seed migrations.

Use `FixtureEntity()` when you want fixture values with explicit primary keys and timestamps:

integration\_test.go

```
type Author struct {

  sqlr.Entity[int64]

  Name string `db:"name"`

}



type Post struct {

  sqlr.Entity[int64]

  AuthorID int64  `db:"author_id"`

  Title    string `db:"title"`

  Author   Author `sqlr:"belongsTo:author_id"`

}
```

integration\_test.go

```
var authors = fixtures.NamedFixtures[Author]{

  fixtures.NewNamedFixture("author_1", Author{

    Entity: sqlr.FixtureEntity[int64](1, "2024-01-01T09:00:00Z", "2024-01-01T09:00:00Z"),

    Name:   "Alice Johnson",

  }),

}



var posts = fixtures.NamedFixtures[Post]{

  fixtures.NewNamedFixture("post_with_author", Post{

    Entity:   sqlr.FixtureEntity[int64](10, "2024-01-05T10:00:00Z", "2024-01-05T10:00:00Z"),

    AuthorID: 1,

    Title:    "Seeded with sqlr fixtures",

  }),

}
```

Use `FixtureSetFactory()` to turn named fixtures into a gosoline fixture set factory backed by an `sqlr` repository:

integration\_test.go

```
func Fixtures() fixtures.FixtureSetsFactory {

  return fixtures.NewFixtureSetsFactory(

    sqlr.FixtureSetFactory[int64, Author](authors),

    sqlr.FixtureSetFactory[int64, Post](posts),

  )

}
```

Register that fixture set factory in your test suite or application setup:

integration\_test.go

```
func (s *FixtureIntegrationTestSuite) SetupSuite() []suite.Option {

  return []suite.Option{

    suite.WithConfigFile("config.dist.yml"),

    suite.WithFixtureSetFactory(Fixtures()),

  }

}



func (s *FixtureIntegrationTestSuite) SetupTest() error {

  s.ctx = s.Env().Context()



  repo, err := sqlr.NewRepository[int64, Post](s.ctx, s.Env().Config(), s.Env().Logger(), "default")

  if err != nil {

    return fmt.Errorf("failed to create post repository: %w", err)

  }



  s.repo = repo



  return nil

}
```

Then read the seeded entity in your test and assert that the preset IDs, timestamps, and associations were written as expected:

integration\_test.go

```
func (s *FixtureIntegrationTestSuite) TestReadSeededPost() {

  post, err := s.repo.Read(s.ctx, 10, func(qb *sqlr.QueryBuilderRead) {

    qb.Preload("Author")

  })

  s.Require().NoError(err)



  s.Equal(&Post{

    Entity: sqlr.Entity[int64]{

      Id:        10,

      CreatedAt: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),

      UpdatedAt: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),

    },

    AuthorID: 1,

    Title:    "Seeded with sqlr fixtures",

    Author: Author{

      Entity: sqlr.Entity[int64]{

        Id:        1,

        CreatedAt: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),

        UpdatedAt: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),

      },

      Name: "Alice Johnson",

    },

  }, post)

}
```

Fixture writes preserve the explicit IDs and timestamps already present on your entities. Internally, `sqlr` creates fixtures with auto updates disabled, so seeded values are inserted as-is instead of being regenerated.

Associations follow the normal `Create()` behavior described above. If your fixtures embed related entities, `sqlr` persists them in the usual association order. If you provide conflicting preset foreign keys, fixture loading fails with the same validation errors as a regular `Create()` call.

`FixtureSetFactory()` is the recommended high-level API. For advanced scenarios, `NewFixtureWriter()` and `NewFixtureWriterWithInterfaces()` are also available when you want to wire the writer manually.

The built-in `sqlr` fixture writer uses the `default` sqlc client name. If you define multiple fixture sets without embedding them as associations, make sure their factory order still respects your foreign-key dependencies.

For more details about gosoline fixture loading, grouping, purge behavior, and named fixture references, see the [fixtures package reference](/docs/pr-1/reference/package-fixtures.md).
