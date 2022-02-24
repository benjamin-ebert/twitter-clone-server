package crud

import (
	"gorm.io/gorm"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// LikeService manages Likes.
// It implements the domain.LikeService interface.
type LikeService struct {
	likeValidator
}

// likeValidator runs validations on incoming Like data.
// On success, it passes the data on to likeGorm.
// Otherwise, it returns the error of the validation that has failed.
type likeValidator struct {
	likeGorm
}

// likeGorm runs CRUD operations on the database using incoming Like data.
// It assumes that data has been validated. On success, it returns nil.
// Otherwise, it returns the error of the operation that has failed.
type likeGorm struct {
	db *gorm.DB
}

// NewLikeService returns an instance of LikeService.
func NewLikeService(db *gorm.DB) *LikeService {
	return &LikeService{
		likeValidator{
			likeGorm{
				db: db,
			},
		},
	}
}

// Ensure the LikeService struct properly implements the domain.LikeService interface.
// If it does not, then this expression becomes invalid and won't compile.
var _ domain.LikeService = &LikeService{}

// Create runs validations needed for creating new Like database records.
func (lv *likeValidator) Create(like *domain.Like) error {
	err := runLikeValFns(like,
		lv.userIdValid,
		lv.likedTweetExists,
		lv.notAlreadyLiked)
	if err != nil {
		return err
	}
	return lv.likeGorm.Create(like)
}

// Delete runs validations needed for deleting existing Like database records.
func (lv *likeValidator) Delete(like *domain.Like) error {
	err := runLikeValFns(like, lv.likeExists)
	if err != nil {
		return err
	}
	return lv.likeGorm.Delete(like)
}

// runLikeValFns runs any number of functions of type likeValFn on the passed in Like object.
// If none of them returns an error, it returns nil. Otherwise, it returns the respective error.
func runLikeValFns(like *domain.Like, fns ...likeValFn) error {
	for _, fn := range fns {
		if err := fn(like); err != nil {
			return err
		}
	}
	return nil
}

// A likeValFn is any function that takes in a pointer to a domain.Like object and returns an error.
type likeValFn func(like *domain.Like) error

// likeExists makes sure that the Like record to be deleted actually exists.
func (lv *likeValidator) likeExists(like *domain.Like) error {
	err := lv.db.First(like, like).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.Errorf(errs.ENOTFOUND, "You cannot unlike a tweet you have not liked.")
		} else {
			return err
		}
	}
	return nil
}

// likedTweetExists makes sure that the tweet to be liked actually exists.
func (lv *likeValidator) likedTweetExists(like *domain.Like) error {
	err := lv.db.First(&domain.Tweet{ID: like.TweetID}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.Errorf(errs.ENOTFOUND, "The liked tweet does not exist.")
		} else {
			return err
		}
	}
	return nil
}

// notAlreadyLiked makes sure that the user doesn't already like the tweet.
func (lv *likeValidator) notAlreadyLiked(like *domain.Like) error {
	err := lv.db.First(like, like).Error
	if err == nil {
		return errs.Errorf(errs.EINVALID, "You already like that tweet.")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

// userIdValid ensures that the userId is not empty.
func (lv *likeValidator) userIdValid(like *domain.Like) error {
	if like.UserID <= 0 {
		return errs.UserIdValid
	}
	return nil
}

// ByUserID retrieves all likes of a user, along with the Tweet belonging to each Like.
func (lg *likeGorm) ByUserID(userId int) ([]domain.Like, error) {
	var likes []domain.Like
	err := lg.db.
		Where("user_id = ?", userId).
		Preload("Tweet.User").
		Find(&likes).Error
	if err != nil {
		return nil, err
	}
	return likes, nil
}

// AuthLikes takes a user ID and a tweet ID and returns a boolean expressing whether
// the given user likes the given tweet or not.
func (lg *likeGorm) AuthLikes(userId, tweetId int) bool {
	err := lg.db.First(&domain.Like{}, &domain.Like{UserID: userId, TweetID: tweetId}).Error
	if err == nil {
		return true
	}
	return false
}

// Create stores the data from the Like object in a new database record.
// On success, it eager-loads (preloads) the tweet relation, so that
// the json response displays the full data of the liked tweet.
func (lg *likeGorm) Create(like *domain.Like) error {
	err := lg.db.Create(like).Error
	if err != nil {
		return err
	}
	lg.db.Preload("Tweet").First(like)
	return nil
}

// Delete permanently deletes the database record matching the data from the Like object.
func (lg *likeGorm) Delete(like *domain.Like) error {
	err := lg.db.Delete(like).Error
	if err != nil {
		return err
	}
	return nil
}
