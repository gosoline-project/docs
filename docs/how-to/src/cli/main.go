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
		ModuleFactory: cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {
			port := config.GetString("httpserver.default.port")
			return func(ctx context.Context) error {
				logger.Info(ctx, "api server listening on port %s", port)
				<-ctx.Done()
				return nil
			}, nil
		}),
	})

	// Register subcommands under the "db" group.
	dbRouter := c.Group(cli.Group{Name: "db"})
	dbRouter.Cmd(cli.Cmd{
		Name: "migrate",
		Flags: []cli.Flag{
			{Short: "e", Long: "env", CfgKey: "app.env", Default: "dev", Description: "target environment"},
		},
		ModuleFactory: cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {
			env := config.GetString("app.env")
			return func(ctx context.Context) error {
				logger.Info(ctx, "running migrations in env: %s", env)
				return nil
			}, nil
		}),
	})

	// Fallback when no command is matched.
	c.DefaultCmd(cli.Cmd{
		ModuleFactory: cli.WithRunFunc(func(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.ModuleRunFunc, error) {
			return func(ctx context.Context) error {
				fmt.Println("Usage: myapp <command> [flags]")
				fmt.Println("Commands:")
				fmt.Println("  api serve    Start the API server")
				fmt.Println("  db migrate   Run database migrations")
				fmt.Println("  version      Print the version")
				return nil
			}, nil
		}),
	})

	c.Run()
}
