package domain

import "time"

// Follow represents a many-to-may relationship between two users.
// In the database Follows are stored in the follows-table, which is essentially a pivot-table.
// A Follow is created when one user decides to follow another user.
// The FollowerID is the ID of the user that follows.
// The FollowedID is the ID of the user that is being followed.
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