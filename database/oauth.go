package database

import (
	"context"
	"gorm.io/gorm"
	"wtfTwitter/domain"
)

type OAuthService struct {
	db *gorm.DB
}

func NewOAuthService(db *gorm.DB) *OAuthService {
	return &OAuthService{
		db: db,
	}
}

func (s *OAuthService) FindOAuthByID(ctx context.Context, id int) (*domain.OAuth, error) {
	return nil, nil
}

func (s *OAuthService) CreateOAuth(ctx context.Context, oauth *domain.OAuth) error {
	if oauth.UserID == 0 {
		//var user domain.User
		//user.Email = oauth
	}
	err := s.db.Create(oauth).Error
	if err != nil {
		return err
	}
	return nil
}