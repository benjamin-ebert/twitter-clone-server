package domain

import (
	"time"
)

// Like represents a many-to-many relationship between a User and a Tweet.
// A Like is created when a user decides to like a tweet, and destroyed when
// a user decides to unlike a previously liked tweet. In that case the Like
// will be hard-deleted. A Like will not be deleted however, if its associated
// Tweet gets deleted. That is to still have the Tweet's Likes, should the Tweet
// need to be restored, since tweets are only being soft-deleted. This solution
// is not meant to be final, since it would fill up the database with obsolete
// Likes of dead Tweets over time. An automated hard-delete of soft-deleted Tweets
// and their associated Likes, running after a certain period past the Tweet's
// initial soft-deletion would be a possible solution.
type Like struct {
	ID int `json:"id"`
	UserID int `json:"user_id" gorm:"notNull;index"`
	TweetID int `json:"-"`
	Tweet Tweet `json:"tweet"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LikeService is a set of methods to manipulate and work with the Like model.
type LikeService interface {
	Create(like *Like) error
	Delete(like *Like) error
}