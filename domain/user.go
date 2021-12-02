package domain

import (
	"context"
	"time"
)

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Auths []*Auth `json:"auths"`
}

type UserService interface {
	FindUserByID(ctx context.Context, id int) (*User, error)
	FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, id int, upd *UserUpdate) (*User, error)
	DeleteUser(ctx context.Context, id int) error
}

type UserFilter struct {
	ID *int `json:"id"`
	Name *string `json:"name"`
	Email *string `json:"email"`

	Offset int `json:"offset"`
	Limit int `json:"limit"`
}

type UserUpdate struct {
	Name *string `json:"name"`
	Email *string `json:"email"`
}
