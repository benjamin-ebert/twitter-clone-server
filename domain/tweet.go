package domain

import (
	"gorm.io/gorm"
	"time"
)

// Tweet represents a tweet. It always has a one-to-many relationship with
// a User. It can also have the following relationships:
// - A self-referential many-to-many rel. with other Tweets. That happens
// if the Tweet gets retweeted or replied to, or if the Tweet itself is a reply
// or a retweet of an existing tweet. In that case the rel. is determined by
// the RepliesToID / RetweetsID, which hold the ID of the existing "target" tweet.
// If both are null, the Tweet is an "original" Tweet (neither reply nor retweet).
// - A many-to-many rel. with users, through the "pivot" of Likes.
// - A one-to-many rel. with images, since up to 4 images can be attached to a Tweet.
type Tweet struct {
	ID int `json:"id"`
	Content string `json:"content"`
	UserID int `json:"user_id"` // TODO: add non null to this and other models. (Fix gorm defs in general)

	RepliesToID int `json:"replies_to_id,omitempty" gorm:"default:null"`
	Replies []Tweet `json:"replies" gorm:"foreignKey:RepliesToID"`
	RetweetsID int `json:"retweets_id,omitempty" gorm:"default:null"`
	Retweets []Tweet `json:"retweets" gorm:"foreignKey:RetweetsID"`
	Likes []Like `json:"likes" gorm:"foreignKey:TweetID"`
	Images []Image `json:"images" gorm:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

// TweetService is a set of methods to manipulate and work with the Tweet model.
type TweetService interface {
	ByID(id int) (*Tweet, error)
	Create(tweet *Tweet) error
	Delete(tweet *Tweet) error
}