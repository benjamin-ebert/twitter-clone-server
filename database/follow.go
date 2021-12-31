package database

import (
	"gorm.io/gorm"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

type FollowService struct {
	followValidator
}

type followValidator struct {
	followGorm
}

type followGorm struct {
	db *gorm.DB
}

func NewFollowService(db *gorm.DB) *FollowService {
	return &FollowService{
		followValidator{
			followGorm{
				db: db,
			},
		},
	}
}

var _ domain.FollowService = &FollowService{}

func (fv *followValidator) Create(follow *domain.Follow) error {
	err := runFollowValFns(follow,
		fv.followedUserExists,
		fv.notAlreadyFollowed,
		fv.followedIsNotFollower)
	if err != nil {
		return err
	}
	return fv.followGorm.Create(follow)
}

func (fv *followValidator) Delete(follow *domain.Follow) error {
	err := runFollowValFns(follow, fv.followExists)
	if err != nil {
		return err
	}
	return fv.followGorm.Delete(follow)
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

func (fv *followValidator) followExists(follow *domain.Follow) error {
	err := fv.db.First(follow, follow).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.Errorf(errs.EINVALID, "You don't follow this user.")
		} else {
			return err
		}
	}
	return nil
}

func (fv *followValidator) followedIsNotFollower(follow *domain.Follow) error {
	if follow.FollowerID == follow.FollowedID {
		return errs.Errorf(errs.EINVALID, "You cannot follow yourself.")
	}
	return nil
}

func (fv *followValidator) followedUserExists(follow *domain.Follow) error {
	err := fv.db.First(&domain.User{ID: follow.FollowedID}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.Errorf(errs.EINVALID, "The user to be followed does not exist.")
		} else {
			return err
		}
	}
	return nil
}

func (fv *followValidator) notAlreadyFollowed(follow *domain.Follow) error {
	err := fv.db.First(follow, follow).Error
	if err == nil {
		return errs.Errorf(errs.EINVALID, "You already follow this user.")
	} else if err != gorm.ErrRecordNotFound {
		return err
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
	err := fg.db.Delete(follow).Error
	if err != nil {
		return err
	}
	return nil
}