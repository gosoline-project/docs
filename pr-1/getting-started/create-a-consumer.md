# Create a consumer

One of the primary use cases for gosoline is to create a message queue consumer. In this tutorial, you'll do just that!

To build a consumer of async message queues you'll' implement the `ConsumerCallback` interface of the `stream` package.

## Before you begin[​](#before-you-begin "Direct link to Before you begin")

Before you begin, make sure you have [Golang](https://go.dev/doc/install) installed on your machine.

## Set up your file structure[​](#set-up-your-file-structure "Direct link to Set up your file structure")

First, you need to set up the following file structure:

```
consumer/

├── consumer.go

├── main.go

├── events.json

└── config.dist.yml
```

For example, in Unix, run:

```
mkdir consumer; cd consumer

touch consumer.go

touch main.go

touch events.json

touch config.dist.yml
```

Those are all the files you need to build your first consumer with gosoline! Next, you'll implement each of these files, starting with `consumer.go`.

## Implement consumer.go[​](#implement-consumergo "Direct link to Implement consumer.go")

In `consumer.go`, add the following code:

consumer.go

```
package main



import (

	"context"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/stream"

)



type Input struct {

	Id   string `json:"id"`

	Body string `json:"body"`

}



type Consumer struct {

	logger log.Logger

}



func NewConsumer(ctx context.Context, config cfg.Config, logger log.Logger) (stream.ConsumerCallback[Input], error) {

	return &Consumer{

		logger: logger,

	}, nil

}



func (c Consumer) Consume(ctx context.Context, input Input, attributes map[string]string) (bool, error) {

	c.logger.Info(ctx, "got input with id %q and body %q", input.Id, input.Body)



	return true, nil

}
```

Now, you'll walkthrough this file in detail to learn how it works.

### Import dependencies[​](#import-dependencies "Direct link to Import dependencies")

At the top of `consumer.go`, you declared the package and imported some dependencies:

```
package main



import (

	"context"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/stream"

)
```

Here, you declared the package as `main`. Then, you imported the `context` module along with three gosoline dependencies:

* [`cfg`](/docs/pr-1/reference/package-cfg.md)
* [`log`](/docs/pr-1/reference/package-log.md)
* `stream`

### Implement your data structs[​](#implement-your-data-structs "Direct link to Implement your data structs")

Then, you created an `Input` struct and a `Consumer` struct:

```
// 1

type Input struct {

	Id   string `json:"id"`

	Body string `json:"body"`

}



// 2

type Consumer struct {

	logger log.Logger

}
```

You'll use these to:

1. Bind data from the message queue. Note that you read an `id` and `body` from Json keys.
2. Store logger information about your consumer.

### Implement `Consumer` methods[​](#implement-consumer-methods "Direct link to implement-consumer-methods")

Next, you implemented some methods for the `Consumer`:

```
// 1

func NewConsumer(ctx context.Context, config cfg.Config, logger log.Logger) (stream.ConsumerCallback, error) {

	return &Consumer{

		logger: logger,

	}, nil

}



// 2

func (c Consumer) GetModel(attributes map[string]any) any {

	return &Input{}

}



// 3

func (c Consumer) Consume(ctx context.Context, model any, attributes map[string]any) (bool, error) {

	input := model.(*Input)



	c.logger.WithContext(ctx).Info("got input with id %q and body %q", input.Id, input.Body)



	return true, nil

}
```

Here, you implemented:

1. A constructor for creating new `Consumer` objects. This implements the `stream.ConsumerCallbackFactory` type and is used to add the callback to your application.
2. `GetModal()`, a method that returns the expected input model struct which is used to unmarshal the data.
3. `Consume()`, a method that loads the model (`Input`) with data, logs the data, and returns `true` because it successfully handled the message. This is called for every incoming message.

Together, these methods implement the `ConsumerCallback` interface.

## Implement `main.go`[​](#implement-maingo "Direct link to implement-maingo")

In `main.go`, add the following code:

main.go

```
package main



import "github.com/justtrackio/gosoline/pkg/application"



func main() {

	application.RunConsumer(NewConsumer,

		application.WithConfigFile("config.dist.yml", "yml"),

	)

}
```

Here, you execute your consumer. `RunConsumer()` expects a parameter of the type `stream.ConsumerCallbackFactory` to create and run the consumer. `NewConsumer()` implements this interface.

## Configure your consumer[​](#configure-your-consumer "Direct link to Configure your consumer")

In `config.dist.yml`, configure your server:

config.dist.yml

```
app:

  env: dev

  name: consumer

  namespace: "{app.tags.project}.{app.env}.{app.tags.family}.{app.tags.group}"

  tags:

    project: gosoline

    family: how-to

    group: grp



stream:

  input:

    consumer:

      type: file

      filename: events.json
```

Here, you set some minimal configurations for your web server. By default, the gosoline expects that there is an input configured with the config key `stream.input.consumer`. This defines the input source. In this tutorial, you'll use a file as source with the configured filename, "events.json".

## Add your data[​](#add-your-data "Direct link to Add your data")

In `events.json`, add some mock events stream data:

```
{"body": "{\"id\": \"1a0a960f-f04f-4c41-9b9a-a5ca0e2637b2\", \"body\": \"Lorem ipsum dolor sit amet.\"}"}
```

Now, the final step is to test it to confirm that it works as expected.

## Test your consumer[​](#test-your-consumer "Direct link to Test your consumer")

In the `consumer` directory, run:

```
go mod init consumer/m

go mod tidy

go run .
```

In the output, you'll see a log that indicates your consumer handled the event data:

```
10:23:57.648 consumerCallback info    got input with id "1a0a960f-f04f-4c41-9b9a-a5ca0e2637b2" and body "Lorem ipsum dolor sit amet."  application: consumer
```

## Conclusion[​](#conclusion "Direct link to Conclusion")

That's it! You created your first gosoline message queue consumer.
