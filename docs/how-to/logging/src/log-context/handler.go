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
	// highlight-next-line
	logger log.Logger
	mu     sync.Mutex
	nextId int
	todos  map[int]*Todo
}

// snippet-end: crud handler v0

// snippet-start: new todo crud handler
func NewTodoHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*TodoHandler, error) {
	return &TodoHandler{
		// highlight-next-line
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

	// highlight-start
	localctx := log.InitContext(ctx)
	text := truncate(localctx, input.Text)
	// highlight-end

	todo := &Todo{
		Id:      h.nextId,
		Text:    text,
		DueDate: input.DueDate,
	}
	h.todos[todo.Id] = todo
	h.nextId++

	// highlight-next-line
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

	// highlight-start
	localctx := log.InitContext(ctx)
	todo.Text = truncate(localctx, input.Text)
	// highlight-end

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
