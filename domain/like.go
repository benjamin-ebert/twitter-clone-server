package domain

import (
	"gorm.io/gorm"
	"time"
)

type Like struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	TweetID int `json:"tweet_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	gorm.DeletedAt `json:"deleted_at"`
}

type LikeService interface {
	Create(like *Like) error
	Delete(like *Like) error
}