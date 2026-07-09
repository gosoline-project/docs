# Implement a log handler

With the `cfg` and `log` packages, you can implement a handler and make it available via config.

## Before you begin[​](#before-you-begin "Direct link to Before you begin")

Here is a preview of all the code you'll cover in this guide:

main.go

```
package main



import (

	"context"

	"fmt"

	"time"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/mdl"

)



type MyCustomHandlerSettings struct {

	Channel string `cfg:"channel"`

}



type MyCustomHandler struct {

	channel string

}



func (h *MyCustomHandler) ChannelLevel(name string) (level *int, err error) {

	if name == h.channel {

		return mdl.Box(log.PriorityDebug), nil

	}



	return nil, nil

}



func (h *MyCustomHandler) Level() int {

	return log.PriorityInfo

}



func (h *MyCustomHandler) Log(ctx context.Context, timestamp time.Time, level int, msg string, args []any, err error, data log.Data) error {

	fmt.Printf("%s happened at %s", msg, timestamp.Format(time.RFC822))



	return nil

}



func MyCustomHandlerFactory(config cfg.Config, name string) (log.Handler, error) {

	settings := &MyCustomHandlerSettings{}

	if err := log.UnmarshalHandlerSettingsFromConfig(config, name, settings); err != nil {

		return nil, fmt.Errorf("can not unmarshal handler settings: %w", err)

	}



	return &MyCustomHandler{

		channel: settings.Channel,

	}, nil

}



func main() {

	log.AddHandlerFactory("my-custom-handler", MyCustomHandlerFactory)

}
```

## Implement your custom log handler[​](#implement-your-custom-log-handler "Direct link to Implement your custom log handler")

### Import your gosoline dependencies[​](#import-your-gosoline-dependencies "Direct link to Import your gosoline dependencies")

Add the following imports to your code:

```
import (

	"context"

	"fmt"

	"time"



	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

	"github.com/justtrackio/gosoline/pkg/mdl"

)
```

Here, you imported some standard dependencies, along with the `cfg` and `log` packages from gosoline.

### Create a settings struct[​](#create-a-settings-struct "Direct link to Create a settings struct")

Create a struct that stores log settings:

```
type MyCustomHandlerSettings struct {

	Channel string `cfg:"channel"`

}
```

`cfg:` defines the key you'll use to bind the `channel` value from the configuration object to this struct.

### Create a new handler[​](#create-a-new-handler "Direct link to Create a new handler")

```
type MyCustomHandler struct {

	channel string

}
```

This handler struct stores the log `channel`. To use your handler, you must implement the required methods of the [`Handler`](/docs/pr-1/reference/package-log.md#handler) interface.

### Implement the `ChannelLevel` method[​](#implement-the-channellevel-method "Direct link to implement-the-channellevel-method")

Create a function which returns the log level for a given channel name:

```
func (h *MyCustomHandler) ChannelLevel(name string) (level *int, err error) {

	if name == h.channel {

		return mdl.Box(log.PriorityDebug), nil

	}



	return nil, nil

}
```

This checks if the requested channel is the channel we want to override, and, if so, returns a hardcoded value. Otherwise, we return nothing to indicate that the return value of the `Level` method should be used.

### Implement the `Level` method[​](#implement-the-level-method "Direct link to implement-the-level-method")

Create a getter for the log priority level:

```
func (h *MyCustomHandler) Level() int {

	return log.PriorityInfo

}
```

Here, you'll return the info level priority.

### Implement the `Log` method[​](#implement-the-log-method "Direct link to implement-the-log-method")

Create a getter for the log message:

```
func (h *MyCustomHandler) Log(ctx context.Context, timestamp time.Time, level int, msg string, args []any, err error, data log.Data) error {

	fmt.Printf("%s happened at %s", msg, timestamp.Format(time.RFC822))



	return nil

}
```

Here, you accept, among other things, a `msg` string and a `timestamp`. Then, you print a formatted log message, using these values.

### Create a handler factory[​](#create-a-handler-factory "Direct link to Create a handler factory")

Create a custom handler factory:

```
func MyCustomHandlerFactory(config cfg.Config, name string) (log.Handler, error) {

	// 1

	settings := &MyCustomHandlerSettings{}



	// 2

	if err := log.UnmarshalHandlerSettingsFromConfig(config, name, settings); err != nil {

		return nil, fmt.Errorf("can not unmarshal handler settings: %w", err)

	}



	// 3

	return &MyCustomHandler{

		channel: settings.Channel,

	}, nil

}
```

This accepts a configuration and a name and returns a Handler. You accomplish this with the following steps:

1. Initialize `settings` to a new, empty `MyCustomHandlerSettings` struct, which you defined in the last step.
2. Store the configuration values from the configuration in the `settings` struct.
3. Create a new `MyCustomHandler`, using the `settings.Channel`.

### Add your handler factory[​](#add-your-handler-factory "Direct link to Add your handler factory")

In `main()`, or wherever is most appropriate for your application, add your custom handler factory:

```
log.AddHandlerFactory("my-custom-handler", MyCustomHandlerFactory)
```

This sets the handler type to "my-custom-handler".

## Conclusion[​](#conclusion "Direct link to Conclusion")

That's it! In this guide, you:

* Implemented a custom Handler.
* Created a handler factory.
* Added your factory to the logging configuration.

Check out these resources to learn more about logging with gosoline:

* [Use loggers](/docs/pr-1/how-to/logging/use-loggers.md)
* [API reference for the log package](/docs/pr-1/reference/package-log.md)
