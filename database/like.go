package database

import (
	"gorm.io/gorm"
	"wtfTwitter/domain"
)

var _ domain.LikeService = &LikeService{}

func NewLikeService(db *gorm.DB) *LikeService {
	return &LikeService{
		likeValidator{
			likeGorm{
				db: db,
			},
		},
	}
}

type LikeService struct {
	likeValidator
}

type likeValidator struct {
	likeGorm
}

type likeGorm struct {
	db *gorm.DB
}

func (lv *likeValidator) Create(like *domain.Like) error {
	return lv.likeGorm.Create(like)
}

func (lg *likeGorm) Create(like *domain.Like) error {
	err := lg.db.Create(like).Error
	if err != nil {
		return err
	}
	return nil
}

func (lg *likeGorm) Delete(like *domain.Like) error {
	err := lg.db.Where(like).Delete(like).Error
	if err != nil {
		return err
	}
	return nil
}
