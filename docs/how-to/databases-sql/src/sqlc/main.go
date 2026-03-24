package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/gosoline-project/sqlc"
	"github.com/justtrackio/gosoline/pkg/application"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/exec"
	"github.com/justtrackio/gosoline/pkg/kernel"
	gosolineLog "github.com/justtrackio/gosoline/pkg/log"
)

// snippet-start: models
type Author struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// snippet-end: models

type Post struct {
	Id        int64     `db:"id"`
	AuthorId  int64     `db:"author_id"`
	Title     string    `db:"title"`
	Body      string    `db:"body"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Comment struct {
	Id        int64     `db:"id"`
	AuthorId  int64     `db:"author_id"`
	PostId    int64     `db:"post_id"`
	Body      string    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Tag struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PostWithAuthor struct {
	Post
	AuthorName  string `db:"author_name"`
	AuthorEmail string `db:"author_email"`
}

type AuditLog struct {
	EntityId   int64  `db:"entity_id"`
	ActorEmail string `db:"actor_email"`
	Action     string `db:"action"`
}

//go:embed config.dist.yml
var config []byte

func main() {
	application.RunFunc(
		func(ctx context.Context, config cfg.Config, logger gosolineLog.Logger) (kernel.ModuleRunFunc, error) {
			// snippet-start: main
			var err error
			var client sqlc.Client

			if client, err = sqlc.NewClient(ctx, config, logger, "default"); err != nil {
				return nil, fmt.Errorf("failed to create sqlc client: %w", err)
			}
			// snippet-end: main

			service := &BlogService{
				client: client,
			}

			return func(ctx context.Context) error {
				timestamp := time.Now().UnixNano()

				alice, err := service.createAuthor(ctx, "Alice", fmt.Sprintf("alice-%d@mail.io", timestamp))
				if err != nil {
					return fmt.Errorf("failed to create author: %w", err)
				}
				logger.Info(ctx, "created author: %s (id=%d, email=%s)", alice.Name, alice.Id, alice.Email)

				tags, err := service.createTags(ctx, []string{"golang", "sql", "tutorial"})
				if err != nil {
					return fmt.Errorf("failed to create tags: %w", err)
				}
				logger.Info(ctx, "created %d tags", len(tags))

				bob, bobPost, err := service.createAuthorWithPost(ctx, "Bob", fmt.Sprintf("bob-%d@mail.io", timestamp), "Hello World", "My first post!")
				if err != nil {
					return fmt.Errorf("failed to create author with post: %w", err)
				}
				logger.Info(ctx, "created author with post: %s (author_id=%d, post_id=%d)", bob.Name, bob.Id, bobPost.Id)

				bobPosts, err := service.queryPostsByAuthor(ctx, bob.Id)
				if err != nil {
					return fmt.Errorf("failed to query posts by author: %w", err)
				}
				logger.Info(ctx, "found %d posts by author %s", len(bobPosts), bob.Name)

				updatedPost, err := service.updatePostStatus(ctx, bobPost.Id, "published")
				if err != nil {
					return fmt.Errorf("failed to update post status: %w", err)
				}
				logger.Info(ctx, "updated post %d status to: %s", updatedPost.Id, updatedPost.Status)

				publishedPosts, err := service.queryPostsWithJoins(ctx)
				if err != nil {
					return fmt.Errorf("failed to query posts with joins: %w", err)
				}
				logger.Info(ctx, "found %d published posts with author info", len(publishedPosts))

				comment, err := service.createComment(ctx, alice.Id, bobPost.Id, "Great post!")
				if err != nil {
					return fmt.Errorf("failed to create comment: %w", err)
				}
				logger.Info(ctx, "created comment %d on post %d", comment.Id, bobPost.Id)

				if err = service.deleteComment(ctx, comment.Id); err != nil {
					return fmt.Errorf("failed to delete comment: %w", err)
				}
				logger.Info(ctx, "deleted comment %d", comment.Id)

				return nil
			}, nil
		},
		application.WithConfigBytes(config, "yml"),
	)
}

// snippet-start: service
type BlogService struct {
	client sqlc.Client
}

// snippet-end: service

func configureSession(ctx context.Context, config cfg.Config, logger gosolineLog.Logger) error {
	// snippet-start: db handle
	db, err := sqlc.NewDB(ctx, config, logger, "default")
	if err != nil {
		return fmt.Errorf("failed to create db handle: %w", err)
	}
	defer db.Close()

	if _, err = db.SQLDB().ExecContext(ctx, "SET SESSION sql_mode = 'STRICT_ALL_TABLES'"); err != nil {
		return fmt.Errorf("failed to configure sql session: %w", err)
	}
	// snippet-end: db handle

	return nil
}

func newClientFromExistingDB(existing *sql.DB, logger gosolineLog.Logger) sqlc.Client {
	// snippet-start: wrap db
	db := sqlc.WrapDB(existing, "mysql")
	client := sqlc.NewClientWithDB(logger, db, exec.NewDefaultExecutor(), sqlc.DefaultConfig())
	// snippet-end: wrap db

	return client
}

// snippet-start: create author
func (s *BlogService) createAuthor(ctx context.Context, name, email string) (*Author, error) {
	author := &Author{
		Name:  name,
		Email: email,
	}

	result, err := s.client.Q().Into("authors").Records(author).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to insert author: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	author.Id = id

	return author, nil
}

// snippet-end: create author

// snippet-start: create tags
func (s *BlogService) createTags(ctx context.Context, names []string) ([]Tag, error) {
	tags := make([]Tag, len(names))
	for i, name := range names {
		tags[i] = Tag{Name: name}
	}

	_, err := s.client.Q().Into("tags").Records(tags).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to insert tags: %w", err)
	}

	return tags, nil
}

// snippet-end: create tags

// snippet-start: query posts
func (s *BlogService) queryPostsByAuthor(ctx context.Context, authorId int64) ([]Post, error) {
	var posts []Post

	err := sqlc.From("posts").
		WithClient(s.client).
		Where(sqlc.Col("author_id").Eq(authorId)).
		OrderBy("created_at DESC").
		Select(ctx, &posts)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}

	return posts, nil
}

// snippet-end: query posts

// snippet-start: query joins
func (s *BlogService) queryPostsWithJoins(ctx context.Context) ([]PostWithAuthor, error) {
	var results []PostWithAuthor

	err := sqlc.From("posts").As("p").
		Columns("p.id", "p.author_id", "p.title", "p.body", "p.status", "p.created_at", "p.updated_at").
		LeftJoin("authors").As("a").On("p.author_id = a.id").
		Column(sqlc.Col("a.name").As("author_name")).
		Column(sqlc.Col("a.email").As("author_email")).
		Where(sqlc.Col("p.status").Eq("published")).
		OrderBy("p.created_at DESC").
		WithClient(s.client).
		Select(ctx, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts with joins: %w", err)
	}

	return results, nil
}

// snippet-end: query joins

func (s *BlogService) streamPublishedPosts(ctx context.Context) ([]Post, error) {
	rows, err := s.client.Query(ctx, "SELECT id, author_id, title, body, status, created_at, updated_at FROM posts WHERE status = ?", "published")
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	// snippet-start: query rows
	posts := make([]Post, 0)
	for rows.Next() {
		var post Post
		if err := rows.StructScan(&post); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}
	// snippet-end: query rows

	return posts, nil
}

// snippet-start: update
func (s *BlogService) updatePostStatus(ctx context.Context, postId int64, status string) (*Post, error) {
	result, err := sqlc.Update("posts").
		WithClient(s.client).
		Set("status", status).
		Where(sqlc.Col("id").Eq(postId)).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("no post found with id %d", postId)
	}

	var post Post
	err = sqlc.From("posts").
		WithClient(s.client).
		Where(sqlc.Col("id").Eq(postId)).
		Get(ctx, &post)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated post: %w", err)
	}

	return &post, nil
}

// snippet-end: update

func (s *BlogService) createAuditLog(ctx context.Context, author *Author) error {
	entry := AuditLog{
		EntityId:   author.Id,
		ActorEmail: author.Email,
		Action:     "author.created",
	}

	// snippet-start: named exec
	_, err := s.client.NamedExec(ctx,
		"INSERT INTO audit_logs (entity_id, actor_email, action) VALUES (:entity_id, :actor_email, :action)",
		entry,
	)
	// snippet-end: named exec
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	return nil
}

func (s *BlogService) publishPostsPrepared(ctx context.Context, postIds []int64) error {
	stmt, err := s.client.Prepare(ctx, "UPDATE posts SET status = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// snippet-start: prepared statements
	return s.client.WithTx(ctx, func(tx sqlc.Tx) error {
		txStmt := stmt.WithTx(ctx, tx.SQLTx())

		for _, postId := range postIds {
			if _, err := txStmt.ExecContext(ctx, "published", postId); err != nil {
				return fmt.Errorf("failed to publish post %d: %w", postId, err)
			}
		}

		return nil
	})
	// snippet-end: prepared statements
}

// snippet-start: transaction
func (s *BlogService) createAuthorWithPost(ctx context.Context, authorName, authorEmail, postTitle, postBody string) (*Author, *Post, error) {
	var author *Author
	var post *Post

	err := s.client.WithTx(ctx, func(tx sqlc.Tx) error {
		author = &Author{
			Name:  authorName,
			Email: authorEmail,
		}

		result, err := tx.Q().Into("authors").Records(author).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert author: %w", err)
		}

		authorId, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		author.Id = authorId

		post = &Post{
			AuthorId: authorId,
			Title:    postTitle,
			Body:     postBody,
			Status:   "draft",
		}

		result, err = tx.Q().Into("posts").Records(post).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert post: %w", err)
		}

		postId, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}
		post.Id = postId

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return author, post, nil
}

// snippet-end: transaction

// snippet-start: delete
func (s *BlogService) deleteComment(ctx context.Context, commentId int64) error {
	result, err := sqlc.Delete("comments").
		WithClient(s.client).
		Where(sqlc.Col("id").Eq(commentId)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no comment found with id %d", commentId)
	}

	return nil
}

// snippet-end: delete

// snippet-start: create comment
func (s *BlogService) createComment(ctx context.Context, authorId, postId int64, body string) (*Comment, error) {
	comment := &Comment{
		AuthorId: authorId,
		PostId:   postId,
		Body:     body,
	}

	result, err := s.client.Q().Into("comments").Records(comment).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to insert comment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	comment.Id = id

	return comment, nil
}

// snippet-end: create comment
