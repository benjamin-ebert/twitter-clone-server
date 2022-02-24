package domain

import (
	"gorm.io/gorm"
	"time"
)

// Like represents a many-to-many relationship between a User and a Tweet.
// A Like is created when a user decides to like a tweet. It's destroyed when
// a user decides to unlike a previously liked tweet, or when the tweet gets deleted.
type Like struct {
	ID int `json:"id"`
	UserID int `json:"user_id" gorm:"notNull;index"`
	TweetID int `json:"-"`
	Tweet Tweet `json:"tweet"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// TODO: Better hard delete?
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// LikeService is a set of methods to manipulate and work with the Like model.
type LikeService interface {
	ByUserID(userId int) ([]Like, error)
	AuthLikes(authedUserId, tweetId int) bool
	Create(like *Like) error
	Delete(like *Like) error
}