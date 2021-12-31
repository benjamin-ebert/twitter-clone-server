package database

import (
	"gorm.io/gorm"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

var _ domain.FollowService = &FollowService{}

func NewFollowService(db *gorm.DB) *FollowService {
	return &FollowService{
		followValidator{
			followGorm{
				db: db,
			},
		},
	}
}

type FollowService struct {
	followValidator
}

type followValidator struct {
	followGorm
}

type followGorm struct {
	db *gorm.DB
}

func (fv *followValidator) Create(follow *domain.Follow) error {
	err := runFollowValFns(follow,
		fv.followedIsNotFollower,
		fv.followDoesNotExist,
		fv.followedExists)
	if err != nil {
		return err
	}
	return fv.followGorm.Create(follow)
}

func runFollowValFns(follow *domain.Follow, fns ...followValFn) error {
	for _, fn := range fns {
		if err := fn(follow); err != nil {
			return err
		}
	}
	return nil
}

type followValFn func(follow *domain.Follow) error

func (fv *followValidator) followDoesNotExist(follow *domain.Follow) error {
	var existing domain.Follow
	query := fv.db.Where(follow)
	err := query.First(&existing).Error
	if err == nil {
		return errs.AlreadyExists
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (fv *followValidator) followedExists(follow *domain.Follow) error {
	err := fv.db.First(&domain.User{ID: follow.FollowedID}).Error
	if err != nil {
		return errs.FollowedDoesNotExist
	}
	return nil
}

func (fv *followValidator) followedIsNotFollower(follow *domain.Follow) error {
	if follow.FollowerID == follow.FollowedID {
		return errs.FollowedIsFollower
	}
	return nil
}

func (fg *followGorm) Create(follow *domain.Follow) error {
	err := fg.db.Create(follow).Error
	fg.db.Preload("Followed").Preload("Follower").First(follow)
	if err != nil {
		return err
	}
	return nil
}

func (fg *followGorm) Delete(follow *domain.Follow) error {
	err := fg.db.Where(follow).Delete(follow).Error
	if err != nil {
		return err
	}
	return nil
}