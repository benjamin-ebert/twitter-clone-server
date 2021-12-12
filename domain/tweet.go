package domain

import "time"

type Tweet struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	Content string `json:"content"`

	// If is a pointer, so the default value will be nil too.
	//RepliesToID *int `json:"replies_to_id"`
	//RetweetsID *int `json:"retweets_id"`
	RepliesToID int `json:"replies_to_id,omitempty" gorm:"default:null"`
	RetweetsID int `json:"retweets_id,omitempty" gorm:"default:null"`
	Replies []Tweet `json:"replies,omitempty" gorm:"foreignKey:RepliesToID"`
	Retweets []Tweet `json:"retweets,omitempty" gorm:"foreignKey:RetweetsID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TweetService interface {
	CreateTweet (tweet *Tweet) error
}