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
	UserID int `json:"user_id" gorm:"notNull;index"`
	User User `json:"user"`
	Content string `json:"content"`

	RepliesToID int `json:"replies_to_id,omitempty" gorm:"default:null"`
	Replies []Tweet `json:"replies" gorm:"foreignKey:RepliesToID"`
	RepliesCount int `json:"replies_count" gorm:"-"`
	AuthReplied bool `json:"auth_replied" gorm:"-"`
	RetweetsID int `json:"retweets_id,omitempty" gorm:"default:null"`
	Retweets []Tweet `json:"retweets" gorm:"foreignKey:RetweetsID"`
	RetweetsCount int `json:"retweets_count" gorm:"-"`
	Likes []Like `json:"likes" gorm:"foreignKey:TweetID"`
	LikesCount int `json:"likes_count" gorm:"-"`
	AuthLikes bool `json:"auth_likes" gorm:"-"`
	Images []Image `json:"images" gorm:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TweetService is a set of methods to manipulate and work with the Tweet model.
type TweetService interface {
	ByID(id int) (*Tweet, error)
	ByUserID(userId int) ([]Tweet, error)
	OriginalsByUserID(userId int) ([]Tweet, error)
	ImageTweetsByUserID(userId int) ([]Tweet, error)
	LikedTweetsByUserID(userId int) ([]Tweet, error)

	// TODO: Put this into user.go?
	CountByUserID(userId int) (int, error)
	CountReplies(id int) (int, error)
	CountRetweets(id int) (int, error)
	CountLikes(id int) (int, error)

	// TODO: Put this somewhere else?
	CheckAuthReplied(authedUserId, tweetId int) bool

	Create(tweet *Tweet) error
	Delete(tweet *Tweet) error
}