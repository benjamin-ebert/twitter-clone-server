package crud

import "gorm.io/gorm"

// A ServicesConfig is any function that takes in a pointer to a Services
// object and returns an error. It's basically just wrapping the constructor
// method of any given crud service. It exists to be able to easily create
// the crud services using functional options in main.go.
type ServicesConfig func(*Services) error

// Services is a container object holding pointers to all the crud services.
// The crud services all share the database connection provided by Services.
type Services struct {
	db *gorm.DB
	User *UserService
	Tweet *TweetService
	Follow *FollowService
	Like *LikeService
	Image *ImageService
	OAuth *OAuthService
}

// NewServices returns a new Services object, containing any crud services
// it's told to create by one of the passed in ServicesConfig functions.
// It shares the passed in database connection with any crud service it creates.
func NewServices(db *gorm.DB, cfgs ...ServicesConfig) (*Services, error) {
	s := Services{
		db: db,
	}
	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

// WithUser wraps the constructor of UserService, NewUserService.
func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db, pepper, hmacKey)
		return nil
	}
}

// WithOAuth wraps the constructor of OAuthService, NewOAuthService.
func WithOAuth() ServicesConfig {
	return func(s *Services) error {
		s.OAuth = NewOAuthService(s.db)
		return nil
	}
}

// WithTweet wraps the constructor of TweetService, NewTweetService.
func WithTweet() ServicesConfig {
	return func(s *Services) error {
		s.Tweet = NewTweetService(s.db)
		return nil
	}
}

// WithFollow wraps the constructor of FollowService, NewFollowService.
func WithFollow() ServicesConfig {
	return func(s *Services) error {
		s.Follow = NewFollowService(s.db)
		return nil
	}
}

// WithLike wraps the constructor of LikeService, NewLikeService.
func WithLike() ServicesConfig {
	return func(s *Services) error {
		s.Like = NewLikeService(s.db)
		return nil
	}
}

// WithImage wraps the constructor of ImageService, NewImageService.
func WithImage() ServicesConfig {
	return func(s *Services) error {
		s.Image = NewImageService()
		return nil
	}
}
