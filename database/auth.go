package database

import (
	"context"
	"wtfTwitter/domain"
)

type AuthService struct {
	db *DB
}

func NewAuthService(db *DB) *AuthService {
	return &AuthService{
		db: db,
	}
}

func (s *AuthService) FindAuthByID(ctx context.Context, id int) (*domain.Auth, error) {
	return nil, nil
}

func (s *AuthService) CreateAuth(ctx context.Context, auth *domain.Auth) error {
	if auth.UserID == 0 {
		//var user domain.User
		//user.Email = auth
	}
	err := s.db.db.Create(auth).Error
	if err != nil {
		return err
	}
	return nil
}