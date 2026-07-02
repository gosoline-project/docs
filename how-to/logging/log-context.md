# Use context with logs

The [go Context](https://pkg.go.dev/context) carries data from the moment the server receives an inbound request to the moment the server makes an outbound request. This means you can use it to propagate data between services and processes. With gosoline, you can use log functions to store data from the request lifecycle in the Context and attach that data to logs to provide more details.

In this guide, you'll add some logs to an HTTP server used for managing a "To do list".

## Before you begin[​](#before-you-begin "Direct link to Before you begin")

Before you begin, make sure you have [Golang](https://go.dev/doc/install) installed on your machine.

## Truncate todo text[​](#truncate-todo-text "Direct link to Truncate todo text")

With this service, users can create and update todos. For the purposes of this tutorial, you'll add some new logic. Instead of accepting any text for a todo, you'll limit the length of that string to prevent users from posting huge amounts of text in their todos.

In `handler.go`, add a new function:

handler.go

```
package main



import (

	"context"

	"fmt"

	"strconv"

	"sync"

	"time"



	"github.com/gosoline-project/httpserver"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

)



type CreateTodoInput struct {

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



type UpdateTodoInput struct {

	Id   int    `uri:"id"`

	Text string `json:"text"`

}



type Todo struct {

	Id      int       `json:"id"`

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



// snippet-start: crud handler v0

type TodoHandler struct {

	logger log.Logger

	mu     sync.Mutex

	nextId int

	todos  map[int]*Todo

}



// snippet-end: crud handler v0



// snippet-start: new todo crud handler

func NewTodoHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*TodoHandler, error) {

	return &TodoHandler{

		logger: logger,

		nextId: 1,

		todos:  map[int]*Todo{},

	}, nil

}



// snippet-end: new todo crud handler



// snippet-start: truncate

func truncate(ctx context.Context, text string) string {

	r := []rune(text)

	length := len(r)



	log.MutateContextFields(ctx, map[string]any{

		"original_length": length,

	})



	if length > 50 {

		text = string(r[:50]) + "..."

	}



	return text

}



// snippet-end: truncate



func (h *TodoHandler) CreateTodo(ctx context.Context, input *CreateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	localctx := log.InitContext(ctx)

	text := truncate(localctx, input.Text)



	todo := &Todo{

		Id:      h.nextId,

		Text:    text,

		DueDate: input.DueDate,

	}

	h.todos[todo.Id] = todo

	h.nextId++



	h.logger.Info(localctx, "creating new task due at %v", todo.DueDate)



	return httpserver.NewJsonResponse(todo), nil

}



func (h *TodoHandler) UpdateTodo(ctx context.Context, input *UpdateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	todo, ok := h.todos[input.Id]

	if !ok {

		return httpserver.NewStatusResponse(404), fmt.Errorf("todo %d not found", input.Id)

	}



	localctx := log.InitContext(ctx)

	todo.Text = truncate(localctx, input.Text)



	return httpserver.NewJsonResponse(todo), nil

}



// snippet-start: parse id helper

func parseId(id string) (int, error) {

	n, err := strconv.Atoi(id)

	if err != nil {

		return 0, fmt.Errorf("invalid id: %s", id)

	}

	return n, nil

}



// snippet-end: parse id helper
```

In this function, you:

1. Accept a `Context` and a string as arguments
2. Capture the length of the string
3. Mutate the `Context` to store the original length of the string
4. Truncate the string if it is longer than 50 runes.
5. Return the potentially truncated string

For this tutorial, the important thing to pay attention to is where you mutate the `Context`:

```
log.MutateContextFields(ctx, map[string]any{

	"original_length": length,

})
```

With Gosoline, you can initialize specific fields that you can use with a `Logger`. (You'll do this in the next step.) Once those fields are initialized, you can append or mutate the fields as you've done here.

info

Read more about appending and mutating context fields in our [log package reference](/docs/reference/package-log.md#appendcontextfields).

## Use your new function[​](#use-your-new-function "Direct link to Use your new function")

Now that you have a function that can truncate todo text, use it when creating a todo:

handler.go

```
package main



import (

	"context"

	"fmt"

	"strconv"

	"sync"

	"time"



	"github.com/gosoline-project/httpserver"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

)



type CreateTodoInput struct {

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



type UpdateTodoInput struct {

	Id   int    `uri:"id"`

	Text string `json:"text"`

}



type Todo struct {

	Id      int       `json:"id"`

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



// snippet-start: crud handler v0

type TodoHandler struct {

	logger log.Logger

	mu     sync.Mutex

	nextId int

	todos  map[int]*Todo

}



// snippet-end: crud handler v0



// snippet-start: new todo crud handler

func NewTodoHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*TodoHandler, error) {

	return &TodoHandler{

		logger: logger,

		nextId: 1,

		todos:  map[int]*Todo{},

	}, nil

}



// snippet-end: new todo crud handler



// snippet-start: truncate

func truncate(ctx context.Context, text string) string {

	r := []rune(text)

	length := len(r)



	log.MutateContextFields(ctx, map[string]any{

		"original_length": length,

	})



	if length > 50 {

		text = string(r[:50]) + "..."

	}



	return text

}



// snippet-end: truncate



func (h *TodoHandler) CreateTodo(ctx context.Context, input *CreateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	localctx := log.InitContext(ctx)

	text := truncate(localctx, input.Text)



	todo := &Todo{

		Id:      h.nextId,

		Text:    text,

		DueDate: input.DueDate,

	}

	h.todos[todo.Id] = todo

	h.nextId++



	h.logger.Info(localctx, "creating new task due at %v", todo.DueDate)



	return httpserver.NewJsonResponse(todo), nil

}



func (h *TodoHandler) UpdateTodo(ctx context.Context, input *UpdateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	todo, ok := h.todos[input.Id]

	if !ok {

		return httpserver.NewStatusResponse(404), fmt.Errorf("todo %d not found", input.Id)

	}



	localctx := log.InitContext(ctx)

	todo.Text = truncate(localctx, input.Text)



	return httpserver.NewJsonResponse(todo), nil

}



// snippet-start: parse id helper

func parseId(id string) (int, error) {

	n, err := strconv.Atoi(id)

	if err != nil {

		return 0, fmt.Errorf("invalid id: %s", id)

	}

	return n, nil

}



// snippet-end: parse id helper
```

And when updating a todo:

handler.go

```
package main



import (

	"context"

	"fmt"

	"strconv"

	"sync"

	"time"



	"github.com/gosoline-project/httpserver"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

)



type CreateTodoInput struct {

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



type UpdateTodoInput struct {

	Id   int    `uri:"id"`

	Text string `json:"text"`

}



type Todo struct {

	Id      int       `json:"id"`

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



// snippet-start: crud handler v0

type TodoHandler struct {

	logger log.Logger

	mu     sync.Mutex

	nextId int

	todos  map[int]*Todo

}



// snippet-end: crud handler v0



// snippet-start: new todo crud handler

func NewTodoHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*TodoHandler, error) {

	return &TodoHandler{

		logger: logger,

		nextId: 1,

		todos:  map[int]*Todo{},

	}, nil

}



// snippet-end: new todo crud handler



// snippet-start: truncate

func truncate(ctx context.Context, text string) string {

	r := []rune(text)

	length := len(r)



	log.MutateContextFields(ctx, map[string]any{

		"original_length": length,

	})



	if length > 50 {

		text = string(r[:50]) + "..."

	}



	return text

}



// snippet-end: truncate



func (h *TodoHandler) CreateTodo(ctx context.Context, input *CreateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	localctx := log.InitContext(ctx)

	text := truncate(localctx, input.Text)



	todo := &Todo{

		Id:      h.nextId,

		Text:    text,

		DueDate: input.DueDate,

	}

	h.todos[todo.Id] = todo

	h.nextId++



	h.logger.Info(localctx, "creating new task due at %v", todo.DueDate)



	return httpserver.NewJsonResponse(todo), nil

}



func (h *TodoHandler) UpdateTodo(ctx context.Context, input *UpdateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	todo, ok := h.todos[input.Id]

	if !ok {

		return httpserver.NewStatusResponse(404), fmt.Errorf("todo %d not found", input.Id)

	}



	localctx := log.InitContext(ctx)

	todo.Text = truncate(localctx, input.Text)



	return httpserver.NewJsonResponse(todo), nil

}



// snippet-start: parse id helper

func parseId(id string) (int, error) {

	n, err := strconv.Atoi(id)

	if err != nil {

		return 0, fmt.Errorf("invalid id: %s", id)

	}

	return n, nil

}



// snippet-end: parse id helper
```

Here, you first call `log.InitContext()`. This function creates two sets of log-related fields in the `Context`:

* `localFields`: These fields are limited to the application in which they are set. They are not propagated to downstream services in any way.
* `globalFields`: These fields aren't limited to the application in which they are set. They are propagated to downstream services.

Then, it returns a `Context` in which these local and global fields are mutable. You pass this `Context` as the first parameter to `truncate()`.

:::info Technical Detail

Actually, this call to `log.InitContext()` is not required because gosoline will have already initialized the `Context` earlier in the request lifecycle (specifically, in the HTTP middleware). In this case, the `ctx` you pass to `log.InitContext()` is returned, unchanged. Therefore, `localctx` and `ctx` are the same, so you could have passed `ctx` to `truncate()` instead.

However, this example illustrates where to call `log.InitContext()` if you were to create or receive a `Context` from somewhere else (e.g., in a background job or CLI command). If you initialized the `Context` inside `truncate()`, the log-related fields would go out of scope when the function returns. Instead, you initialize the `Context` and pass it in, so you can make use of the log-related fields later.

:::

## Use your `Context` with logs[​](#use-your-context-with-logs "Direct link to use-your-context-with-logs")

If you run your service now, you'll see the results of your work. Gosoline has some built-in logs that will show your `Context` fields. However, you can also manually add the `Context` to a new logger.

First, add a logger to your `TodoHandler`:

handler.go

```
package main



import (

	"context"

	"fmt"

	"strconv"

	"sync"

	"time"



	"github.com/gosoline-project/httpserver"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

)



type CreateTodoInput struct {

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



type UpdateTodoInput struct {

	Id   int    `uri:"id"`

	Text string `json:"text"`

}



type Todo struct {

	Id      int       `json:"id"`

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



// snippet-start: crud handler v0

type TodoHandler struct {

	logger log.Logger

	mu     sync.Mutex

	nextId int

	todos  map[int]*Todo

}



// snippet-end: crud handler v0



// snippet-start: new todo crud handler

func NewTodoHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*TodoHandler, error) {

	return &TodoHandler{

		logger: logger,

		nextId: 1,

		todos:  map[int]*Todo{},

	}, nil

}



// snippet-end: new todo crud handler



// snippet-start: truncate

func truncate(ctx context.Context, text string) string {

	r := []rune(text)

	length := len(r)



	log.MutateContextFields(ctx, map[string]any{

		"original_length": length,

	})



	if length > 50 {

		text = string(r[:50]) + "..."

	}



	return text

}



// snippet-end: truncate



func (h *TodoHandler) CreateTodo(ctx context.Context, input *CreateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	localctx := log.InitContext(ctx)

	text := truncate(localctx, input.Text)



	todo := &Todo{

		Id:      h.nextId,

		Text:    text,

		DueDate: input.DueDate,

	}

	h.todos[todo.Id] = todo

	h.nextId++



	h.logger.Info(localctx, "creating new task due at %v", todo.DueDate)



	return httpserver.NewJsonResponse(todo), nil

}



func (h *TodoHandler) UpdateTodo(ctx context.Context, input *UpdateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	todo, ok := h.todos[input.Id]

	if !ok {

		return httpserver.NewStatusResponse(404), fmt.Errorf("todo %d not found", input.Id)

	}



	localctx := log.InitContext(ctx)

	todo.Text = truncate(localctx, input.Text)



	return httpserver.NewJsonResponse(todo), nil

}



// snippet-start: parse id helper

func parseId(id string) (int, error) {

	n, err := strconv.Atoi(id)

	if err != nil {

		return 0, fmt.Errorf("invalid id: %s", id)

	}

	return n, nil

}



// snippet-end: parse id helper
```

Now, you can make use of this logger in any of the handler's methods.

Next, when you initialize the handler, pass a logger:

handler.go

```
package main



import (

	"context"

	"fmt"

	"strconv"

	"sync"

	"time"



	"github.com/gosoline-project/httpserver"

	"github.com/justtrackio/gosoline/pkg/cfg"

	"github.com/justtrackio/gosoline/pkg/log"

)



type CreateTodoInput struct {

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



type UpdateTodoInput struct {

	Id   int    `uri:"id"`

	Text string `json:"text"`

}



type Todo struct {

	Id      int       `json:"id"`

	Text    string    `json:"text"`

	DueDate time.Time `json:"dueDate"`

}



// snippet-start: crud handler v0

type TodoHandler struct {

	logger log.Logger

	mu     sync.Mutex

	nextId int

	todos  map[int]*Todo

}



// snippet-end: crud handler v0



// snippet-start: new todo crud handler

func NewTodoHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*TodoHandler, error) {

	return &TodoHandler{

		logger: logger,

		nextId: 1,

		todos:  map[int]*Todo{},

	}, nil

}



// snippet-end: new todo crud handler



// snippet-start: truncate

func truncate(ctx context.Context, text string) string {

	r := []rune(text)

	length := len(r)



	log.MutateContextFields(ctx, map[string]any{

		"original_length": length,

	})



	if length > 50 {

		text = string(r[:50]) + "..."

	}



	return text

}



// snippet-end: truncate



func (h *TodoHandler) CreateTodo(ctx context.Context, input *CreateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	localctx := log.InitContext(ctx)

	text := truncate(localctx, input.Text)



	todo := &Todo{

		Id:      h.nextId,

		Text:    text,

		DueDate: input.DueDate,

	}

	h.todos[todo.Id] = todo

	h.nextId++



	h.logger.Info(localctx, "creating new task due at %v", todo.DueDate)



	return httpserver.NewJsonResponse(todo), nil

}



func (h *TodoHandler) UpdateTodo(ctx context.Context, input *UpdateTodoInput) (httpserver.Response, error) {

	h.mu.Lock()

	defer h.mu.Unlock()



	todo, ok := h.todos[input.Id]

	if !ok {

		return httpserver.NewStatusResponse(404), fmt.Errorf("todo %d not found", input.Id)

	}



	localctx := log.InitContext(ctx)

	todo.Text = truncate(localctx, input.Text)



	return httpserver.NewJsonResponse(todo), nil

}



// snippet-start: parse id helper

func parseId(id string) (int, error) {

	n, err := strconv.Atoi(id)

	if err != nil {

		return 0, fmt.Errorf("invalid id: %s", id)

	}

	return n, nil

}



// snippet-end: parse id helper
```

Finally, in `CreateTodo()`, you can use the logger:

```


h.logger.Info(localctx, "creating new task due at %v", todo.DueDate)
```

Here, you pass the `Context` to the logger.

## Test your work[​](#test-your-work "Direct link to Test your work")

Now it's time to test your work.

### Run your server[​](#run-your-server "Direct link to Run your server")

Navigate to the project directory and spin up your server:

```
go run .
```

You'll see logs of your server running.

### Make requests[​](#make-requests "Direct link to Make requests")

In another shell, make requests to your service. For example, create a todo:

```
curl -d '{"text":"do it!", "dueDate":"2023-09-08T15:00:00Z"}' -H "Content-Type: application/json" -X POST localhost:8088/todos
```

Update the todo:

```
curl -d '{"text":"do it!!!"}' -H "Content-Type: application/json" -X PUT localhost:8088/todos/1
```

In your logs, you should see the `original_length` field you added in the first step:

```
13:32:05.145 http    info    POST /todos HTTP/1.1

original_length: 115 
```

This is included in the log because we automatically resolve the local and global fields and include them in the log output.

If you need to create a new logger, you have to resolve the fields yourself. However, we've made this easy for you. Just pass the context to the logger:

```
logger := log.NewLogger()

logger.Info(ctx, "My message with context")
```

## Conclusion[​](#conclusion "Direct link to Conclusion")

Great work! In this tutorial, you used Gosoline to add some context to your logs.
