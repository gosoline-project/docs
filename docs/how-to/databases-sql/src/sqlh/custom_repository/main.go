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

type AuthorCreateInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type AuthorUpdateInput struct {
	Name string `json:"name" binding:"required"`
}

// snippet-end: entities

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

// snippet-start: repository wrapper
type ReportingAuthorRepository struct {
	delegate sqlr.Repository[int64, Author]
}

func NewReportingAuthorRepository(ctx context.Context, config cfg.Config, logger log.Logger, name string) (sqlr.Repository[int64, Author], error) {
	repo, err := sqlr.NewRepository[int64, Author](ctx, config, logger, name)
	if err != nil {
		return nil, err
	}

	return &ReportingAuthorRepository{delegate: repo}, nil
}

func (r *ReportingAuthorRepository) Create(ctx context.Context, entity *Author, opts ...func(qb *sqlr.QueryBuilderCreate)) error {
	return r.delegate.Create(ctx, entity, opts...)
}

func (r *ReportingAuthorRepository) Read(ctx context.Context, id int64, opts ...func(qb *sqlr.QueryBuilderRead)) (*Author, error) {
	return r.delegate.Read(ctx, id, opts...)
}

func (r *ReportingAuthorRepository) Query(ctx context.Context, opts ...func(qb *sqlr.QueryBuilderSelect)) ([]Author, error) {
	return r.delegate.Query(ctx, append(opts, func(qb *sqlr.QueryBuilderSelect) {
		qb.OrderBy("created_at DESC")
		qb.Limit(100)
	})...)
}

func (r *ReportingAuthorRepository) Update(ctx context.Context, entity *Author, opts ...func(qb *sqlr.QueryBuilderUpdate)) (*Author, error) {
	return r.delegate.Update(ctx, entity, opts...)
}

func (r *ReportingAuthorRepository) Delete(ctx context.Context, id int64, opts ...func(qb *sqlr.QueryBuilderDelete)) error {
	return r.delegate.Delete(ctx, id, opts...)
}

func (r *ReportingAuthorRepository) Close() error {
	return r.delegate.Close()
}

// snippet-end: repository wrapper

//go:embed config.dist.yml
var config []byte

// snippet-start: custom repository
func main() {
	application.New(
		application.WithConfigBytes(config, "yml"),
		application.WithLoggerHandlersFromConfig,
		application.WithModuleFactory("http", httpserver.NewServer(
			"default",
			func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
				router.HandleWith(sqlh.WithCrudHandlers(
					1,
					"author",
					sqlh.SimpleTransformer(&AuthorTransformer{}),
					sqlh.WithClientName[int64, Author]("reporting"),
					sqlh.WithRepositoryFactory[int64, Author](NewReportingAuthorRepository),
				))

				return nil
			},
		)),
	).Run()
}

// snippet-end: custom repository
