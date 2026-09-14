package main

import (
	"context"
	"embed"
	"fmt"

	"github.com/gosoline-project/httpserver"
	"github.com/gosoline-project/sqlh"
	"github.com/gosoline-project/sqlr"
	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

// snippet-start: transaction types
type Post struct {
	sqlr.Entity[int64]
	Title  string `db:"title"`
	Body   string `db:"body"`
	Status string `db:"status"`
}

type PostCreateInput struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

type PostOutput struct {
	Id     int64  `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

// snippet-end: transaction types

// snippet-start: transaction handler
type PostHandler struct {
	runner *sqlh.TxRunner
}

func NewPostHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*PostHandler, error) {
	runner, err := sqlh.NewTxRunner(ctx, config, logger, "default")
	if err != nil {
		return nil, err
	}

	return &PostHandler{runner: runner}, nil
}

func (h *PostHandler) CreatePost(ctx context.Context, input *PostCreateInput) (PostOutput, error) {
	return h.runner.RunValue(ctx, input, func(ctx context.Context, tx sqlr.TTx, input *PostCreateInput) (PostOutput, error) {
		post := &Post{
			Title:  input.Title,
			Body:   input.Body,
			Status: "draft",
		}

		if _, err := tx.Q().Into("posts").Records(post).Exec(tx); err != nil {
			return PostOutput{}, fmt.Errorf("failed to create post: %w", err)
		}

		return PostOutput{
			Id:     post.Id,
			Title:  post.Title,
			Body:   post.Body,
			Status: post.Status,
		}, nil
	})
}

// snippet-end: transaction handler

//go:embed config.dist.yml
var config embed.FS

// snippet-start: transaction register
func main() {
	configBytes, err := config.ReadFile("config.dist.yml")
	if err != nil {
		panic(err)
	}

	application.New(
		application.WithConfigBytes(configBytes, "yml"),
		application.WithLoggerHandlersFromConfig,
		application.WithModuleFactory("http", httpserver.NewServer(
			"default",
			func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
				router.HandleWith(httpserver.With(NewPostHandler, func(router *httpserver.Router, handler *PostHandler) {
					router.POST("/v1/posts", httpserver.Bind(handler.CreatePost))
				}))

				return nil
			},
		)),
	).Run()
}

// snippet-end: transaction register
