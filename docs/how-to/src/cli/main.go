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
