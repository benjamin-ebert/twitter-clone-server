package domain

import (
	"context"
	"time"
)

type OAuth struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	User *User `json:"user"`
	Source string `json:"source"`
	SourceID int `json:"source_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OAuthService interface {
	FindOAuthByID(ctx context.Context, id int) (*OAuth, error)
	//FindOAuths(ctx context.Context, filter OAuthFilter) ([]*OAuth, int, error)
	CreateOAuth(ctx context.Context, oauth *OAuth) error
	//DeleteOAuth(ctx context.Context, id int) error
}

type OAuthFilter struct {
	ID *int `json:"id"`
	UserID *int `json:"user_id"`
	User *string `json:"user"`
	Source *string `json:"source"`

	Offset int `json:"offset"`
	Limit int `json:"limit"`
}
