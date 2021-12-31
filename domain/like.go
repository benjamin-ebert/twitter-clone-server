package domain

import (
	"time"
)

type Like struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	TweetID int `json:"-"`
	Tweet Tweet `json:"tweet"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LikeService interface {
	Create(like *Like) error
	Delete(like *Like) error
}