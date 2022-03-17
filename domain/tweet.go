package domain

import (
	"gorm.io/gorm"
	"time"
)

// Tweet represents a tweet. It always has a one-to-many relationship with
// the User who created it. It can also have the following relationships:
// - A many-to-many rel. with Likes (which in a sense is a "pivot" to users).
// - A self-referential many-to-many rel. with other Tweets. That happens
// if the Tweet gets retweeted or replied to, or if the Tweet itself is a reply
// or a retweet of an existing tweet. In that case the rel. is determined by
// the RepliesToID / RetweetsID, which hold the ID of the existing "parent" tweet.
// If both are null, the Tweet is an "original" Tweet (neither a reply nor a retweet).
// Originals can have both Replies and Retweets. Same goes for Replies. Retweets can
// have none. If a Retweet gets Replies or Retweets, those will reference the "parent"
// of the Retweet.
// - A kind of one-to-many rel. with images, since Originals or Replies can have up
// to four images attached to them. However, it's not a relationship in the traditional
// "database-sense", since tweet images have no representation in the database.
// They are only stored in the filesystem. Which tweet they belong to is resolved through
// the path of their location in the filesystem.
type Tweet struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id" gorm:"notNull;index"`
	User    User   `json:"user"`
	Content string `json:"content"`

	RepliesToID  *int    `json:"replies_to_id,omitempty" gorm:"default:null"`
	RepliesTo    *Tweet  `json:"replies_to,omitempty" gorm:"foreignKey:RepliesToID;references:ID"`
	Replies      []Tweet `json:"replies" gorm:"foreignKey:RepliesToID"`
	RepliesCount int     `json:"replies_count" gorm:"-"`
	AuthReplied  bool    `json:"auth_replied" gorm:"-"`

	RetweetsID    *int    `json:"retweets_id,omitempty" gorm:"default:null"`
	RetweetsTweet *Tweet  `json:"retweets_tweet,omitempty" gorm:"foreignKey:RetweetsID;references:ID"`
	Retweets      []Tweet `json:"retweets" gorm:"foreignKey:RetweetsID"`
	RetweetsCount int     `json:"retweets_count" gorm:"-"`
	AuthRetweet   *Tweet  `json:"auth_retweet,omitempty" gorm:"foreignKey:RetweetsID;references:ID"`

	Likes      []Like `json:"likes" gorm:"foreignKey:TweetID"`
	LikesCount int    `json:"likes_count" gorm:"-"`
	AuthLike   *Like  `json:"auth_like,omitempty" gorm:"foreignKey:TweetID;references:ID"`

	Images []Image `json:"images" gorm:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TweetService is a set of methods to manipulate and work with the Tweet model.
type TweetService interface {
	ByID(id int) (*Tweet, error)
	ByUserID(userId, offset int) ([]Tweet, error)

	GetFeed(offset int) ([]Tweet, error)
	OriginalsByUserID(userId, offset int) ([]Tweet, error)
	ImageTweetsByUserID(userId, offset int) ([]Tweet, error)
	LikedTweetsByUserID(userId, offset int) ([]Tweet, error)

	CountReplies(id int) (int, error)
	CountRetweets(id int) (int, error)
	CountLikes(id int) (int, error)

	GetAuthRepliedBool(authedUserId, tweetId int) (bool, error)
	GetAuthRetweet(authUserId, tweetId int) (*Tweet, error)
	GetAuthLike(authUserId, tweetId int) (*Like, error)

	Create(tweet *Tweet) error
	Delete(tweet *Tweet) error
}
