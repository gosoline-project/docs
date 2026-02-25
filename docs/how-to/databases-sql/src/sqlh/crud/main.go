package main

import (
	"context"
	_ "embed"

	"github.com/gosoline-project/httpserver"
	"github.com/gosoline-project/sqlh"
	"github.com/gosoline-project/sqlr"
	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

// snippet-start: entities
type Author struct {
	sqlr.Entity[int64]
	Name  string `db:"name"`
	Email string `db:"email"`
}

// snippet-end: entities

// snippet-start: input output
type AuthorCreateInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type AuthorUpdateInput struct {
	Name string `json:"name" binding:"required"`
}

// snippet-end: input output

// snippet-start: transformer
type AuthorTransformer struct{}

func (t *AuthorTransformer) TransformCreateInput(_ context.Context, input *AuthorCreateInput) (*Author, error) {
	return &Author{
		Name:  input.Name,
		Email: input.Email,
	}, nil
}

func (t *AuthorTransformer) TransformUpdateInput(_ context.Context, entity *Author, input *AuthorUpdateInput) (*Author, error) {
	entity.Name = input.Name

	return entity, nil
}

func (t *AuthorTransformer) RenderEntityResponse(_ context.Context, entity *Author) (httpserver.Response, error) {
	return httpserver.NewJsonResponse(entity), nil
}

func (t *AuthorTransformer) RenderQueryResponse(_ context.Context, entities []Author) (httpserver.Response, error) {
	return httpserver.NewJsonResponse(entities), nil
}

// snippet-end: transformer

//go:embed config.dist.yml
var config []byte

// snippet-start: main
func main() {
	application.New(
		application.WithConfigBytes(config, "yml"),
		application.WithLoggerHandlersFromConfig,
		application.WithModuleFactory("http", httpserver.NewServer(
			"default",
			// snippet-start: crud handlers
			func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
				router.HandleWith(sqlh.WithCrudHandlers(1, "author", sqlh.SimpleTransformer(&AuthorTransformer{})))

				return nil
			},
			// snippet-end: crud handlers
		)),
	).Run()
}

// snippet-end: main
