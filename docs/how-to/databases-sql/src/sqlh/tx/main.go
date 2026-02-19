package main

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/gosoline-project/httpserver"
	"github.com/gosoline-project/sqlc"
	"github.com/gosoline-project/sqlh"
	"github.com/gosoline-project/sqlr"
	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

// snippet-start: post entities
type Post struct {
	sqlr.Entity[int64]
	AuthorID int64  `db:"author_id"`
	Title    string `db:"title"`
	Body     string `db:"body"`
	Status   string `db:"status"`
}

type PostCreateInput struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

type PostOutput struct {
	Id        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// snippet-end: post entities

// snippet-start: with tx handler
type PostHandler struct {
}

func NewPostHandler() httpserver.HandlerFactory[PostHandler] {
	return func(ctx context.Context, config cfg.Config, logger log.Logger) (*PostHandler, error) {
		return &PostHandler{}, nil
	}
}

// snippet-end: with tx handler

// snippet-start: bind tx handler
func (h *PostHandler) HandleCreatePost(cttx sqlc.Tx, input *PostCreateInput) (httpserver.Response, error) {
	// The transaction is automatically managed — commit on success, rollback on error.
	// Use cttx.Q() to execute queries within the transaction scope.

	post := &Post{
		Title:  input.Title,
		Body:   input.Body,
		Status: "draft",
	}

	_, err := cttx.Q().Into("posts").Records(post).Exec(cttx)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return httpserver.NewJsonResponse(post), nil
}

// snippet-end: bind tx handler

// snippet-start: bind tx no input
func (h *PostHandler) HandleReadPost(cttx sqlc.Tx) (httpserver.Response, error) {
	// BindTxN is used when no request body input is needed.
	// The transaction is still available for database operations.

	var posts []Post
	err := cttx.Q().From("posts").Where(sqlc.Col("status").Eq("draft")).Select(cttx, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}

	return httpserver.NewJsonResponse(posts), nil
}

// snippet-end: bind tx no input

//go:embed config.dist.yml
var config []byte

// snippet-start: main
func main() {
	application.New(
		application.WithConfigBytes(config, "yml"),
		application.WithLoggerHandlersFromConfig,
		application.WithModuleFactory("http", httpserver.NewServer(
			"default",
			// snippet-start: with tx register
			func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
				router.HandleWith(sqlh.WithTx(NewPostHandler(), func(router *httpserver.Router, handler *PostHandler) {
					router.POST("/v1/authors/:id/posts", sqlh.BindTx(handler.HandleCreatePost))
					router.GET("/v1/posts/:id", sqlh.BindTxN(handler.HandleReadPost))
				}))

				return nil
			},
			// snippet-end: with tx register
		)),
	).Run()
}

// snippet-end: main
