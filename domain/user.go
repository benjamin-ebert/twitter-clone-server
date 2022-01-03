package domain

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Avatar string `json:"avatar"`
	Header string `json:"header"`

	Password string `json:"password" gorm:"-"`
	PasswordHash string `json:"password_hash"`
	Remember string `json:"remember" gorm:"-"`
	RememberHash string `json:"remember_hash"`

	Tweets []Tweet `json:"tweets" gorm:"foreignKey:UserID"`
	Likes []Like `json:"likes" gorm:"foreignKey:UserID"`
	Followers []Follow `json:"followers" gorm:"foreignKey:FollowedID"`
	Followeds []Follow `json:"follows" gorm:"foreignKey:FollowerID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type UserService interface {
	ByID(id int) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}

type UserFilter struct {
	ID *int `json:"id"`
	Name *string `json:"name"`
	Email *string `json:"email"`

	Offset int `json:"offset"`
	Limit int `json:"limit"`
}

type UserUpdate struct {
	Name *string `json:"name"`
	Email *string `json:"email"`
}
