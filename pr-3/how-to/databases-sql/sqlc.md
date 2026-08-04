# sqlc - SQL Client

The `sqlc` package provides a fluent API for building and executing SQL queries against relational databases. It supports MySQL and PostgreSQL, and offers connection management, query building, and transaction handling.

## Getting Started[​](#getting-started "Direct link to Getting Started")

Add the dependency to your Go module:

```
go get github.com/gosoline-project/sqlc@v0.3.0
```

Then import the package in your Go code:

```
import "github.com/gosoline-project/sqlc"
```

## Configuration[​](#configuration "Direct link to Configuration")

A gosoline application can have **multiple `sqlc` clients**, each identified by a unique name (e.g., `default`, `analytics`, `readonly`). Every client you intend to use **must be configured before use** - attempting to create a client or DB handle without a matching configuration entry returns an error when `sqlc.NewClient()`, `sqlc.ProvideClient()`, or `sqlc.NewDB()` is called.

Configure database connections under the `sqlc` key in your configuration file, with each connection name as a sub-key:

```
sqlc:

  default:       # primary database

    # ...

  analytics:     # separate analytics database

    # ...
```

Each connection requires driver and connection details as shown below:

config.dist.yml

```
sqlc:

  default:

    driver: mysql

    collation: utf8mb4_general_ci

    parse_time: true

    multi_statements: true

    uri:

      host: 127.0.0.1

      port: 3306

      user: root

      password: gosoline

      database: blog

    max_open_connections: 25

    max_idle_connections: 5

    connection_max_lifetime: 5m

    connection_max_idletime: 2m

    migrations:

      enabled: true

      application: "{app.name}"

      path: migrations
```

### Connection Settings[​](#connection-settings "Direct link to Connection Settings")

| Setting                   | Description                                                                        | Default              |
| ------------------------- | ---------------------------------------------------------------------------------- | -------------------- |
| `driver`                  | Database driver (`mysql` or `postgres`)                                            | Required             |
| `uri.host`                | Database host                                                                      | `localhost`          |
| `uri.port`                | Database port. Set `5432` explicitly for PostgreSQL.                               | `3306`               |
| `uri.user`                | Database user                                                                      | Required             |
| `uri.password`            | Database password                                                                  | Required             |
| `uri.database`            | Database name                                                                      | Required             |
| `charset`                 | MySQL driver character set                                                         | `utf8mb4`            |
| `collation`               | MySQL driver collation                                                             | `utf8mb4_general_ci` |
| `parse_time`              | MySQL driver option to parse date/time columns into Go time values                 | `true`               |
| `multi_statements`        | MySQL driver option to allow multiple SQL statements per query                     | `true`               |
| `parameters`              | Additional driver-specific DSN parameters                                          | None                 |
| `max_open_connections`    | Maximum open connections                                                           | `0` (unlimited)      |
| `max_idle_connections`    | Maximum idle connections                                                           | `2`                  |
| `connection_max_lifetime` | Maximum connection lifetime                                                        | `120s`               |
| `connection_max_idletime` | Maximum idle time                                                                  | `120s`               |
| `retry.enabled`           | Retry failed database operations                                                   | `false`              |
| `timeouts.readTimeout`    | MySQL driver I/O read timeout                                                      | `0`                  |
| `timeouts.writeTimeout`   | MySQL driver I/O write timeout                                                     | `0`                  |
| `timeouts.timeout`        | MySQL driver connection timeout. For PostgreSQL, use `parameters.connect_timeout`. | `0`                  |

Use `parameters` for driver-specific settings such as PostgreSQL `sslmode` or `connect_timeout`. PostgreSQL connections use `uri.*` and `parameters`; the MySQL-specific settings above are ignored by the PostgreSQL driver.

## Migrations[​](#migrations "Direct link to Migrations")

The `sqlc` package can automatically run database migrations when the client is created. It uses [goose](https://github.com/pressly/goose) as the default migration provider.

### Enabling Migrations[​](#enabling-migrations "Direct link to Enabling Migrations")

Add migration settings to your database configuration:

```
sqlc:

  default:

    migrations:

      enabled: true

      path: migrations
```

### Migration Settings[​](#migration-settings "Direct link to Migration Settings")

| Setting           | Description                                                                                                                              | Default      |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| `enabled`         | Run migrations automatically                                                                                                             | `false`      |
| `application`     | Reserved for migration providers that support application-specific table prefixes. The built-in goose runner does not currently use it.  | `{app.name}` |
| `path`            | Path to migration files. Set this when `enabled` is `true`; if empty, migrations are skipped.                                            | None         |
| `provider`        | Migration provider                                                                                                                       | `goose`      |
| `reset`           | Reset the database before migrations. The built-in implementation issues MySQL `DROP DATABASE`, `CREATE DATABASE`, and `USE` statements. | `false`      |
| `prefixed_tables` | Reserved for migration providers that support table prefixing. The built-in goose runner does not currently use it.                      | `false`      |

### Migration Files[​](#migration-files "Direct link to Migration Files")

Create migration files in the specified path using the goose format:

```
-- +goose Up

CREATE TABLE authors (

    id BIGINT AUTO_INCREMENT PRIMARY KEY,

    name VARCHAR(255) NOT NULL,

    email VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP

);



-- +goose Down

DROP TABLE authors;
```

The `reset: true` option is useful for local MySQL development - it drops and recreates the database before running migrations, ensuring a clean state.

## Creating the Client[​](#creating-the-client "Direct link to Creating the Client")

Create a client using `sqlc.NewClient()` within a gosoline application. The client reads configuration from the specified connection name.

main.go

```
var err error

var client sqlc.Client



if client, err = sqlc.NewClient(ctx, config, logger, "default"); err != nil {

  return nil, fmt.Errorf("failed to create sqlc client: %w", err)

}
```

The client provides:

**Query Methods (from Querier interface):**

* `Get(ctx, dest, query, args...)` - Executes a query returning at most one row, scans into dest
* `Select(ctx, dest, query, args...)` - Executes a query and scans all rows into a slice
* `Query(ctx, query, args...)` - Executes a query returning `*sqlc.Rows` for iteration
* `QueryRow(ctx, query, args...)` - Executes a query returning at most one row
* `Exec(ctx, query, args...)` - Executes a query without returning rows (INSERT, UPDATE, DELETE)
* `NamedExec(ctx, query, arg)` - Executes a named query using `:name` syntax from struct/map
* `Prepare(ctx, query)` - Creates a `*sqlc.Stmt` for later execution

**Transaction Methods:**

* `WithTx(ctx, fn, opts...)` - Executes a function within a transaction (auto commit/rollback)
* `BeginTx(ctx, opts...)` - Starts a new transaction manually

**Other Methods:**

* `Q()` - Returns a QueryBuilder for constructing SQL queries
* `Close()` - Closes the database connection

## Data Models[​](#data-models "Direct link to Data Models")

Define struct types that map to your database tables using `db` struct tags:

main.go

```
type Author struct {

  Id        int64     `db:"id"`

  Name      string    `db:"name"`

  Email     string    `db:"email"`

  CreatedAt time.Time `db:"created_at"`

  UpdatedAt time.Time `db:"updated_at"`

}
```

The `db` tag specifies the column name. Create composite structs for join results by embedding base types and adding additional fields.

## INSERT Operations[​](#insert-operations "Direct link to INSERT Operations")

### Single Record[​](#single-record "Direct link to Single Record")

Use `Into()` to create an INSERT builder, then `Records()` to pass a struct:

main.go

```
func (s *BlogService) createAuthor(ctx context.Context, name, email string) (*Author, error) {

  author := &Author{

    Name:  name,

    Email: email,

  }



  result, err := s.client.Q().Into("authors").Records(author).Exec(ctx)

  if err != nil {

    return nil, fmt.Errorf("failed to insert author: %w", err)

  }



  id, err := result.LastInsertId()

  if err != nil {

    return nil, fmt.Errorf("failed to get last insert id: %w", err)

  }



  author.Id = id



  return author, nil

}
```

The `Exec()` method returns a `Result` with `LastInsertId()` and `RowsAffected()`.

### Bulk Insert[​](#bulk-insert "Direct link to Bulk Insert")

Pass a slice of structs to `Records()` for bulk insertion:

main.go

```
func (s *BlogService) createTags(ctx context.Context, names []string) ([]Tag, error) {

  tags := make([]Tag, len(names))

  for i, name := range names {

    tags[i] = Tag{Name: name}

  }



  _, err := s.client.Q().Into("tags").Records(tags).Exec(ctx)

  if err != nil {

    return nil, fmt.Errorf("failed to insert tags: %w", err)

  }



  return tags, nil

}
```

## Query Operations[​](#query-operations "Direct link to Query Operations")

### Simple Queries[​](#simple-queries "Direct link to Simple Queries")

Use `From()` to create a SELECT builder. Chain methods like `Where()`, `OrderBy()`, and `Limit()`:

main.go

```
func (s *BlogService) queryPostsByAuthor(ctx context.Context, authorId int64) ([]Post, error) {

  var posts []Post



  err := sqlc.From("posts").

    WithClient(s.client).

    Where(sqlc.Col("author_id").Eq(authorId)).

    OrderBy("created_at DESC").

    Select(ctx, &posts)

  if err != nil {

    return nil, fmt.Errorf("failed to query posts: %w", err)

  }



  return posts, nil

}
```

### Queries with JOINs[​](#queries-with-joins "Direct link to Queries with JOINs")

Build conditional joins using `LeftJoin()`, `InnerJoin()`, `RightJoin()`, and `FullOuterJoin()`. These methods return a `JoinBuilder` that must be finalized with `On()`. `CrossJoin()` and `Natural*Join()` return the select builder directly and do not take `On()`:

main.go

```
func (s *BlogService) queryPostsWithJoins(ctx context.Context) ([]PostWithAuthor, error) {

  var results []PostWithAuthor



  err := sqlc.From("posts").As("p").

    Columns("p.id", "p.author_id", "p.title", "p.body", "p.status", "p.created_at", "p.updated_at").

    LeftJoin("authors").As("a").On("p.author_id = a.id").

    Column(sqlc.Col("a.name").As("author_name")).

    Column(sqlc.Col("a.email").As("author_email")).

    Where(sqlc.Col("p.status").Eq("published")).

    OrderBy("p.created_at DESC").

    WithClient(s.client).

    Select(ctx, &results)

  if err != nil {

    return nil, fmt.Errorf("failed to query posts with joins: %w", err)

  }



  return results, nil

}
```

Use `As()` for table aliases and column aliases. `Columns()` replaces the current projection list, while `Column()` appends one more projected column.

### Streaming Rows[​](#streaming-rows "Direct link to Streaming Rows")

Use `Query()` when you want to iterate row-by-row. The returned `*sqlc.Rows` supports both primitive `Scan()` calls and `StructScan()` for structs:

main.go

```
posts := make([]Post, 0)

for rows.Next() {

  var post Post

  if err := rows.StructScan(&post); err != nil {

    return nil, fmt.Errorf("failed to scan post: %w", err)

  }



  posts = append(posts, post)

}



if err := rows.Err(); err != nil {

  return nil, fmt.Errorf("row iteration failed: %w", err)

}
```

## UPDATE Operations[​](#update-operations "Direct link to UPDATE Operations")

Use `Update()` to create an UPDATE builder. Chain `Set()` for column values and `Where()` for conditions:

main.go

```
func (s *BlogService) updatePostStatus(ctx context.Context, postId int64, status string) (*Post, error) {

  result, err := sqlc.Update("posts").

    WithClient(s.client).

    Set("status", status).

    Where(sqlc.Col("id").Eq(postId)).

    Exec(ctx)

  if err != nil {

    return nil, fmt.Errorf("failed to update post: %w", err)

  }



  rowsAffected, err := result.RowsAffected()

  if err != nil {

    return nil, fmt.Errorf("failed to get rows affected: %w", err)

  }



  if rowsAffected == 0 {

    return nil, fmt.Errorf("no post found with id %d", postId)

  }



  var post Post

  err = sqlc.From("posts").

    WithClient(s.client).

    Where(sqlc.Col("id").Eq(postId)).

    Get(ctx, &post)

  if err != nil {

    return nil, fmt.Errorf("failed to get updated post: %w", err)

  }



  return &post, nil

}
```

Multiple `Set()` calls are combined in the UPDATE clause. Use `SetExpr()` for SQL expressions:

```
sqlc.Update("posts").

    SetExpr("updated_at", "NOW()").

    Set("status", "published")
```

## Named Parameters[​](#named-parameters "Direct link to Named Parameters")

Use `NamedExec()` when you prefer named placeholders instead of positional arguments:

main.go

```
_, err := s.client.NamedExec(ctx,

  "INSERT INTO audit_logs (entity_id, actor_email, action) VALUES (:entity_id, :actor_email, :action)",

  entry,

)
```

`NamedExec()` accepts structs using `db` tags, `map[string]any`, repeated placeholders, and batch inserts from slices.

## Prepared Statements[​](#prepared-statements "Direct link to Prepared Statements")

Use `Prepare()` when you want to reuse the same SQL statement multiple times. Prepared statements expose `ExecContext()`, `GetContext()`, `QueryContext()`, `SelectContext()`, and `Close()`. `QueryxContext()` also exists as a deprecated compatibility alias for `QueryContext()`.

Prepared statements can also be rebound to a transaction using `Stmt.WithTx()` together with `tx.SQLTx()`. When you do this, prepare the statement from the same underlying client or connection source that created the transaction:

main.go

```
return s.client.WithTx(ctx, func(tx sqlc.Tx) error {

  txStmt := stmt.WithTx(ctx, tx.SQLTx())



  for _, postId := range postIds {

    if _, err := txStmt.ExecContext(ctx, "published", postId); err != nil {

      return fmt.Errorf("failed to publish post %d: %w", postId, err)

    }

  }



  return nil

})
```

## DELETE Operations[​](#delete-operations "Direct link to DELETE Operations")

Use `Delete()` to create a DELETE builder with `Where()` conditions:

main.go

```
func (s *BlogService) deleteComment(ctx context.Context, commentId int64) error {

  result, err := sqlc.Delete("comments").

    WithClient(s.client).

    Where(sqlc.Col("id").Eq(commentId)).

    Exec(ctx)

  if err != nil {

    return fmt.Errorf("failed to delete comment: %w", err)

  }



  rowsAffected, err := result.RowsAffected()

  if err != nil {

    return fmt.Errorf("failed to get rows affected: %w", err)

  }



  if rowsAffected == 0 {

    return fmt.Errorf("no comment found with id %d", commentId)

  }



  return nil

}
```

## Transactions[​](#transactions "Direct link to Transactions")

Use `WithTx()` to execute multiple operations atomically. The transaction automatically commits on success or rolls back on error:

main.go

```
func (s *BlogService) createAuthorWithPost(ctx context.Context, authorName, authorEmail, postTitle, postBody string) (*Author, *Post, error) {

  var author *Author

  var post *Post



  err := s.client.WithTx(ctx, func(tx sqlc.Tx) error {

    author = &Author{

      Name:  authorName,

      Email: authorEmail,

    }



    result, err := tx.Q().Into("authors").Records(author).Exec(ctx)

    if err != nil {

      return fmt.Errorf("failed to insert author: %w", err)

    }



    authorId, err := result.LastInsertId()

    if err != nil {

      return fmt.Errorf("failed to get last insert id: %w", err)

    }

    author.Id = authorId



    post = &Post{

      AuthorId: authorId,

      Title:    postTitle,

      Body:     postBody,

      Status:   "draft",

    }



    result, err = tx.Q().Into("posts").Records(post).Exec(ctx)

    if err != nil {

      return fmt.Errorf("failed to insert post: %w", err)

    }



    postId, err := result.LastInsertId()

    if err != nil {

      return fmt.Errorf("failed to get last insert id: %w", err)

    }

    post.Id = postId



    return nil

  })

  if err != nil {

    return nil, nil, err

  }



  return author, post, nil

}
```

Inside the transaction callback, use `tx.Q()` instead of `client.Q()` to execute queries within the transaction scope. If the callback returns an error, all changes are rolled back; if it returns `nil`, changes are committed.

`tx.Q()` uses the default query builder configuration. If you rely on a custom `QueryBuilderConfig` - for example PostgreSQL `$1` placeholders and `"` identifier quotes - apply the matching config explicitly on the builders you create inside the transaction.

For interoperability with libraries that require `database/sql`, `sqlc.Tx` also exposes `SQLTx()` to return the underlying `*sql.Tx`.

## Working with DB Handles[​](#working-with-db-handles "Direct link to Working with DB Handles")

Besides `sqlc.Client`, the package also provides a sqlc-owned `DB` wrapper for direct `database/sql` interop while preserving sqlc's query, scan, and named parameter behavior.

Create a database handle directly when you need access to the underlying `*sql.DB`:

main.go

```
db, err := sqlc.NewDB(ctx, config, logger, "default")

if err != nil {

  return fmt.Errorf("failed to create db handle: %w", err)

}

defer db.Close()



if _, err = db.SQLDB().ExecContext(ctx, "SET SESSION sql_mode = 'STRICT_ALL_TABLES'"); err != nil {

  return fmt.Errorf("failed to configure sql session: %w", err)

}
```

If you already have an existing `*sql.DB`, wrap it and create a client from it:

main.go

```
db := sqlc.WrapDB(existing, "mysql")

client := sqlc.NewClientWithDB(logger, db, exec.NewDefaultExecutor(), sqlc.DefaultConfig())
```

When creating a client from a wrapped handle, make sure the `QueryBuilderConfig` matches your driver. `sqlc.DefaultConfig()` uses MySQL-style placeholders (`?`) and identifier quotes (`` ` ``).

## Integration Notes[​](#integration-notes "Direct link to Integration Notes")

* `sqlr.RepositoryTx` can reuse prepared statements inside transactions, but the repository client must come from the same connection source that opens the transaction.
* `sqlh.WithTx()` currently opens request-scoped transactions from the `default` sqlc client. If your handlers or repositories use another named client, use matching middleware or a custom transaction setup.
* In `sqlh`, the request-scoped transaction created by `WithTx()` is consumed by the `BindTx*` helpers. Generated CRUD handlers use repositories directly and do not automatically switch to that request transaction.
