package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gosoline-project/httpserver"
	"github.com/justtrackio/gosoline/pkg/cfg"
	"github.com/justtrackio/gosoline/pkg/log"
)

func main() {
	httpserver.RunDefaultServer(func(ctx context.Context, config cfg.Config, logger log.Logger, router *httpserver.Router) error {
		router.Group("/api/users").HandleWith(httpserver.With(NewHandler, func(r *httpserver.Router, h *Handler) {
			r.GET("", httpserver.Bind(h.ListUsers))
			r.POST("", httpserver.Bind(h.CreateUser))
			r.GET("/:id", httpserver.Bind(h.GetUser))
			r.DELETE("/:id", httpserver.Bind(h.DeleteUser))
			r.GET("/health", httpserver.BindN(h.Health))
		}))

		return nil
	})
}

type ListUsersInput struct {
	Role   string `form:"role"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type CreateUserInput struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"omitempty,oneof=admin user guest"`
}

type UserIdInput struct {
	Id int `uri:"id" binding:"required"`
}

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Handler struct {
	users map[int]*User
	next  int
}

func NewHandler(ctx context.Context, config cfg.Config, logger log.Logger) (*Handler, error) {
	return &Handler{
		users: map[int]*User{},
		next:  1,
	}, nil
}

func (h *Handler) ListUsers(ctx context.Context, input *ListUsersInput) (httpserver.Response, error) {
	var result []*User
	for _, u := range h.users {
		if input.Role != "" && u.Role != input.Role {
			continue
		}
		result = append(result, u)
		if input.Limit > 0 && len(result) >= input.Limit {
			break
		}
	}
	return httpserver.NewJsonResponse(result), nil
}

func (h *Handler) CreateUser(ctx context.Context, input *CreateUserInput) (httpserver.Response, error) {
	user := &User{
		Id:    h.next,
		Name:  input.Name,
		Email: input.Email,
		Role:  input.Role,
	}
	h.users[user.Id] = user
	h.next++

	return httpserver.NewJsonResponse(user, httpserver.WithStatusCode(http.StatusCreated)), nil
}

func (h *Handler) GetUser(ctx context.Context, input *UserIdInput) (httpserver.Response, error) {
	user, ok := h.users[input.Id]
	if !ok {
		return httpserver.GetErrorHandler()(http.StatusNotFound, errors.New("user not found")), nil
	}
	return httpserver.NewJsonResponse(user), nil
}

func (h *Handler) DeleteUser(ctx context.Context, input *UserIdInput) (httpserver.Response, error) {
	delete(h.users, input.Id)
	return httpserver.NewStatusResponse(http.StatusNoContent), nil
}

func (h *Handler) Health(ctx context.Context) (httpserver.Response, error) {
	if len(h.users) > 10000 {
		return nil, fmt.Errorf("too many users")
	}
	return httpserver.NewJsonResponse(map[string]string{"status": "ok"}), nil
}
