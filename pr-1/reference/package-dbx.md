# dbx package

The dbx package provides a fluent API to build SQL queries.

## Client[​](#client "Direct link to Client")

The client is the main entry point to the package. It is used to create query builders.

### YourModel struct[​](#yourmodel-struct "Direct link to YourModel struct")

```
type YourModel struct {

    Id   int    `db:"id"`

    Name string `db:"name"`

}
```

### [NewClient()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#newclient "Direct link to newclient")

#### Usage[​](#usage "Direct link to Usage")

```
client, err := dbx.NewClient[YourModel](ctx, config, logger, "default", "your_table")
```

#### Description[​](#description "Direct link to Description")

Creates a new dbx client. You need to provide a model (a struct with `db` tags), which will be used to map the database table to your struct.

The `clientName` argument is the name of the database client to use, as configured in your application's configuration file.<br /><!-- -->The `table` argument is the name of the database table.

## Query Builders[​](#query-builders "Direct link to Query Builders")

The dbx package provides query builders for SELECT, INSERT, UPDATE, DELETE, and GET statements.

### [Get()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#get "Direct link to get")

#### Usage[​](#usage-1 "Direct link to Usage")

```
result, err := client.Get().Where(dbx.Eq{"id": 1}).Exec(ctx)

if errors.Is(err, dbx.ErrNotFound) {

    // handle not found

}
```

```
SELECT id, name FROM your_table WHERE id = 1 LIMIT 2
```

#### Description[​](#description-1 "Direct link to Description")

Creates a new GET query builder. This is a convenience method that wraps `Select()` and returns a single element or an error.

**Return values:**

* Returns a single element of type `T` if exactly one row is found
* Returns `dbx.ErrNotFound` if no rows match the query
* Returns an error if more than one row is found

The `Get` builder automatically adds `LIMIT 2` to detect when multiple rows match the query.

You can also pass a struct of type `T` to the `Where` clause:

```
result, err := client.Get().Where(YourModel{Id: 1}).Exec(ctx)
```

```
SELECT id, name FROM your_table WHERE id = 1 LIMIT 2
```

The `Get` builder supports the same methods as the `Select` builder, including `Join`, `OrderBy`, `GroupBy`, and `Having`.

**Advanced Usage:**

```
// Get with JOIN

result, err := client.Get().

    Join("profiles p ON your_table.id = p.user_id").

    Where(dbx.Eq{"p.verified": true}).

    Exec(ctx)



// Get with ORDER BY (returns first match after ordering)

result, err := client.Get().

    Where(dbx.Eq{"status": "active"}).

    OrderBy("created_at DESC").

    Exec(ctx)
```

### [Select()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#select "Direct link to select")

#### Usage[​](#usage-2 "Direct link to Usage")

```
results, err := client.Select().Where(dbx.Eq{"id": 1}).Exec(ctx)
```

```
SELECT id, name FROM your_table WHERE id = 1
```

#### Description[​](#description-2 "Direct link to Description")

Creates a new SELECT query builder. You can use the `Where` method to add conditions to your query. The `Exec` method executes the query and returns the results.

You can also pass a struct of type `T` to the `Where` clause. In this case, the struct will be converted to a map and the non-zero fields will be used to build an `Eq` expression. This is available for all query builders, not just `Select`.

```
results, err := client.Select().Where(YourModel{Id: 1}).Exec(ctx)
```

```
SELECT id, name FROM your_table WHERE id = 1
```

### [Insert()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#insert "Direct link to insert")

#### Usage[​](#usage-3 "Direct link to Usage")

##### Single Insert[​](#single-insert "Direct link to Single Insert")

```
_, err := client.Insert(YourModel{Id: 1, Name: "test"}).Exec(ctx)
```

```
INSERT INTO your_table (id,name) VALUES (1,'test')
```

##### Batch Insert[​](#batch-insert "Direct link to Batch Insert")

```
_, err := client.Insert(

    YourModel{Id: 1, Name: "test1"},

    YourModel{Id: 2, Name: "test2"},

).Exec(ctx)
```

```
INSERT INTO your_table (id,name) VALUES (1,'test1'), (2,'test2');
```

#### Description[​](#description-3 "Direct link to Description")

Creates a new INSERT query builder. You need to provide the model with the values to insert. You can also provide multiple models to insert multiple rows at once. The `Exec` method executes the query.

### [Replace()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#replace "Direct link to replace")

#### Usage[​](#usage-4 "Direct link to Usage")

```
_, err := client.Replace(YourModel{Id: 1, Name: "test"}).Exec(ctx)
```

```
REPLACE INTO your_table (id,name) VALUES (1,'test')
```

#### Description[​](#description-4 "Direct link to Description")

Creates a new REPLACE query builder. You need to provide the model with the values to insert. The `Exec` method executes the query.

### [Update()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#update "Direct link to update")

#### Usage[​](#usage-5 "Direct link to Usage")

```
_, err := client.Update(map[string]any{"name": "new_name"}).Where(dbx.Eq{"id": 1}).Exec(ctx)
```

```
UPDATE your_table SET name = 'new_name' WHERE id = 1
```

#### Description[​](#description-5 "Direct link to Description")

Creates a new UPDATE query builder. You can provide one or more maps with the new values, which will be merged together. You can use the `Where` method to add conditions to your query. The `Exec` method executes the query.

You can also pass a struct of type `T` to the `Update` method. In this case, the struct will be converted to a map and the non-zero fields will be used to build the `SET` clause of the query.

```
_, err := client.Update(YourModel{Name: "new_name"}).Where(dbx.Eq{"id": 1}).Exec(ctx)
```

```
UPDATE your_table SET name = 'new_name' WHERE id = 1
```

### [Delete()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#delete "Direct link to delete")

#### Usage[​](#usage-6 "Direct link to Usage")

```
_, err := client.Delete().Where(dbx.Eq{"id": 1}).Exec(ctx)
```

```
DELETE FROM your_table WHERE id = 1
```

#### Description[​](#description-6 "Direct link to Description")

Creates a new DELETE query builder. You can use the `Where` method to add conditions to your query. The `Exec` method executes the query.

## Error Handling[​](#error-handling "Direct link to Error Handling")

### [ErrNotFound](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/client.go)[​](#errnotfound "Direct link to errnotfound")

The `dbx.ErrNotFound` error is returned by the `Get()` method when no rows match the query. You can use `errors.Is()` to check for this error:

```
result, err := client.Get().Where(dbx.Eq{"id": 1}).Exec(ctx)

if errors.Is(err, dbx.ErrNotFound) {

    // handle not found case

    return nil, fmt.Errorf("user not found")

}

if err != nil {

    // handle other errors

    return nil, err

}

// use result
```

## Query Builder Methods[​](#query-builder-methods "Direct link to Query Builder Methods")

The `Select` and `Get` builders provide the following methods to build complex queries. Note that `Get` automatically adds `LIMIT 2` to detect multiple results.

### [Distinct()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/select.go)[​](#distinct "Direct link to distinct")

Adds a `DISTINCT` clause to the query.

```
results, err := client.Select().Distinct().Exec(ctx)
```

```
SELECT DISTINCT id, name FROM your_table
```

### [Join()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/select.go)[​](#join "Direct link to join")

Adds a `JOIN` clause to the query.

```
results, err := client.Select().Join("other_table ON other_table.id = your_table.id").Exec(ctx)
```

```
SELECT id, name FROM your_table JOIN other_table ON other_table.id = your_table.id
```

You can also use `LeftJoin`, `RightJoin`, `InnerJoin` and `CrossJoin`.

### [GroupBy()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/select.go)[​](#groupby "Direct link to groupby")

Adds a `GROUP BY` clause to the query.

```
results, err := client.Select().GroupBy("name").Exec(ctx)
```

```
SELECT id, name FROM your_table GROUP BY name
```

### [Having()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/select.go)[​](#having "Direct link to having")

Adds a `HAVING` clause to the query.

```
results, err := client.Select().GroupBy("name").Having(dbx.Gt{"COUNT(id)": 1}).Exec(ctx)
```

```
SELECT id, name FROM your_table GROUP BY name HAVING COUNT(id) > 1
```

### [OrderBy()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/select.go)[​](#orderby "Direct link to orderby")

Adds an `ORDER BY` clause to the query.

```
results, err := client.Select().OrderBy("name DESC").Exec(ctx)
```

```
SELECT id, name FROM your_table ORDER BY name DESC
```

### [Limit()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/select.go)[​](#limit "Direct link to limit")

Adds a `LIMIT` clause to the query.

```
results, err := client.Select().Limit(10).Exec(ctx)
```

```
SELECT id, name FROM your_table LIMIT 10
```

### [Offset()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/select.go)[​](#offset "Direct link to offset")

Adds an `OFFSET` clause to the query.

```
results, err := client.Select().Limit(10).Offset(10).Exec(ctx)
```

```
SELECT id, name FROM your_table LIMIT 10 OFFSET 10
```

## Expressions[​](#expressions "Direct link to Expressions")

The `Sqlizer` interface can be converted to SQL. There are many implementations of this interface.

### [Expr](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#expr "Direct link to expr")

Builds an expression from a SQL fragment and arguments.

```
client.Select().Where(dbx.Expr("FROM_UNIXTIME(?) > ?", 1672531200, 1000))
```

```
SELECT id, name FROM your_table WHERE FROM_UNIXTIME(1672531200) > 1000
```

### [ConcatExpr](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#concatexpr "Direct link to concatexpr")

Builds an expression by concatenating strings and other expressions.

```
client.Select().Column(dbx.ConcatExpr("COALESCE(full_name, ", dbx.Expr("CONCAT(?, ' ', ?)", "first", "last"), ")"))
```

```
SELECT COALESCE(full_name, CONCAT('first', ' ', 'last')) FROM your_table
```

### [Alias](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#alias "Direct link to alias")

Allows to define an alias for a column.

```
client.Select().Column(dbx.Alias(dbx.Expr("COUNT(*)"), "total"))
```

```
SELECT (COUNT(*)) AS total FROM your_table
```

### [Eq](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#eq "Direct link to eq")

Creates an "equal" expression.

```
client.Select().Where(dbx.Eq{"id": 1})
```

```
SELECT id, name FROM your_table WHERE id = 1
```

### [NotEq](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#noteq "Direct link to noteq")

Creates a "not equal" expression.

```
client.Select().Where(dbx.NotEq{"id": 1})
```

```
SELECT id, name FROM your_table WHERE id <> 1
```

### [Like](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#like "Direct link to like")

Creates a "like" expression.

```
client.Select().Where(dbx.Like{"name": "%test%"})
```

```
SELECT id, name FROM your_table WHERE name LIKE '%test%'
```

### [NotLike](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#notlike "Direct link to notlike")

Creates a "not like" expression.

```
client.Select().Where(dbx.NotLike{"name": "%test%"})
```

```
SELECT id, name FROM your_table WHERE name NOT LIKE '%test%'
```

### [ILike](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#ilike "Direct link to ilike")

Creates a case-insensitive "like" expression.

```
client.Select().Where(dbx.ILike{"name": "test%"})
```

```
SELECT id, name FROM your_table WHERE name ILIKE 'test%'
```

### [NotILike](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#notilike "Direct link to notilike")

Creates a case-insensitive "not like" expression.

```
client.Select().Where(dbx.NotILike{"name": "test%"})
```

```
SELECT id, name FROM your_table WHERE name NOT ILIKE 'test%'
```

### [Lt](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#lt "Direct link to lt")

Creates a "less than" expression.

```
client.Select().Where(dbx.Lt{"age": 18})
```

```
SELECT id, name FROM your_table WHERE age < 18
```

### [LtOrEq](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#ltoreq "Direct link to ltoreq")

Creates a "less than or equal" expression.

```
client.Select().Where(dbx.LtOrEq{"age": 18})
```

```
SELECT id, name FROM your_table WHERE age <= 18
```

### [Gt](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#gt "Direct link to gt")

Creates a "greater than" expression.

```
client.Select().Where(dbx.Gt{"age": 18})
```

```
SELECT id, name FROM your_table WHERE age > 18
```

### [GtOrEq](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#gtoreq "Direct link to gtoreq")

Creates a "greater than or equal" expression.

```
client.Select().Where(dbx.GtOrEq{"age": 18})
```

```
SELECT id, name FROM your_table WHERE age >= 18
```

### [And](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#and "Direct link to and")

Creates an "and" conjunction of expressions.

```
client.Select().Where(dbx.And(dbx.Eq{"id": 1}, dbx.Like{"name": "%test%"}))
```

```
SELECT id, name FROM your_table WHERE (id = 1 AND name LIKE '%test%')
```

### [Or](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/expr.go)[​](#or "Direct link to or")

Creates an "or" conjunction of expressions.

```
client.Select().Where(dbx.Or(dbx.Eq{"id": 1}, dbx.Eq{"id": 2}))
```

```
SELECT id, name FROM your_table WHERE (id = 1 OR id = 2)
```

## Placeholder Formats[​](#placeholder-formats "Direct link to Placeholder Formats")

The dbx package supports two placeholder formats: `Question` and `Dollar`. The default is `Question`. You can change the placeholder format by passing it to the `NewClientWithInterfaces` function.

### [Question](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/placeholder.go)[​](#question "Direct link to question")

The `Question` format uses `?` as a placeholder.

### [Dollar](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/dbx/placeholder.go)[​](#dollar "Direct link to dollar")

The `Dollar` format uses `$` as a placeholder.
