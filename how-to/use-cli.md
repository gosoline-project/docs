# Build a CLI tool

The `pkg/cli` package lets you build multi-command CLI tools that run inside the gosoline application lifecycle — giving each command access to the same config, logger, and kernel modules as any other gosoline application.

In this guide, you'll learn how to:

* Define a simple command
* Group related commands (`api serve`, `db migrate`)
* Add flags and read their values from config
* Add a built-in `version` command

## Overview[​](#overview "Direct link to Overview")

The entry point is `cli.NewCli()`. It embeds a `*Router`, which is a tree of named groups and commands. When you call `Run()`, it parses `os.Args`, resolves the matching command, injects parsed flags into the gosoline config, and starts the kernel with the command's application options.

```
c := cli.NewCli(/* options... */)

// register groups and commands on c

c.Run()
```

## Defining a command[​](#defining-a-command "Direct link to Defining a command")

Register a `Cmd` directly on the `Cli` (or on any child `*Router` returned by `Group()`). Each `Cmd` needs a `Name` and `AppOptions` containing the module to run:

```
c.Cmd(cli.Cmd{

    Name: "migrate",

    AppOptions: []application.Option{

        application.WithModuleFactory("main", cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {

            return func(ctx context.Context) error {

                logger.Info(ctx, "running migrations")

                return nil

            }, nil

        })),

    },

})
```

`application.WithModuleFactory` registers the command's kernel module. `cli.WithRunFunc` is a convenience wrapper that turns a plain function into a `kernel.ModuleFactory`.

## Grouping commands[​](#grouping-commands "Direct link to Grouping commands")

`Group()` returns a child `*Router`. Register `Cmd`s on that router to create namespaced subcommands:

```
apiRouter := c.Group(cli.Group{Name: "api"})

apiRouter.Cmd(cli.Cmd{Name: "serve", AppOptions: /* ... */})



dbRouter := c.Group(cli.Group{Name: "db"})

dbRouter.Cmd(cli.Cmd{Name: "migrate", AppOptions: /* ... */})
```

This produces commands like `myapp api serve` and `myapp db migrate`. Groups can be nested arbitrarily deep.

## Adding flags[​](#adding-flags "Direct link to Adding flags")

Declare flags on a `Cmd` (or `Group`) with `cli.Flag`. Each flag has a short name, a long name, an optional default, and an optional `CfgKey` to map the value to an arbitrary config path:

```
cli.Cmd{

    Name: "serve",

    Flags: []cli.Flag{

        {Short: "p", Long: "port", CfgKey: "httpserver.default.port", Default: "8080", Description: "port to listen on"},

    },

    AppOptions: /* ... */,

}
```

| Field         | Description                                           |
| ------------- | ----------------------------------------------------- |
| `Short`       | Single-character flag (`-p`)                          |
| `Long`        | Long-form flag (`--port`)                             |
| `Default`     | Value used when the flag is absent                    |
| `CfgKey`      | If set, the value is also written to this config path |
| `Description` | Human-readable description                            |

The flag value is always available at `cli.flags.<long>` in the config. If `CfgKey` is set, it is also written there. The short flag takes precedence over the long flag when both are provided.

Read the value inside your module registration:

```
AppOptions: []application.Option{

    application.WithModuleFactory("main", cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {

        port, err := config.GetString("httpserver.default.port") // from CfgKey

        // or: config.GetString("cli.flags.port")                // always available

        if err != nil {

            return nil, err

        }



        return func(ctx context.Context) error {

            logger.Info(ctx, "listening on port %s", port)

            <-ctx.Done()

            return nil

        }, nil

    })),

},
```

## Default command[​](#default-command "Direct link to Default command")

Use `DefaultCmd` to register a fallback that runs when no command matches — useful for showing usage:

```
c.DefaultCmd(cli.Cmd{

    AppOptions: []application.Option{

        application.WithModuleFactory("main", cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {

            return func(ctx context.Context) error {

                fmt.Println("Usage: myapp <command> [flags]")

                return nil

            }, nil

        })),

    },

})
```

## Built-in version command[​](#built-in-version-command "Direct link to Built-in version command")

Pass `cli.WithVersion` to add a `version` subcommand that prints a string and exits:

```
c := cli.NewCli(

    cli.WithVersion("1.0.0"),

)
```

```
$ myapp version

1.0.0
```

## Complete example[​](#complete-example "Direct link to Complete example")

main.go

main.go

```
package main



import (

	"context"

	"fmt"



	"github.com/justtrackio/gosoline/pkg/application"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/cli"

	"github.com/justtrackio/gosoline/pkg/kernel"

	"github.com/justtrackio/gosoline/pkg/log"

)



func main() {

	c := cli.NewCli(

		cli.WithVersion("1.0.0"),

		cli.WithAppOptions(

			application.WithConfigFile("config.dist.yml", "yml"),

		),

	)



	// Register subcommands under the "api" group.

	apiRouter := c.Group(cli.Group{Name: "api"})

	apiRouter.Cmd(cli.Cmd{

		Name: "serve",

		Flags: []cli.Flag{

			{Short: "p", Long: "port", CfgKey: "httpserver.default.port", Default: "8080", Description: "port to listen on"},

		},

		AppOptions: []application.Option{

			application.WithModuleFactory("main", cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {

				port, err := config.GetString("httpserver.default.port")

				if err != nil {

					return nil, err

				}



				return func(ctx context.Context) error {

					logger.Info(ctx, "api server listening on port %s", port)

					<-ctx.Done()

					return nil

				}, nil

			})),

		},

	})



	// Register subcommands under the "db" group.

	dbRouter := c.Group(cli.Group{Name: "db"})

	dbRouter.Cmd(cli.Cmd{

		Name: "migrate",

		Flags: []cli.Flag{

			{Short: "e", Long: "env", CfgKey: "app.env", Default: "dev", Description: "target environment"},

		},

		AppOptions: []application.Option{

			application.WithModuleFactory("main", cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {

				env, err := config.GetString("app.env")

				if err != nil {

					return nil, err

				}



				return func(ctx context.Context) error {

					logger.Info(ctx, "running migrations in env: %s", env)

					return nil

				}, nil

			})),

		},

	})



	// Fallback when no command is matched.

	c.DefaultCmd(cli.Cmd{

		AppOptions: []application.Option{

			application.WithModuleFactory("main", cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {

				return func(ctx context.Context) error {

					fmt.Println("Usage: myapp <command> [flags]")

					fmt.Println("Commands:")

					fmt.Println("  api serve    Start the API server")

					fmt.Println("  db migrate   Run database migrations")

					fmt.Println("  version      Print the version")

					return nil

				}, nil

			})),

		},

	})



	c.Run()

}
```

config.dist.yml

config.dist.yml

```
app:

  env: dev

  name: myapp
```

Run the commands:

```
# Show usage (default command)

go run main.go



# Start the API server on a custom port

go run main.go api serve --port 9090



# Run migrations against production

go run main.go db migrate --env prod



# Print the version

go run main.go version
```
