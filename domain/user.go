package domain

import (
	"context"
	"time"
)

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	PasswordHash string `json:"password_hash"`
	Remember string `json:"remember"`
	RememberHash string `json:"remember_hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	OAuths []*OAuth `json:"oauths"`
}

type UserService interface {
	//FindUserByID(ctx context.Context, id int) (*User, error)
	FindUserByEmail(email string) (*User, error)
	//FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	//DeleteUser(ctx context.Context, id int) error
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
