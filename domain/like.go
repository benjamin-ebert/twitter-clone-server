package domain

import (
	"time"
)

// Like represents a many-to-many relationship between a User and a Tweet.
// A Like is created when a user decides to like a tweet. It's destroyed when
// a user decides to unlike a previously liked tweet, or when the tweet gets deleted.
type Like struct {
	ID      int   `json:"id"`
	UserID  int   `json:"user_id" gorm:"notNull;index"`
	TweetID int   `json:"tweet_id"`
	Tweet   Tweet `json:"tweet"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LikeService is a set of methods to manipulate and work with the Like model.
type LikeService interface {
	ByID(id int) (*Like, error)
	Create(like *Like) error
	Delete(like *Like) error
}
