package main

import (
	"context"
	"time"
)

type Auth struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	User *User `json:"user"`
	Source string `json:"source"`
	SourceID int `json:"source_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthService interface {
	FindAuthByID(ctx context.Context, id int) (*Auth, error)
	FindAuths(ctx context.Context, filter AuthFilter) ([]*Auth, int, error)
	CreateAuth(ctx context.Context, auth *Auth) error
	DeleteAuth(ctx context.Context, id int) error
}

type AuthFilter struct {
	ID *int `json:"id"`
	UserID *int `json:"user_id"`
	User *string `json:"user"`
	Source *string `json:"source"`

	Offset int `json:"offset"`
	Limit int `json:"limit"`
}
