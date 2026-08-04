# Build a CLI tool

The `pkg/cli` package lets you build multi-command CLI tools that run inside the gosoline application lifecycle, giving each command access to the same config, logger, and kernel modules as any other gosoline application.

In this guide, you'll learn how to:

* Define a simple command
* Group related commands (`api serve`, `db migrate`)
* Add flags, positional arguments, and read their values from config
* Add built-in help output
* Add a built-in `version` command

## Overview[​](#overview "Direct link to Overview")

The entry point is `cli.NewCli()`. It embeds a `*Router`, which is a tree of named groups and commands. When you call `Run()`, it parses `os.Args`, resolves the matching command, injects parsed flags and positional arguments into the gosoline config, and starts the kernel with the command's application options.

```
c := cli.NewCli(/* options... */)

// register groups and commands on c

c.Run()
```

## Defining a command[​](#defining-a-command "Direct link to Defining a command")

Register a `Cmd` directly on the `Cli` (or on any child `*Router` returned by `Group()`). Each `Cmd` needs a `Name` and `AppOptions` containing the module to run. Add `Description` and `Examples` to improve generated help output:

```
c.Cmd(cli.Cmd{

    Name:        "migrate",

    Description: "Run database migrations.",

    Examples: []cli.CmdExample{

        {Description: "Run migrations against production:", Args: "myapp migrate --env prod"},

    },

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

For simple command modules, you can also use `cli.Module` to wrap a typed factory and handler into the required `application.Option`.

## Grouping commands[​](#grouping-commands "Direct link to Grouping commands")

`Group()` returns a child `*Router`. Register `Cmd`s on that router to create namespaced subcommands. Groups can also have a `Description`, shared `Flags`, and shared `AppOptions`:

```
apiRouter := c.Group(cli.Group{Name: "api", Description: "Manage the API."})

apiRouter.Cmd(cli.Cmd{Name: "serve", Description: "Start the API server.", AppOptions: /* ... */})



dbRouter := c.Group(cli.Group{Name: "db", Description: "Manage the database."})

dbRouter.Cmd(cli.Cmd{Name: "migrate", Description: "Run database migrations.", AppOptions: /* ... */})
```

This produces commands like `myapp api serve` and `myapp db migrate`. Groups can be nested arbitrarily deep.

## Adding flags[​](#adding-flags "Direct link to Adding flags")

Declare flags on a `Cmd`, `Group`, or globally with `cli.WithFlag`. Each flag has a short name, a long name, an optional default, and an optional `CfgKey` to map the value to an arbitrary config path:

```
cli.Cmd{

    Name: "serve",

    Flags: []cli.Flag{

        {Short: "p", Long: "port", CfgKey: "httpserver.default.port", Default: "8080", Description: "port to listen on"},

    },

    AppOptions: /* ... */,

}
```

| Field         | Description                                            |
| ------------- | ------------------------------------------------------ |
| `Short`       | Single-character flag (`-p`)                           |
| `Long`        | Long-form flag (`--port`)                              |
| `Kind`        | Parse mode: `cli.FlagKindString` or `cli.FlagKindList` |
| `Default`     | Value used when the flag is absent                     |
| `CfgKey`      | If set, the value is also written to this config path  |
| `Description` | Human-readable description                             |

The flag value is always available at `cli.flags.<long>` in the config, with hyphens replaced by underscores. For example, `--dry-run` is available as `cli.flags.dry_run`. If `CfgKey` is set, the value is also written there.

String flags use `cli.FlagKindString`, which is also the default when `Kind` is omitted. When the same string flag is provided multiple times, the last matching short or long value wins.

List flags use `cli.FlagKindList` and collect all matching short and long values in order:

```
cli.Flag{

    Short:       "i",

    Long:        "include",

    Kind:        cli.FlagKindList,

    Description: "file or pattern to include",

}
```

```
myapp run --include users.json -i orders.json
```

This writes `[]string{"users.json", "orders.json"}` to `cli.flags.include`.

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

You can also unmarshal all values from `cli.flags` into a struct:

```
type ServeFlags struct {

    Port string `cfg:"port"`

}



flags, err := cli.UnmarshalFlags[ServeFlags](config)

if err != nil {

    return nil, err

}
```

## Positional arguments[​](#positional-arguments "Direct link to Positional arguments")

Commands can document the positional arguments they accept. This affects help output:

```
cli.Cmd{

    Name:        "import",

    Description: "Import one or more files.",

    Arguments:   cli.CmdArgumentsMultiple,

    AppOptions:  /* ... */,

}
```

| Value                      | Help usage              |
| -------------------------- | ----------------------- |
| `cli.CmdArgumentsNone`     | no positional arguments |
| `cli.CmdArgumentsSingle`   | `<arg>`                 |
| `cli.CmdArgumentsMultiple` | `<args...>`             |

Runtime arguments are written to `cli.args`. Read them with `cli.GetArguments`:

```
args, err := cli.GetArguments(config)

if err != nil {

    return nil, err

}
```

The selected command path is written to `cli.cmd`.

## Built-in help[​](#built-in-help "Direct link to Built-in help")

Pass `cli.WithHelp` to configure top-level help text and register the built-in `help` command:

```
c := cli.NewCli(

    cli.WithHelp("myapp", "Example CLI application."),

)
```

Users can request help with the `help` command or with `-h` / `--help`:

```
myapp help

myapp help api serve

myapp api serve --help

myapp api serve -h
```

Help output includes command and group descriptions, usage, flags, defaults, positional argument markers, and command examples. Unknown commands print an error and then the relevant help output. You can change help wrapping with `cli.WithHelpLineLength`; pass a negative value to disable wrapping.

## Default command[​](#default-command "Direct link to Default command")

Use `DefaultCmd` to register a fallback that runs when no explicit command is selected. This is useful when empty input should execute a command instead of showing help:

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

Unknown commands do not run the default command. They return an error and, when help is configured, print contextual help.

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



	"github.com/justtrackio/gosoline/pkg/application"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/cli"

	"github.com/justtrackio/gosoline/pkg/kernel"

	"github.com/justtrackio/gosoline/pkg/log"

)



func main() {

	c := cli.NewCli(

		cli.WithHelp("myapp", "Example CLI application."),

		cli.WithVersion("1.0.0"),

		cli.WithAppOptions(

			application.WithConfigFile("config.dist.yml", "yml"),

		),

	)



	// Register subcommands under the "api" group.

	apiRouter := c.Group(cli.Group{Name: "api", Description: "Manage the API."})

	apiRouter.Cmd(cli.Cmd{

		Name:        "serve",

		Description: "Start the API server.",

		Examples: []cli.CmdExample{

			{Description: "Start the API server on a custom port:", Args: "myapp api serve --port 9090"},

		},

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

	dbRouter := c.Group(cli.Group{Name: "db", Description: "Manage the database."})

	dbRouter.Cmd(cli.Cmd{

		Name:        "migrate",

		Description: "Run database migrations.",

		Arguments:   cli.CmdArgumentsMultiple,

		Examples: []cli.CmdExample{

			{Description: "Run migrations against production:", Args: "myapp db migrate --env prod"},

		},

		Flags: []cli.Flag{

			{Short: "e", Long: "env", CfgKey: "app.env", Default: "dev", Description: "target environment"},

			{Short: "i", Long: "include", Kind: cli.FlagKindList, Description: "migration file or pattern to include"},

		},

		AppOptions: []application.Option{

			application.WithModuleFactory("main", cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {

				env, err := config.GetString("app.env")

				if err != nil {

					return nil, err

				}



				args, err := cli.GetArguments(config)

				if err != nil {

					return nil, err

				}



				flags, err := cli.UnmarshalFlags[struct {

					Include []string `cfg:"include"`

				}](config)

				if err != nil {

					return nil, err

				}



				return func(ctx context.Context) error {

					logger.Info(ctx, "running migrations in env: %s", env)

					logger.Info(ctx, "migration arguments: %v", args)

					logger.Info(ctx, "included migrations: %v", flags.Include)



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
# Show top-level help

go run main.go help



# Show command help

go run main.go api serve --help



# Start the API server on a custom port

go run main.go api serve --port 9090



# Run migrations against production with positional args and a repeated list flag

go run main.go db migrate migrations/*.sql --env prod --include users -i orders



# Print the version

go run main.go version
```
