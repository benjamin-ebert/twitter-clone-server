package database

import (
	"gorm.io/gorm"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// FollowService manages Follows.
// It implements the domain.FollowService interface.
type FollowService struct {
	followValidator
}

// followValidator runs validations on incoming Follow data.
// On success, it passes the data on to followGorm.
// Otherwise, it returns the error of the validation that has failed.
type followValidator struct {
	followGorm
}

// followGorm runs CRUD operations on the database using incoming Follow data.
// It assumes that data has been validated. On success, it returns nil.
// Otherwise, it returns the error of the operation that has failed.
type followGorm struct {
	db *gorm.DB
}

// NewFollowService returns an instance of FollowService.
func NewFollowService(db *gorm.DB) *FollowService {
	return &FollowService{
		followValidator{
			followGorm{
				db: db,
			},
		},
	}
}

// Ensure the FollowService struct properly implements the domain.FollowService interface.
// If it does not, then this expression becomes invalid and won't compile.
var _ domain.FollowService = &FollowService{}

// Create runs validations needed for creating new Follow database records.
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

// Delete runs validations needed for deleting existing Follow database records.
func (fv *followValidator) Delete(follow *domain.Follow) error {
	err := runFollowValFns(follow, fv.followExists)
	if err != nil {
		return err
	}
	return fv.followGorm.Delete(follow)
}

// runFollowValFns runs any number of functions of type followValFn on the passed in Follow object.
// If none of them returns an error, it returns nil. Otherwise, it returns the respective error.
func runFollowValFns(follow *domain.Follow, fns ...followValFn) error {
	for _, fn := range fns {
		if err := fn(follow); err != nil {
			return err
		}
	}
	return nil
}

// A followValFn is any function that takes in a pointer to a domain.Follow object and returns an error.
type followValFn func(follow *domain.Follow) error

// followExists makes sure that the Follow record to be deleted actually exists.
func (fv *followValidator) followExists(follow *domain.Follow) error {
	err := fv.db.First(follow, follow).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.Errorf(errs.EINVALID, "You cannot unfollow a user you're not following.")
		} else {
			return err
		}
	}
	return nil
}

// followedIsNotFollower makes sure that the followed user and the follower user are not the same person.
func (fv *followValidator) followedIsNotFollower(follow *domain.Follow) error {
	if follow.FollowerID == follow.FollowedID {
		return errs.Errorf(errs.EINVALID, "You cannot follow yourself.")
	}
	return nil
}

// followedUserExists makes sure that the user to be followed actually exists.
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

// notAlreadyFollowed makes sure that the follower user isn't already following the followed user.
func (fv *followValidator) notAlreadyFollowed(follow *domain.Follow) error {
	err := fv.db.First(follow, follow).Error
	if err == nil {
		return errs.Errorf(errs.EINVALID, "You already follow this user.")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

// Create stores the data from the Follow object in a new database record.
// On success, it eager-loads (preloads) the follower and followed user relations,
// so that the json response displays the full user data of each.
func (fg *followGorm) Create(follow *domain.Follow) error {
	err := fg.db.Create(follow).Error
	if err != nil {
		return err
	}
	fg.db.Preload("Followed").Preload("Follower").First(follow)
	return nil
}

// Delete permanently deletes the database record matching the data from the Follow object.
func (fg *followGorm) Delete(follow *domain.Follow) error {
	err := fg.db.Delete(follow).Error
	if err != nil {
		return err
	}
	return nil
}