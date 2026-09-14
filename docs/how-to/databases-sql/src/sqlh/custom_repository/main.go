package main

import (
	"context"
	"embed"

	"github.com/gosoline-project/httpserver"
	"github.com/gosoline-project/sqlc"
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

type AuthorCreateInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type AuthorUpdateInput struct {
	sqlh.InputById[int64]
	Name string `json:"name" binding:"required"`
}

// snippet-end: entities

type AuthorMapper struct{}

func (t *AuthorMapper) TransformCreateInput(_ context.Context, input *AuthorCreateInput) (*Author, error) {
	return &Author{Name: input.Name, Email: input.Email}, nil
}

func (t *AuthorMapper) TransformUpdateInput(_ context.Context, entity *Author, input *AuthorUpdateInput) (*Author, error) {
	entity.Name = input.Name

	return entity, nil
}

func (t *AuthorMapper) TransformPatchInputFromEntity(_ context.Context, entity *Author) (*AuthorUpdateInput, error) {
	return &AuthorUpdateInput{
		InputById: sqlh.InputById[int64]{Id: entity.Id},
		Name:      entity.Name,
	}, nil
}

func (t *AuthorMapper) TransformOutput(_ context.Context, entity *Author) (*Author, error) {
	return entity, nil
}

// snippet-start: repository wrapper
type ReportingAuthorRepository struct {
	delegate sqlr.RepositoryTx[int64, Author]
}

func NewReportingAuthorRepository(client sqlc.Client, settings sqlr.Settings) (sqlr.RepositoryTx[int64, Author], error) {
	repo, err := sqlr.NewRepositoryTxWithSettings[int64, Author](client, settings)
	if err != nil {
		return nil, err
	}

	return &ReportingAuthorRepository{delegate: repo}, nil
}

func (r *ReportingAuthorRepository) Create(tx sqlr.TTx, entity *Author, opts ...func(qb *sqlr.QueryBuilderCreate)) error {
	return r.delegate.Create(tx, entity, opts...)
}

func (r *ReportingAuthorRepository) Read(tx sqlr.TTx, id int64, opts ...func(qb *sqlr.QueryBuilderRead)) (*Author, error) {
	return r.delegate.Read(tx, id, opts...)
}

func (r *ReportingAuthorRepository) Query(tx sqlr.TTx, opts ...func(qb *sqlr.QueryBuilderSelect)) ([]Author, error) {
	opts = append(opts, func(qb *sqlr.QueryBuilderSelect) {
		qb.OrderBy("created_at DESC")
		qb.Limit(100)
	})

	return r.delegate.Query(tx, opts...)
}

func (r *ReportingAuthorRepository) Count(tx sqlr.TTx, qb *sqlr.QueryBuilderSelect) (int, error) {
	return r.delegate.Count(tx, qb)
}

func (r *ReportingAuthorRepository) Update(tx sqlr.TTx, entity *Author, opts ...func(qb *sqlr.QueryBuilderUpdate)) (*Author, error) {
	return r.delegate.Update(tx, entity, opts...)
}

func (r *ReportingAuthorRepository) Delete(tx sqlr.TTx, id int64, opts ...func(qb *sqlr.QueryBuilderDelete)) error {
	return r.delegate.Delete(tx, id, opts...)
}

func (r *ReportingAuthorRepository) Close() error {
	return r.delegate.Close()
}

// snippet-end: repository wrapper

//go:embed config.dist.yml
var config embed.FS

// snippet-start: custom repository
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
				definition := sqlh.NewCrudDefinition(
					(&AuthorMapper{}).TransformCreateInput,
					(&AuthorMapper{}).TransformUpdateInput,
					(&AuthorMapper{}).TransformPatchInputFromEntity,
					(&AuthorMapper{}).TransformOutput,
				)
				router.HandleWith(sqlh.WithCrudHandlers(
					1,
					"author",
					sqlh.SimpleCrudDefinition(definition),
					sqlh.WithClientName[int64, Author]("reporting"),
					sqlh.WithRepositoryTxFactory[int64, Author](NewReportingAuthorRepository),
				))

				return nil
			},
		)),
	).Run()
}

// snippet-end: custom repository
