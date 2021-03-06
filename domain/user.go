package domain

import (
	"gorm.io/gorm"
	"time"
)

// User represents a user account. It stores an email address and a password,
// so people can log in and access their content.
// It can have the following relationships:
// - A self-referential many-to-many rel. with other Users. That happens
// if the User decides to follow, or starts to get followed by another User.
// - A one-to-many rel. with a Tweet, since the User can create many Tweets.
// - A many-to-many rel. with Tweets, through the "pivot" of Likes, since the
// user can like many Tweets, which in turn can be liked by many Users.
// - A one-to-many rel. with up to two images, since the user can upload one image
// for his avatar and one image for his header-picture. The two columns hold the
// paths to the stored images on the server.
type User struct {
	ID         int     `json:"id"`
	Email      string  `json:"email" gorm:"notNull;uniqueIndex"`
	Name       string  `json:"name" gorm:"notNull"`
	Handle     string  `json:"handle" gorm:"notNull"`
	Bio        string  `json:"bio"`
	Avatar     string  `json:"avatar"`
	Header     string  `json:"header"`
	AuthFollow *Follow `json:"auth_follow,omitempty" gorm:"foreignKey:FollowedID;references:ID"`

	Password     string `json:"password" gorm:"-"`
	PasswordHash string `json:"password_hash"`
	Remember     string `json:"remember" gorm:"-"`
	RememberHash string `json:"remember_hash" gorm:"notNull;uniqueIndex"`

	// If NoPasswordNeeded ist true on a User object, the database record
	// can be created / updated without a password or password hash.
	// It's set to true when a user signs in using oauth.
	NoPasswordNeeded bool `json:"no_password_needed" gorm:"-"`

	OAuths        []OAuth  `json:"o_auths" gorm:"foreignKey:UserID"`
	Tweets        []Tweet  `json:"tweets" gorm:"foreignKey:UserID"`
	TweetCount    int      `json:"tweet_count" gorm:"-"`
	Likes         []Like   `json:"likes" gorm:"foreignKey:UserID"`
	Followers     []Follow `json:"followers" gorm:"foreignKey:FollowedID"`
	FollowerCount int      `json:"follower_count" gorm:"-"`
	Followeds     []Follow `json:"follows" gorm:"foreignKey:FollowerID"`
	FollowedCount int      `json:"followed_count" gorm:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// UserService is a set of methods to manipulate and work with the User model.
// It also contains the bulk of the authentication system. Specifically it contains
// that part of the auth-system that needs to interact with the database (hashing
// and storing remember tokens and passwords, updating those values etc.).
// It does not contain the part of the auth-system that handles cookies, middleware
// redirects etc. - this is done by auth.go in the http package.
// Errors returned by UserService are usually errs.EINVALID or errs.ENOTFOUND and contain
// info messages for the user. Other errors typically result in an errs.EINTERNAL and are
// just displaying code 500 with no message.
type UserService interface {
	Authenticate(email, password string) (*User, error)
	MakeRememberToken() (string, error)
	ByID(id int) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	Search(searchTerm string) []User
	CountTweets(userId int) (int, error)
	CountFollowers(userId int) (int, error)
	CountFolloweds(userId int) (int, error)
	GetAuthFollow(authUserId, userId int) (*Follow, error)

	Create(user *User) error
	Update(user *User) error
}
