package domain

import "time"

// Follow represents a self-referential many-to-may relationship between two users.
// A Follow is created when one user decides to follow another user.
// The FollowerID is the ID of the user that follows, and the FollowedID is the ID of the
// user that is being followed. In the database Follows are stored within the follows-table.
type Follow struct {
	ID int `json:"id"`
	FollowerID int `json:"-" gorm:"notNull;index"`
	Follower User `json:"follower"`
	FollowedID int `json:"-" gorm:"notNull;index"`
	Followed User `json:"followed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FollowService is a set of methods to manipulate and work with the Follow model.
type FollowService interface {
	Create(follow *Follow) error
	Delete(follow *Follow) error
}