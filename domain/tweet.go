package domain

import (
	"gorm.io/gorm"
	"time"
)

type Tweet struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	Content string `json:"content"`

	// If is a pointer, so the default value will be nil too.
	//RepliesToID *int `json:"replies_to_id"`
	//RetweetsID *int `json:"retweets_id"`
	RepliesToID int `json:"replies_to_id,omitempty" gorm:"default:null"`
	RetweetsID int `json:"retweets_id,omitempty" gorm:"default:null"`
	Replies []Tweet `json:"replies" gorm:"foreignKey:RepliesToID"`
	Retweets []Tweet `json:"retweets" gorm:"foreignKey:RetweetsID"`
	Likes []Like `json:"likes" gorm:"foreignKey:TweetID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	gorm.DeletedAt `json:"deleted_at"`
}

type TweetService interface {
	CreateTweet (tweet *Tweet) error
	DeleteTweet (tweet *Tweet) error
}