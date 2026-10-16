# log package

The gosoline logger is based upon a simple interface that uses handlers internally to enable fully customizable log handling.

## Functions[​](#functions "Direct link to Functions")

### [NewLogger()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/logger.go#L91)[​](#newlogger "Direct link to newlogger")

#### Usage[​](#usage "Direct link to Usage")

```
logger := log.NewLogger()
```

#### Description[​](#description "Direct link to Description")

Create a logger with no handlers and a real time clock. This provides an extended interface, including the `Option(opt ...Option) error` function to change the behaviour of the logger.

### [NewLoggerWithInterfaces()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/logger.go#L95)[​](#newloggerwithinterfaces "Direct link to newloggerwithinterfaces")

#### Usage[​](#usage-1 "Direct link to Usage")

```
logger := log.NewLoggerWithInterfaces(myClock, []log.Handler{handler1, handler2})
```

#### Description[​](#description-1 "Direct link to Description")

Provide a clock and some handlers when you create a new logger. Like [`NewLogger()`](#newlogger), this provides an extended interface, including the `Option(opt ...Option) error` function to change the behaviour of the logger.

### [InitContext()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/context.go#L80)[​](#initcontext "Direct link to initcontext")

#### Usage[​](#usage-2 "Direct link to Usage")

```
ctx = log.InitContext(ctx)
```

#### Description[​](#description-2 "Direct link to Description")

Return a new context capable of carrying (mutable) local and global logger fields.

### [AppendContextFields()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/context.go#L115C6-L115C25)[​](#appendcontextfields "Direct link to appendcontextfields")

#### Usage[​](#usage-3 "Direct link to Usage")

```
localCtx := log.AppendContextFields(ctx, map[string]any{

  "field": "value",

})
```

#### Description[​](#description-3 "Direct link to Description")

Appends fields to the existing **local** context fields, creating and returning a new context containing the merged fields.

caution

Any existing fields with the same key as any new field provided will be overwritten.

#### Related methods[​](#related-methods "Direct link to Related methods")

MutateContextFields()

Mutates **local** context fields if the context already contains fields which can be mutated. Otherwise, it initializes a new context able to carry fields in the future.

```
localCtx = log.MutateContextFields(localCtx, map[string]any{

	"field": "new_value",

})
```

AppendGlobalContextFields()

Appends fields to the existing **global** context fields, creating a new context containing the merged fields.

```
localCtx = log.AppendGlobalContextFields(globalCtx, map[string]any{

	"field": "new_value",

})
```

MutateGlobalContextFields()

Mutates **global** context fields if the context already contains fields which can be mutated. Otherwise, it initializes a new context able to carry fields in the future.

```
localCtx = log.MutateGlobalContextFields(globalCtx, map[string]any{

	"field": "new_value",

})
```

caution

Global fields override local fields when they have the same name.

### [ContextFieldsResolver()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/context.go#L209C6-L209C27)[​](#contextfieldsresolver "Direct link to contextfieldsresolver")

#### Usage[​](#usage-4 "Direct link to Usage")

```
localFields := log.ContextFieldsResolver(ctx)

print(localFields["field"])
```

#### Description[​](#description-4 "Direct link to Description")

Extracts the local and global fields from a context and returns a map.

#### Related methods[​](#related-methods-1 "Direct link to Related methods")

GlobalContextFieldsResolver()

Extracts the global fields from a context and returns a map.

```
localFields := log.GlobalContextFieldsResolver(ctx)

print(localFields["field"])
```

## Methods[​](#methods "Direct link to Methods")

### [Debug()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/logger.go#L119)[​](#debug "Direct link to debug")

```
logger.Debug(ctx, "Message")
```

#### Description[​](#description-5 "Direct link to Description")

Logs a message at the Debug log level.

#### Related methods[​](#related-methods-2 "Direct link to Related methods")

Info()

```
logger.Info(ctx, "Message")
```

Warn()

```
logger.Warn(ctx, "Message")
```

Error()

```
logger.Error(ctx, "Message")
```

### [WithContextFieldsResolver()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/options.go#L5)[​](#withcontextfieldsresolver "Direct link to withcontextfieldsresolver")

#### Usage[​](#usage-5 "Direct link to Usage")

```
if err := logger.Option(log.WithContextFieldsResolver(log.ContextFieldsResolver)); err != nil {

	panic(err)

}
```

#### Description[​](#description-6 "Direct link to Description")

Adds a context fields resolver to the logger.

### [WithFields()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/options.go#L13)[​](#withfields "Direct link to withfields")

#### Usage[​](#usage-6 "Direct link to Usage")

```
loggerWithFields := logger.WithFields(log.Fields{

	"b": true,

})
```

#### Description[​](#description-7 "Direct link to Description")

Adds global fields to the logger, which will be set on every log message.

### [WithHandlers()](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/options.go#L21)[​](#withhandlers "Direct link to withhandlers")

#### Usage[​](#usage-7 "Direct link to Usage")

```
	logHandler := log.NewHandlerIoWriter(cfg.New(), log.PriorityInfo, log.FormatterConsole, "main", "15:04:05.000", os.Stdout)

	loggerOptions := []log.Option{

		log.WithHandlers(logHandler),

	}
```

#### Description[​](#description-8 "Direct link to Description")

Adds additional handlers to the logger.

## Interfaces[​](#interfaces "Direct link to Interfaces")

### [Handler](https://github.com/justtrackio/gosoline/blob/v0.63.7/pkg/log/handler.go#L10)[​](#handler "Direct link to handler")

#### Definition[​](#definition "Direct link to Definition")

```
type Handler interface {

	ChannelLevel(name string) (level *int, err error)

	Level() int

	Log(ctx context.Context, timestamp time.Time, level int, msg string, args []any, err error, data Data) error

}
```

#### Description[​](#description-9 "Direct link to Description")

* `ChannelLevel(name string) (level *int, err error)` and `Level() int` are called on every log action to check if the handler should be applied.
* `Log` does the actual logging afterwards.

## Log configurations[​](#log-configurations "Direct link to Log configurations")

| setting             | description                                                    | default                                                                            |
| ------------------- | -------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| log.level           | default level for all handlers without an explicit level value | info                                                                               |
| log.handlers        | a map of handlers that will be called for every log message    | every logger gets a 'main' handler by default if there is no other handler defined |
| log.handlers.X.type | defines the type of the handler                                | -                                                                                  |

## Built-in handlers[​](#built-in-handlers "Direct link to Built-in handlers")

Gosoline has a couple of built-in handlers, which are ready to use out of the box:

### iowriter[​](#iowriter "Direct link to iowriter")

Multitool, which is able to write logs to everything which implements the `io.Writer` interface. Config options are:

| Setting           | Description                                                        | Default      |
| ----------------- | ------------------------------------------------------------------ | ------------ |
| level             | Levels of this and higher priority will get logged                 | info         |
| channels          | Messages logged into these channels will be handled                |              |
| formatter         | Which format should be used by this handler                        | console      |
| timestamp\_format | A golang time format string to control the format of the timestamp | 15:04:05.000 |
| writer            | Which io.writer implementation to use                              | stdout       |

#### Log to STDOUT[​](#log-to-stdout "Direct link to Log to STDOUT")

```
log:

  handlers:

    main:

      type: iowriter

      level: info

      channels: {}

      formatter: console

      timestamp_format: 15:04:05.000

      writer: stdout
```

#### Log to a file[​](#log-to-a-file "Direct link to Log to a file")

```
log:

  handlers:

    main:

      type: iowriter

      level: info

      channels: {}

      formatter: console

      timestamp_format: 15:04:05.000

      writer: file

      path: logs.log
```

### Metric[​](#metric "Direct link to Metric")

No configuration needed. Writes a metric data point for every warn and error log.

### Sentry[​](#sentry "Direct link to Sentry")

No configuration needed. Publishes every logged error to Sentry.
