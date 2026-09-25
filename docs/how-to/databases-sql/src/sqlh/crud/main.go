package main

import (
	"context"
	"embed"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/gosoline-project/httpserver"
	"github.com/gosoline-project/sqlh"
	"github.com/gosoline-project/sqlr"
	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
	"github.com/justtrackio/gosoline/pkg/validation"
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
	Email string `json:"email" binding:"required,email"`
}

type AuthorUpdateInput struct {
	sqlh.InputById[int64]
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
type AuthorMapper struct{}

func (t *AuthorMapper) TransformCreateInput(_ context.Context, input *AuthorCreateInput) (*Author, error) {
	return &Author{
		Name:  input.Name,
		Email: input.Email,
	}, nil
}

func (t *AuthorMapper) TransformUpdateInput(_ context.Context, entity *Author, input *AuthorUpdateInput) (*Author, error) {
	if err := binding.Validator.ValidateStruct(input); err != nil {
		return nil, validation.NewError(err)
	}

	entity.Name = input.Name

	return entity, nil
}

func (t *AuthorMapper) TransformPatchInputFromEntity(_ context.Context, entity *Author) (*AuthorUpdateInput, error) {
	return &AuthorUpdateInput{
		InputById: sqlh.InputById[int64]{Id: entity.Id},
		Name:      entity.Name,
	}, nil
}

func (t *AuthorMapper) TransformOutput(_ context.Context, entity *Author) (AuthorOutput, error) {
	return AuthorOutput{
		Id:        entity.Id,
		Name:      entity.Name,
		Email:     entity.Email,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}, nil
}

// snippet-end: transformer

// snippet-start: crud definition
func NewAuthorCrud() httpserver.RegisterFactoryFunc {
	transformer := &AuthorMapper{}
	definition := sqlh.NewCrudDefinition(
		transformer.TransformCreateInput,
		transformer.TransformUpdateInput,
		transformer.TransformPatchInputFromEntity,
		transformer.TransformOutput,
	)

	return sqlh.WithCrudHandlers(1, "author", sqlh.SimpleCrudDefinition(definition))
}

// snippet-end: crud definition
// snippet-start: manual handler
type AuthorCrudHandler = sqlh.CrudHandler[int64, Author, int64, AuthorCreateInput, AuthorUpdateInput, sqlh.ListInput, AuthorOutput]

func NewAuthorHandler() httpserver.HandlerFactory[AuthorCrudHandler] {
	transformer := &AuthorMapper{}
	definition := sqlh.NewCrudDefinition(
		transformer.TransformCreateInput,
		transformer.TransformUpdateInput,
		transformer.TransformPatchInputFromEntity,
		transformer.TransformOutput,
	)
	definition.DeleteOutput = definition.Output

	return sqlh.NewCrudHandler(sqlh.SimpleCrudDefinition(definition))
}

// snippet-end: manual handler

//go:embed config.dist.yml
var config embed.FS

// snippet-start: main
func main() {
	application.New(
		application.WithConfigBytes(mustReadConfig(), "yml"),
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

func mustReadConfig() []byte {
	data, err := config.ReadFile("config.dist.yml")
	if err != nil {
		panic(err)
	}

	return data
}
