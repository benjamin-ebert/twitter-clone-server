package domain

import "time"

type Tweet struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TweetService interface {
	CreateTweet (tweet *Tweet) error
}