package domain

import "time"

type Follow struct {
	ID int `json:"id"`
	FollowerID int `json:"-"`
	Follower User `json:"follower"`
	FollowedID int `json:"-"`
	Followed User `json:"followed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FollowService interface {
	Create(follow *Follow) error
	Delete(follow *Follow) error
}