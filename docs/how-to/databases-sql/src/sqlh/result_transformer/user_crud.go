package main

import (
	"context"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/gosoline-project/httpserver"
	"github.com/gosoline-project/sqlh"
	"github.com/gosoline-project/sqlr"
	"github.com/justtrackio/gosoline/pkg/validation"
)

// snippet-start: types
type (
	UserCreateInput struct {
		Name string `json:"name"`
	}
	UserUpdateInput struct {
		sqlh.InputById[int]
		Name string `json:"name"`
	}
	User struct {
		sqlr.Entity[int]
		Name string
	}
	UserOutput struct {
		Id        int       `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

// snippet-end: types

// snippet-start: crud
func NewUserCrud() httpserver.RegisterFactoryFunc {
	transformer := &UserMapper{}
	definition := sqlh.NewCrudDefinition(
		transformer.TransformCreateInput,
		transformer.TransformUpdateInput,
		transformer.TransformPatchInputFromEntity,
		transformer.TransformOutput,
	)

	return sqlh.WithCrudHandlers(0, "user", sqlh.SimpleCrudDefinition(definition))
}

// snippet-end: crud

// snippet-start: transformer
type UserMapper struct{}

func (t *UserMapper) TransformCreateInput(_ context.Context, input *UserCreateInput) (*User, error) {
	return &User{
		Name: input.Name,
	}, nil
}

func (t *UserMapper) TransformUpdateInput(_ context.Context, user *User, input *UserUpdateInput) (*User, error) {
	if err := binding.Validator.ValidateStruct(input); err != nil {
		return nil, validation.NewError(err)
	}

	user.Name = input.Name

	return user, nil
}

func (t *UserMapper) TransformPatchInputFromEntity(_ context.Context, user *User) (*UserUpdateInput, error) {
	return &UserUpdateInput{
		InputById: sqlh.InputById[int]{Id: user.Id},
		Name:      user.Name,
	}, nil
}

func (t *UserMapper) TransformOutput(_ context.Context, user *User) (UserOutput, error) {
	return UserOutput{
		Id:        user.Id,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// snippet-end: transformer
