# Use loggers

With gosoline, there are multiple ways to configure and use a logger.

## Implement a logger directly in code[​](#implement-a-logger-directly-in-code "Direct link to Implement a logger directly in code")

One way to implement a logger is using log functions, without appealing to an external configuration:

main.go

```
package main



import (

	"context"

	"os"



	// 1

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

)



func main() {

	ctx := context.Background()

	// 2

	logHandler := log.NewHandlerIoWriter(

		cfg.New(), log.PriorityInfo, log.FormatterConsole, "main", "15:04:05.000", os.Stdout,

	)



	// 3

	loggerOptions := []log.Option{log.WithHandlers(logHandler)}



	// 4

	logger := log.NewLogger()



	// 5

	if err := logger.Option(loggerOptions...); err != nil {

		logger.Error(ctx, "Failed to apply logger options: %w", err)

		os.Exit(1)

	}



	logger.Info(ctx, "Message")

}
```

Here, you:

1. Import the `cfg` package.
2. Create a new handler for sending logs to STDOUT.
3. Create a log option that includes your handler.
4. Create your logger.
5. Apply the options.

## Implement a logger with an app configuration[​](#implement-a-logger-with-an-app-configuration "Direct link to Implement a logger with an app configuration")

If you have an application, you can create a logger from your app's configuration.

First, configure your logger in your configuration. For example, in a config file:

config.dist.yml

```
app:

  env: dev

  name: hello-world

  namespace: "{app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}"

  tags:

    project: gosoline

    family: get-started

    group: grp



log:

    level: info

    handlers:

        main:

            type: iowriter

            channels:

                metrics:

                    level: error

                http:

                    level: error

            formatter: console

            level: info

            timestamp_format: 15:04:05.000

            writer: stdout
```

Then, implement your logger alongside your application:

main.go

```
package main



import (

	"context"



	"github.com/justtrackio/gosoline/pkg/application"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/kernel"

	"github.com/justtrackio/gosoline/pkg/log"

)



func NewHelloWorldModule(ctx context.Context, config cfg.Config, logger log.Logger) (kernel.Module, error) {

	return &HelloWorldModule{

		logger: logger.WithChannel("hello-world"),

	}, nil

}



type HelloWorldModule struct {

	logger log.Logger

}



func (h HelloWorldModule) Run(ctx context.Context) error {

	h.logger.Info(ctx, "Hello World")



	return nil

}



func main() {

	application.Run(

		application.WithConfigFile("config.dist.yml", "yml"),

		application.WithModuleFactory("hello-world", NewHelloWorldModule),

	)

}
```

## Conclusion[​](#conclusion "Direct link to Conclusion")

In this guide, you've learned multiple ways to implement a logger with gosoline.

Check out these resources to learn more about logging with gosoline:

* [Implement a log handler](/docs/how-to/logging/implement-a-log-handler.md)
* [API reference for the log package](/docs/reference/package-log.md)
