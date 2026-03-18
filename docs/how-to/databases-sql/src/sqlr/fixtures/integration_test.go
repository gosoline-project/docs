//go:build integration && fixtures

package sqlrfixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gosoline-project/sqlr"
	"github.com/justtrackio/gosoline/pkg/fixtures"
	"github.com/justtrackio/gosoline/pkg/test/suite"
)

// snippet-start: fixture entities
type Author struct {
	sqlr.Entity[int64]
	Name string `db:"name"`
}

type Post struct {
	sqlr.Entity[int64]
	AuthorID int64  `db:"author_id"`
	Title    string `db:"title"`
	Author   Author `db:"-,belongsTo:author_id"`
}

// snippet-end: fixture entities

// snippet-start: named fixtures
var authors = fixtures.NamedFixtures[Author]{
	fixtures.NewNamedFixture("author_1", Author{
		Entity: sqlr.FixtureEntity[int64](1, "2024-01-01T09:00:00Z", "2024-01-01T09:00:00Z"),
		Name:   "Alice Johnson",
	}),
}

var posts = fixtures.NamedFixtures[Post]{
	fixtures.NewNamedFixture("post_with_author", Post{
		Entity:   sqlr.FixtureEntity[int64](10, "2024-01-05T10:00:00Z", "2024-01-05T10:00:00Z"),
		AuthorID: 1,
		Title:    "Seeded with sqlr fixtures",
	}),
}

// snippet-end: named fixtures

// snippet-start: fixture set factory
func Fixtures() fixtures.FixtureSetsFactory {
	return fixtures.NewFixtureSetsFactory(
		sqlr.FixtureSetFactory[int64, Author](authors),
		sqlr.FixtureSetFactory[int64, Post](posts),
	)
}

// snippet-end: fixture set factory

type FixtureIntegrationTestSuite struct {
	suite.Suite

	ctx  context.Context
	repo sqlr.Repository[int64, Post]
}

func TestFixtureIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(FixtureIntegrationTestSuite))
}

// snippet-start: integration suite
func (s *FixtureIntegrationTestSuite) SetupSuite() []suite.Option {
	return []suite.Option{
		suite.WithConfigFile("config.dist.yml"),
		suite.WithFixtureSetFactory(Fixtures()),
	}
}

func (s *FixtureIntegrationTestSuite) SetupTest() error {
	s.ctx = s.Env().Context()

	repo, err := sqlr.NewRepository[int64, Post](s.ctx, s.Env().Config(), s.Env().Logger(), "default")
	if err != nil {
		return fmt.Errorf("failed to create post repository: %w", err)
	}

	s.repo = repo

	return nil
}

// snippet-end: integration suite

// snippet-start: fixture assertion
func (s *FixtureIntegrationTestSuite) TestReadSeededPost() {
	post, err := s.repo.Read(s.ctx, 10, func(qb *sqlr.QueryBuilderRead) {
		qb.Preload("Author")
	})
	s.Require().NoError(err)

	s.Equal(&Post{
		Entity: sqlr.Entity[int64]{
			Id:        10,
			CreatedAt: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC),
		},
		AuthorID: 1,
		Title:    "Seeded with sqlr fixtures",
		Author: Author{
			Entity: sqlr.Entity[int64]{
				Id:        1,
				CreatedAt: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			Name: "Alice Johnson",
		},
	}, post)
}

// snippet-end: fixture assertion
