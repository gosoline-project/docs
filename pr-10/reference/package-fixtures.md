# Gosoline Fixture Loader

Define fixture data with `fixtures.NamedFixtures[T]`, pass it to a writer, and register fixture-set factories with the application.

## Usage[​](#usage "Direct link to Usage")

Enable fixture loading in `config.dist.yml`:

```
fixtures:

  enabled: true

  groups: [default]
```

Build or run the application with the `fixtures` build tag. Without that tag, the fixture loader is a no-op. Only the configured groups are loaded; `default` is the default group.

`fixtures.FixtureSetFactory` creates one set. `fixtures.FixtureSetsFactory` creates several sets and receives the group name:

```
type FixtureSetFactory func(context.Context, cfg.Config, log.Logger) (fixtures.FixtureSet, error)

type FixtureSetsFactory func(context.Context, cfg.Config, log.Logger, string) ([]fixtures.FixtureSet, error)
```

Use `fixtures.NewFixtureSetsFactory` to combine single-set factories. Register the resulting factory with `application.WithFixtureSetFactory("default", factory)`.

`fixtures.NewSimpleFixtureSet(data, writer, options...)` implements the `fixtures.FixtureSet` interface. `fixtures.WithEnabled(false)` skips a set. The old `fixtures.WithPurge` option and the `Purge` field no longer exist; cleanup behavior belongs to the selected writer or an application-owned set.

Fixture writers live in the package for their destination:

| Destination                | Constructor                                   | Single-set factory                             |
| -------------------------- | --------------------------------------------- | ---------------------------------------------- |
| Blob store                 | `blob.NewBlobFixtureWriter`                   | `blob.BlobFixtureSetFactory`                   |
| DynamoDB                   | `ddb.NewDynamoDbFixtureWriter`                | `ddb.DynamoDbFixtureSetFactory`                |
| DynamoDB key-value store   | `kvstore.NewDynamoDbKvStoreFixtureWriter`     | `kvstore.DynamoDbKvStoreFixtureSetFactory`     |
| Configured key-value store | `kvstore.NewConfigurableKvStoreFixtureWriter` | `kvstore.ConfigurableKvStoreFixtureSetFactory` |
| MySQL ORM                  | `db_repo.NewMysqlOrmFixtureWriter`            | `db_repo.MysqlOrmFixtureSetFactory`            |
| MySQL columns and values   | `db.NewMysqlPlainFixtureWriter`               | `db.MysqlPlainFixtureSetFactory`               |
| MySQL named SQL statements | `db.NewMysqlSqlxFixtureWriter`                | `db.MysqlSqlxFixtureSetFactory`                |
| Redis                      | `redis.NewRedisFixtureWriter`                 | `redis.RedisFixtureSetFactory`                 |
| Redis key-value store      | `kvstore.NewRedisKvStoreFixtureWriter`        | `kvstore.RedisKvStoreFixtureSetFactory`        |

The import for `db_repo` is `github.com/justtrackio/gosoline/pkg/db-repo`. Constructors, metadata, and write behavior differ by destination; inspect the [released writer implementations](https://github.com/justtrackio/gosoline/tree/v0.68.3/pkg) before choosing one.

## Quick Usage[​](#quick-usage "Direct link to Quick Usage")

The following example registers a `fixtures.FixtureSetsFactory` under the `default` group. Configure its backing services and blob fixture files before running it with `go run -tags fixtures .` from the example directory.

## ID Generation[​](#id-generation "Direct link to ID Generation")

Use `fixtures.NewNumberSequence` or `fixtures.NewNumberSequenceFrom` for sequential identifiers. The older `NewAutoNumbered` and `NewAutoNumberedFrom` names remain deprecated aliases. Use `fixtures.NewUuidSequence` or `fixtures.NewByteSequence` when a deterministic seeded UUID or byte sequence is needed.

## Example with available fixture writers[​](#example-with-available-fixture-writers "Direct link to Example with available fixture writers")

This example is read directly from the maintained source file:

examples/fixtures/simple/main.go

```
//go:build fixtures



package main



import (

	"context"

	"fmt"

	"time"



	"github.com/justtrackio/gosoline/pkg/application"

	"github.com/justtrackio/gosoline/pkg/blob"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/db"

	"github.com/justtrackio/gosoline/pkg/db-repo"

	"github.com/justtrackio/gosoline/pkg/ddb"

	"github.com/justtrackio/gosoline/pkg/fixtures"

	"github.com/justtrackio/gosoline/pkg/kernel"

	"github.com/justtrackio/gosoline/pkg/kvstore"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/mdl"

	"github.com/justtrackio/gosoline/pkg/redis"

)



type DynamoDbExampleModel struct {

	Name  string `ddb:"key=hash"`

	Value string `ddb:"global=hash"`

}



type OrmFixtureExample struct {

	db_repo.Model

	Name *string

}



var autoNumbered = fixtures.NewNumberSequenceFrom(2)



func mysqlOrmFixtureSet(ctx context.Context, config cfg.Config, logger log.Logger) (fixtures.FixtureSet, error) {

	mysqlMetadata := &db_repo.Metadata{

		ModelId: mdl.ModelId{

			Name: "orm_fixture_example",

		},

		TableName: "orm_fixture_examples",

	}

	mysqlOrmWriter, err := db_repo.NewMysqlOrmFixtureWriter(ctx, config, logger, mysqlMetadata)

	if err != nil {

		return nil, fmt.Errorf("failed to initialize mysql orm writer: %w", err)

	}



	return fixtures.NewSimpleFixtureSet(fixtures.NamedFixtures[*OrmFixtureExample]{

		&fixtures.NamedFixture[*OrmFixtureExample]{

			Name: "foo",

			Value: &OrmFixtureExample{

				Model: db_repo.Model{

					Id: autoNumbered.GetNext(),

				},

				Name: mdl.Box("example"),

			},

		},

	}, mysqlOrmWriter), nil

}



func mysqlPlainFixtureSet(ctx context.Context, config cfg.Config, logger log.Logger) (fixtures.FixtureSet, error) {

	mysqlPlainWriter, err := db.NewMysqlPlainFixtureWriter(ctx, config, logger, &db.MysqlPlainMetaData{

		TableName: "plain_fixture_example",

		Columns:   []string{"id", "name"},

	})

	if err != nil {

		return nil, fmt.Errorf("failed to initialize mysql plain writer: %w", err)

	}



	return fixtures.NewSimpleFixtureSet(fixtures.NamedFixtures[db.MysqlPlainFixtureValues]{

		{

			Name:  "foo2",

			Value: db.MysqlPlainFixtureValues{1, "testName1"},

		},

		{

			Name:  "foo3",

			Value: db.MysqlPlainFixtureValues{2, "testName2"},

		},

	}, mysqlPlainWriter), nil

}



func dynamodbKvstoreFixtureSet(ctx context.Context, config cfg.Config, logger log.Logger) (fixtures.FixtureSet, error) {

	dynamoDbKvStoreWriter, err := kvstore.NewDynamoDbKvStoreFixtureWriter[DynamoDbExampleModel](ctx, config, logger, &mdl.ModelId{

		Name:        "exampleModel",

		Env:         "dev",

		Application: "fixture-loader",

		Tags: map[string]string{

			"project": "gosoline",

			"family":  "example",

		},

	})

	if err != nil {

		return nil, fmt.Errorf("failed to initialize dynamodb kvstore writer: %w", err)

	}



	return fixtures.NewSimpleFixtureSet(fixtures.NamedFixtures[*kvstore.KvStoreFixture]{

		{

			Name: "kv_somekey",

			Value: &kvstore.KvStoreFixture{

				Key:   "SomeKey",

				Value: DynamoDbExampleModel{Name: "Some Name", Value: "Some Value"},

			},

		},

	}, dynamoDbKvStoreWriter), nil

}



func redisFixtureSet(ctx context.Context, config cfg.Config, logger log.Logger) (fixtures.FixtureSet, error) {

	redisWriter, err := redis.NewRedisFixtureWriter(ctx, config, logger, "default", redis.RedisOpSet)

	if err != nil {

		return nil, fmt.Errorf("failed to initialize redis writer: %w", err)

	}



	return fixtures.NewSimpleFixtureSet(fixtures.NamedFixtures[*redis.RedisFixture]{

		{

			Name: "redis_example",

			Value: &redis.RedisFixture{

				Key:    "example-key",

				Value:  "bar",

				Expiry: 1 * time.Hour,

			},

		},

	}, redisWriter), nil

}



func dynamodbFixtureSet(ctx context.Context, config cfg.Config, logger log.Logger) (fixtures.FixtureSet, error) {

	dynamodbWriter, err := ddb.NewDynamoDbFixtureWriter(ctx, config, logger, &ddb.Settings{

		ModelId: mdl.ModelId{

			Name:        "exampleModel",

			Env:         "dev",

			Application: "fixture-loader",

			Tags: map[string]string{

				"project": "gosoline",

				"family":  "example",

			},

		},

		Main: ddb.MainSettings{

			Model: DynamoDbExampleModel{},

		},

		Global: []ddb.GlobalSettings{

			{

				Name:               "IDX_Name",

				Model:              DynamoDbExampleModel{},

				ReadCapacityUnits:  1,

				WriteCapacityUnits: 1,

			},

		},

	})

	if err != nil {

		return nil, fmt.Errorf("failed to initialize dynamodb writer: %w", err)

	}



	return fixtures.NewSimpleFixtureSet(fixtures.NamedFixtures[*DynamoDbExampleModel]{

		{

			Name:  "ddb",

			Value: &DynamoDbExampleModel{Name: "Some Name", Value: "Some Value"},

		},

	}, dynamodbWriter), nil

}



func blobFixtureSet(ctx context.Context, config cfg.Config, logger log.Logger) (fixtures.FixtureSet, error) {

	fileReaderFactory := blob.NewFileReader("../../../test/fixtures/blob/test_data/fixtures_test_data")



	blobWriter, err := blob.NewBlobFixtureWriter(ctx, config, logger, fileReaderFactory, "test")

	if err != nil {

		return nil, fmt.Errorf("failed to initialize blob writer: %w", err)

	}



	return fixtures.NewSimpleFixtureSet(fixtures.NamedFixtures[*blob.BlobFixture]{}, blobWriter), nil

}



func fixtureSetsFactory(ctx context.Context, config cfg.Config, logger log.Logger, group string) ([]fixtures.FixtureSet, error) {

	mysqlOrmFs, err := mysqlOrmFixtureSet(ctx, config, logger)

	if err != nil {

		return nil, err

	}



	mysqlPlainFs, err := mysqlPlainFixtureSet(ctx, config, logger)

	if err != nil {

		return nil, err

	}



	dynamodbKvstoreFs, err := dynamodbKvstoreFixtureSet(ctx, config, logger)

	if err != nil {

		return nil, err

	}



	redisFs, err := redisFixtureSet(ctx, config, logger)

	if err != nil {

		return nil, err

	}



	dynamodbFs, err := dynamodbFixtureSet(ctx, config, logger)

	if err != nil {

		return nil, err

	}



	blobFs, err := blobFixtureSet(ctx, config, logger)

	if err != nil {

		return nil, err

	}



	return []fixtures.FixtureSet{

		mysqlOrmFs,

		mysqlPlainFs,

		dynamodbKvstoreFs,

		redisFs,

		dynamodbFs,

		blobFs,

	}, nil

}



type mod struct {

	kernel.ForegroundModule

}



func (m *mod) Run(ctx context.Context) error {

	<-ctx.Done()



	return nil

}



func main() {

	app := application.Default(

		application.WithConfigFile("config.dist.yml", "yml"),

		application.WithFixtureSetFactory("default", fixtureSetsFactory),

		application.WithModuleFactory("main", func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.Module, error) {

			return &mod{}, nil

		}),

	)



	app.Run()

}
```

## Named Fixtures[​](#named-fixtures "Direct link to Named Fixtures")

Named fixture collections let later fixtures refer to an already-created value with `GetValueByName`:

examples/fixtures/simple/named/main.go

```
//go:build fixtures



package main



import (

	"context"

	"fmt"



	"github.com/justtrackio/gosoline/pkg/application"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/db-repo"

	"github.com/justtrackio/gosoline/pkg/fixtures"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/mdl"

)



type DynamoDbExampleModel struct {

	Name  string `ddb:"key=hash"`

	Value string `ddb:"global=hash"`

}



var namedFixtures = fixtures.NamedFixtures[*DynamoDbExampleModel]{

	{

		Name:  "test",

		Value: &DynamoDbExampleModel{Name: "Some Name", Value: "Some Value"},

	},

}



func fixtureSetsFactory(ctx context.Context, config cfg.Config, logger log.Logger, group string) ([]fixtures.FixtureSet, error) {

	mysqlWriter, err := db_repo.NewMysqlOrmFixtureWriter(ctx, config, logger, &db_repo.Metadata{

		ModelId: mdl.ModelId{

			Name: "orm_named_fixture_example",

		},

	})

	if err != nil {

		return nil, fmt.Errorf("failed to create orm fixture writer: %w", err)

	}



	return []fixtures.FixtureSet{

		fixtures.NewSimpleFixtureSet(namedFixtures, mysqlWriter),

	}, nil

}



func main() {

	app := application.Default(

		application.WithFixtureSetFactory("default", fixtureSetsFactory),

	)



	app.Run()



	// then you can access them later

	fx := namedFixtures.GetValueByName("test")

	_ = fx.Value

}
```

## Further Information[​](#further-information "Direct link to Further Information")

* Writers control insert, update, and cleanup behavior; do not assume all writers upsert or purge their destinations.
* MySQL ORM fixture models embed `db_repo.Model`.
* See the [released fixture interfaces and factory helper](https://github.com/justtrackio/gosoline/blob/v0.68.3/pkg/fixtures/types.go) and [simple set options](https://github.com/justtrackio/gosoline/blob/v0.68.3/pkg/fixtures/fixture_set_options.go).
