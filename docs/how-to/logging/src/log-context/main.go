// snippet-start: imports
package main

import (
	"context"
	"fmt"

	"github.com/gosoline-project/httpserver"
	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

// snippet-end: imports

// snippet-start: main
func main() {
	httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
		router.HandleWith(httpserver.With(NewTodoHandler, func(r *httpserver.Router, h *TodoHandler) {
			r.POST("/todos", httpserver.Bind(h.CreateTodo))
			r.PUT("/todos/:id", httpserver.Bind(h.UpdateTodo))
		}))

		return nil
	},
		application.WithConfigFile("config.dist.yml", "yml"),
	)
}

// snippet-end: main
