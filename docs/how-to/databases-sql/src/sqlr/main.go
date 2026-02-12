package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/gosoline-project/sqlc"
	"github.com/gosoline-project/sqlr"
	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/kernel"
	gosolineLog "github.com/justtrackio/gosoline/pkg/log"
)

// snippet-start: entities
type Author struct {
	sqlr.Entity[int64]
	Name  string `db:"name"`
	Email string `db:"email"`
	Posts []Post `db:"-,foreignKey:author_id"`
}

type Post struct {
	sqlr.Entity[int64]
	AuthorID int64  `db:"author_id"`
	Title    string `db:"title"`
	Body     string `db:"body"`
	Status   string `db:"status"`
	Author   Author `db:"-,belongsTo:author_id"`
	Tags     []Tag  `db:"-,many2many:post_tags"`
}

type Tag struct {
	sqlr.Entity[int64]
	Name string `db:"name"`
}

// snippet-end: entities

// snippet-start: entity auto-preload
type PostWithPreloads struct {
	sqlr.Entity[int64]
	AuthorID int64  `db:"author_id"`
	Title    string `db:"title"`
	Body     string `db:"body"`
	Status   string `db:"status"`
	Author   Author `db:"-,belongsTo:author_id,preload"`
	Tags     []Tag  `db:"-,many2many:post_tags,preload"`
}

// snippet-end: entity auto-preload

//go:embed config.dist.yml
var config []byte

func main() {
	application.RunFunc(
		func(ctx context.Context, config cfg.Config, logger gosolineLog.Logger) (kernel.ModuleRunFunc, error) {
			// snippet-start: create repository
			authorRepo, err := sqlr.NewRepository[int64, Author](ctx, config, logger, "default")
			if err != nil {
				return nil, fmt.Errorf("failed to create author repository: %w", err)
			}
			// snippet-end: create repository

			postRepo, err := sqlr.NewRepository[int64, Post](ctx, config, logger, "default")
			if err != nil {
				return nil, fmt.Errorf("failed to create post repository: %w", err)
			}

			tagRepo, err := sqlr.NewRepository[int64, Tag](ctx, config, logger, "default")
			if err != nil {
				return nil, fmt.Errorf("failed to create tag repository: %w", err)
			}

			service := &BlogService{
				authorRepo: authorRepo,
				postRepo:   postRepo,
				tagRepo:    tagRepo,
			}

			return func(ctx context.Context) error {
				timestamp := time.Now().UnixNano()

				// Create entities
				alice, err := service.createAuthor(ctx, "Alice", fmt.Sprintf("alice-%d@mail.io", timestamp))
				if err != nil {
					return fmt.Errorf("failed to create author: %w", err)
				}
				logger.Info(ctx, "created author: %s (id=%d)", alice.Name, alice.Id)

				post, err := service.createPost(ctx, alice.Id, "Hello World", "My first post!", "draft")
				if err != nil {
					return fmt.Errorf("failed to create post: %w", err)
				}
				logger.Info(ctx, "created post: %s (id=%d)", post.Title, post.Id)

				tag, err := service.createTag(ctx, "golang")
				if err != nil {
					return fmt.Errorf("failed to create tag: %w", err)
				}
				logger.Info(ctx, "created tag: %s (id=%d)", tag.Name, tag.Id)

				// Read by ID
				readAuthor, err := service.readAuthor(ctx, alice.Id)
				if err != nil {
					return fmt.Errorf("failed to read author: %w", err)
				}
				logger.Info(ctx, "read author: %s", readAuthor.Name)

				// Read with joins
				readPost, err := service.readPostWithAuthor(ctx, post.Id)
				if err != nil {
					return fmt.Errorf("failed to read post with author: %w", err)
				}
				logger.Info(ctx, "read post with author: %s by %s", readPost.Title, readPost.Author.Name)

				// Update entity
				updatedPost, err := service.updatePostStatus(ctx, post, "published")
				if err != nil {
					return fmt.Errorf("failed to update post: %w", err)
				}
				logger.Info(ctx, "updated post status to: %s", updatedPost.Status)

				// Query with filtering
				posts, err := service.queryPublishedPosts(ctx)
				if err != nil {
					return fmt.Errorf("failed to query posts: %w", err)
				}
				logger.Info(ctx, "found %d published posts", len(posts))

				// Query with preloading
				postsWithAuthor, err := service.queryPostsWithAuthor(ctx)
				if err != nil {
					return fmt.Errorf("failed to query posts with author: %w", err)
				}
				logger.Info(ctx, "found %d posts with author preloaded", len(postsWithAuthor))

				// Query with joins
				postsWithJoin, err := service.queryPostsWithAuthorJoin(ctx)
				if err != nil {
					return fmt.Errorf("failed to query posts with join: %w", err)
				}
				logger.Info(ctx, "found %d posts with author joined", len(postsWithJoin))

				// Delete entity
				if err = service.deleteTag(ctx, tag.Id); err != nil {
					return fmt.Errorf("failed to delete tag: %w", err)
				}
				logger.Info(ctx, "deleted tag %d", tag.Id)

				// Error handling
				_, err = service.readAuthor(ctx, 999999)
				if errors.Is(err, sqlr.ErrNotFound) {
					logger.Info(ctx, "author 999999 not found (as expected)")
				}

				return nil
			}, nil
		},
		application.WithConfigBytes(config, "yml"),
	)
}

type BlogService struct {
	authorRepo sqlr.Repository[int64, Author]
	postRepo   sqlr.Repository[int64, Post]
	tagRepo    sqlr.Repository[int64, Tag]
}

// snippet-start: create
func (s *BlogService) createAuthor(ctx context.Context, name, email string) (*Author, error) {
	author := &Author{
		Name:  name,
		Email: email,
	}

	if err := s.authorRepo.Create(ctx, author); err != nil {
		return nil, fmt.Errorf("failed to create author: %w", err)
	}

	// author.Id is automatically set for auto-increment primary keys
	// author.CreatedAt and author.UpdatedAt are set via autoCreateTime/autoUpdateTime

	return author, nil
}

// snippet-end: create

func (s *BlogService) createPost(ctx context.Context, authorId int64, title, body, status string) (*Post, error) {
	post := &Post{
		AuthorID: authorId,
		Title:    title,
		Body:     body,
		Status:   status,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (s *BlogService) createTag(ctx context.Context, name string) (*Tag, error) {
	tag := &Tag{
		Name: name,
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

// snippet-start: read
func (s *BlogService) readAuthor(ctx context.Context, id int64) (*Author, error) {
	author, err := s.authorRepo.Read(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to read author: %w", err)
	}

	return author, nil
}

// snippet-end: read

// snippet-start: read with join
func (s *BlogService) readPostWithAuthor(ctx context.Context, postId int64) (*Post, error) {
	post, err := s.postRepo.Read(ctx, postId, func(qb *sqlr.QueryBuilderRead) {
		qb.LeftJoin("Author")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read post with author: %w", err)
	}

	return post, nil
}

// snippet-end: read with join

// snippet-start: update
func (s *BlogService) updatePostStatus(ctx context.Context, post *Post, status string) (*Post, error) {
	post.Status = status

	updated, err := s.postRepo.Update(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	// updated.UpdatedAt is automatically refreshed via autoUpdateTime

	return updated, nil
}

// snippet-end: update

// snippet-start: delete
func (s *BlogService) deleteTag(ctx context.Context, id int64) error {
	if err := s.tagRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// snippet-end: delete

// snippet-start: query
func (s *BlogService) queryPublishedPosts(ctx context.Context) ([]Post, error) {
	posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {
		qb.Where(sqlc.Col("status").Eq("published")).
			OrderBy("created_at DESC").
			Limit(10).
			Offset(0)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}

	return posts, nil
}

// snippet-end: query

// snippet-start: query preload
func (s *BlogService) queryPostsWithAuthor(ctx context.Context) ([]Post, error) {
	posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {
		qb.Where(sqlc.Col("status").Eq("published")).
			Preload("Author").
			Preload("Tags")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}

	// Each post now has post.Author and post.Tags populated

	return posts, nil
}

// snippet-end: query preload

// snippet-start: query join
func (s *BlogService) queryPostsWithAuthorJoin(ctx context.Context) ([]Post, error) {
	posts, err := s.postRepo.Query(ctx, func(qb *sqlr.QueryBuilderSelect) {
		qb.Where(sqlc.Col("status").Eq("published")).
			LeftJoin("Author")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}

	// Each post now has post.Author populated via a SQL JOIN

	return posts, nil
}

// snippet-end: query join
