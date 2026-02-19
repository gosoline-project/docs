package main

import (
	"context"
	_ "embed"
	"time"

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

type AuthorOutput struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// snippet-end: input output

// snippet-start: transformer
type AuthorTransformer struct{}

func (t *AuthorTransformer) TransformCreate(_ context.Context, input *AuthorCreateInput) (*Author, error) {
	return &Author{
		Name:  input.Name,
		Email: input.Email,
	}, nil
}

func (t *AuthorTransformer) TransformUpdate(_ context.Context, entity *Author, input *AuthorUpdateInput) (*Author, error) {
	entity.Name = input.Name

	return entity, nil
}

func (t *AuthorTransformer) TransformOutput(_ context.Context, entity *Author) (*AuthorOutput, error) {
	return &AuthorOutput{
		Id:        entity.Id,
		Name:      entity.Name,
		Email:     entity.Email,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}, nil
}

// snippet-end: transformer

// snippet-start: crud handlers
func NewAuthorCrud() httpserver.RegisterFactoryFunc {
	return sqlh.WithCrudHandlers[int64, Author, AuthorCreateInput, AuthorUpdateInput, AuthorOutput](
		1,
		"author",
		sqlh.SimpleTransformer[int64, Author, AuthorCreateInput, AuthorUpdateInput, AuthorOutput](
			&AuthorTransformer{},
		),
	)
}

// snippet-end: crud handlers

//go:embed config.dist.yml
var config []byte

// snippet-start: main
func main() {
	application.New(
		application.WithConfigBytes(config, "yml"),
		application.WithLoggerHandlersFromConfig,
		application.WithModuleFactory("http", httpserver.NewServer(
			"default",
			func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
				router.HandleWith(NewAuthorCrud())

				return nil
			},
		)),
	).Run()
}

// snippet-end: main
