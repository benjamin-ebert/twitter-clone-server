package database

import (
	"gorm.io/gorm"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
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
	err := runLikeValFuncs(like,
		lv.likedTweetExists,
		lv.tweetNotYetLiked)
	if err != nil {
		return err
	}
	return lv.likeGorm.Create(like)
}

func (lv *likeValidator) Delete(like *domain.Like) error {
	err := runLikeValFuncs(like, lv.likeExists)
	if err != nil {
		return err
	}
	return lv.likeGorm.Delete(like)
}

type likeValFunc func(like *domain.Like) error
func runLikeValFuncs(like *domain.Like, fns ...likeValFunc) error {
	for _, fn := range fns {
		if err := fn(like); err != nil {
			return err
		}
	}
	return nil
}

func (lv *likeValidator) likeExists(like *domain.Like) error {
	err := lv.db.First(like, "user_id = ? AND tweet_id = ?", like.UserID, like.TweetID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.Errorf(errs.ENOTFOUND, "The like does not exist.")
		} else {
			return err
		}
	}
	return nil
}

func (lv *likeValidator) likedTweetExists(like *domain.Like) error {
	var tweet domain.Tweet
	err := lv.db.First(&tweet, "id = ?", like.TweetID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errs.Errorf(errs.ENOTFOUND, "The liked tweet does not exist.")
		} else {
			return err
		}
	}
	return nil
}

func (lv *likeValidator) tweetNotYetLiked(like *domain.Like) error {
	var existing domain.Like
	err := lv.db.First(&existing, "user_id = ? AND tweet_id = ?", like.UserID, like.TweetID).Error
	if err == nil {
		return errs.Errorf(errs.EINVALID, "You already like that tweet.")
	}
	return nil
}

func (lg *likeGorm) Create(like *domain.Like) error {
	err := lg.db.Create(like).Error
	// Eager load the like's tweet so that it properly shows in the json response.
	lg.db.Preload("Tweet").First(like)
	if err != nil {
		return err
	}
	return nil
}

func (lg *likeGorm) Delete(like *domain.Like) error {
	err := lg.db.Delete(like, "user_id = ? AND tweet_id = ?", like.UserID, like.TweetID).Error
	if err != nil {
		return err
	}
	return nil
}
