# sqlc Documentation Gaps

Features present in the codebase that are not covered by the current documentation.

---

## High Impact

### Expression System

The docs only show `sqlc.Col("column").Eq(value)` in passing. The expression system is the core building block for all query conditions and should have a dedicated section.

- Column references: `Col("table", "column")`, `Literal("raw sql")`, `Lit(value)`, `Param(value)`
- Comparisons: `.Eq()`, `.NotEq()`, `.Gt()`, `.Gte()`, `.Lt()`, `.Lte()`
- Set ops: `.In()`, `.NotIn()`
- Null checks: `.IsNull()`, `.IsNotNull()`
- Pattern matching: `.Like()`, `.NotLike()`
- Range: `.Between()`, `.NotBetween()`
- Logical operators: `And(...)`, `Or(...)`, `Not(...)`
- Aliasing and ordering: `.As("alias")`, `.Asc()`, `.Desc()`
- Eq map: shorthand equality conditions from a `map[string]any`

### SQL Functions

A large library of type-safe SQL functions is completely undocumented.

- **Aggregates**: `Count()`, `Sum()`, `Avg()`, `Min()`, `Max()`, `GroupConcat()`, `StdDev()`, `Variance()`
- **Numeric**: `Abs()`, `Ceil()`, `Floor()`, `Round()`, `RoundN()`, `Sqrt()`, `Pow()`, `Mod()`, `Sign()`, `Truncate()`, `Rand()`
- **String**: `Upper()`, `Lower()`, `Concat()`, `ConcatWs()`, `Trim()`, `Ltrim()`, `Rtrim()`, `Substring()`, `Left()`, `Right()`, `Replace()`, `Reverse()`, `Repeat()`, `Length()`, `CharLength()`, `Locate()`, `Lpad()`, `Rpad()`
- **MySQL 8 — JSON**: `JsonExtract()`, `JsonUnquote()`, `JsonContains()`, `JsonLength()`, `JsonType()`, `JsonKeys()`, `JsonValid()`
- **MySQL 8 — Date/Time**: `Now()`, `CurDate()`, `CurTime()`, `Date()`, `DateFormat()`, `StrToDate()`, `UnixTimestamp()`, `FromUnixTime()`, `Year()`, `Month()`, `Day()`, `Hour()`, `Minute()`, `Second()`, `DateDiff()`, `TimestampDiff()`, `LastDay()`
- **MySQL 8 — Regex**: `RegexpLike()`, `RegexpReplace()`, `RegexpInstr()`, `RegexpSubstr()`
- **MySQL 8 — Conditional**: `IfNull()`, `NullIf()`, `If()`, `Greatest()`, `Least()`, `Cast()`

### Generic Query Builders

The package provides a full set of type-parameterized query builders that return typed results without needing a `dest` pointer. Completely undocumented.

- `FromG[T](table)` — type-safe SELECT
- `IntoG[T](table)` — type-safe INSERT
- `UpdateG[T](table)` — type-safe UPDATE
- `DeleteG[T](table)` — type-safe DELETE
- `QG[T](client)` — generic query builder factory attached to a client

---

## Medium Impact

### GROUP BY / HAVING / DISTINCT

Fundamental SQL aggregation features available on the SELECT builder but not mentioned in the docs.

- `.GroupBy(cols...)`
- `.Having(condition, params...)`
- `.Distinct()`

### Prepared Statements

The method list mentions `Prepare` but never explains or demonstrates it. The package has dedicated types for prepared statement use.

- `PreparedSelect` — `Get()`, `Select()`, `Query()`, `Close()`
- `PreparedExec` — `Exec()`, `Close()`
- `PreparedSelectG[T]` — generic, type-safe variant

### INSERT Advanced Features

Several INSERT builder capabilities are undocumented.

- `Replace()` — MySQL `REPLACE INTO`
- `Ignore()` — `INSERT IGNORE`
- `OnDuplicateKeyUpdate(assignments...)` — upsert support via `Assign()` / `AssignExpr()`
- `LowPriority()`, `HighPriority()`, `Delayed()` — priority modifiers
- `Values()`, `ValuesRows()`, `ValuesMaps()` — non-struct insert data formats

### JSON Column Type

`JSON[T, NullBehaviour]` — a generic type for transparent JSON column marshaling/unmarshaling via the standard `database/sql` `Scanner`/`Valuer` interfaces. Supports both nullable and non-nullable variants.

### JSON Filter System

`JsonFilter` converts JSON-structured filter definitions into SQL WHERE expressions. Useful for building filterable API endpoints. Has a standalone `JSON_FILTER.md` but nothing in the main documentation.

- `JsonFilterFromJSON(jsonStr)` — parse a JSON string into a filter
- `JsonFilter.ToExpression()` — convert to an `Expression` for use in queries
- Supported filter types: `and`, `or`, `not`, `eq`, `ne`, `lt`, `lte`, `gt`, `gte`, `in`, `notIn`, `between`, `like`, `notLike`, `isNull`, `isNotNull`

---

## Low Impact

### UPDATE Additional Methods

- `SetMap(map[string]any)` — set multiple columns from a map
- `SetRecord(record)` — set columns from a struct
- `OrderBy()` and `Limit()` on the UPDATE builder

### Retry / Error Handling

The executor provides automatic retry with configurable backoff for transient database errors. Not documented at all.

- Handled errors: deadlocks (MySQL 1213), invalid connection, bad connection, I/O timeout
- Configuration key: `db.<name>.retry`

### Metrics / Observability

Connection pool metrics are automatically published every minute under the metric name `DbConnectionCount`. Tracked dimensions: new connections, open connections, in-use connections, idle connections.

### SQL Parser

`ParseWhere(input string)` parses a raw SQL WHERE string into a type-safe `Expression`. Useful for accepting dynamic filter strings from external sources.

### Custom Migration Providers

`AddMigrationProvider(name, provider)` — register a custom migration provider beyond the default `goose` integration.

### Custom Driver Registration

`AddDriverFactory(name, factory)` — register a custom database driver beyond the built-in `mysql` and `postgres` drivers.

### Testing Support

The `mocks/` directory provides mock implementations for `Client`, `Querier`, `Tx`, and other interfaces, suitable for use with `github.com/stretchr/testify/mock`. Not mentioned anywhere in the docs.
