package domain

import "time"

type Follow struct {
	ID int `json:"id"`
	FollowerID int `json:"follower_id"`
	FollowedID int `json:"followed_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FollowService interface {
	Create(follow *Follow) error
	Delete(follow *Follow) error
}