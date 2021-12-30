package database

import (
	"gorm.io/gorm"
	"unicode/utf8"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

var _ domain.TweetService = &TweetService{}

func NewTweetService(db *gorm.DB) *TweetService {
	return &TweetService{
		tweetValidator{
			tweetGorm{
				db: db,
			},
		},
	}
}

type TweetService struct {
	tweetValidator
}

type tweetValidator struct {
	tweetGorm
}

type tweetGorm struct {
	db *gorm.DB
}

func (tv *tweetValidator) CreateTweet(tweet *domain.Tweet) error {
	err := runTweetValFns(tweet,
		tv.repliedToTweetExists,
		tv.retweetedTweetExists,
		tv.contentMinLength,
		tv.contentMaxLength)
	if err != nil {
		return err
	}
	return tv.tweetGorm.CreateTweet(tweet)
}

func (tv *tweetValidator) DeleteTweet(tweet *domain.Tweet) error {
	err := runTweetValFns(tweet, tv.idValid)
	if err != nil {
		return err
	}
	return tv.tweetGorm.DeleteTweet(tweet)
}

type tweetValFn = func(tweet *domain.Tweet) error
func runTweetValFns(tweet *domain.Tweet, fns ...tweetValFn) error {
	for _, fn := range fns {
		if err := fn(tweet); err != nil {
			return err
		}
	}
	return nil
}

func (tv *tweetValidator) contentMinLength(tweet *domain.Tweet) error {
	if tweet.Content == "" {
		return errs.Errorf(errs.EINVALID, "Tweet content must not be empty.")
	}
	return nil
}

func (tv *tweetValidator) contentMaxLength(tweet *domain.Tweet) error {
	if utf8.RuneCountInString(tweet.Content) > 280 {
		return errs.Errorf(errs.EINVALID, "Tweet content max length is 280 characters.")
	}
	return nil
}

func (tv *tweetValidator) idValid(tweet *domain.Tweet) error {
	if tweet.ID <= 0 {
		return errs.Errorf(errs.EINVALID, "Tweet ID is invalid.")
	}
	return nil
}

func (tv *tweetValidator) repliedToTweetExists(tweet *domain.Tweet) error {
	if tweet.RepliesToID > 0 {
		err := tv.db.First(tweet, "id = ?", tweet.RepliesToID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errs.Errorf(errs.ENOTFOUND, "Tweet replied to does not exist.")
			} else {
				return err
			}
		}
	}
	return nil
}

func (tv *tweetValidator) retweetedTweetExists(tweet *domain.Tweet) error {
	if tweet.RetweetsID > 0 {
		err := tv.tweetGorm.db.First(tweet, "id = ?", tweet.RetweetsID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errs.Errorf(errs.ENOTFOUND, "Tweet retweeted tweet does not exist.")
			} else {
				return err
			}
		}
	}
	return nil
}

func (tg *tweetGorm) ByID(id int) (*domain.Tweet, error) {
	var tweet domain.Tweet
	err := tg.db.First(&tweet, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.Errorf(errs.ENOTFOUND, "The tweet does not exist.")
		} else {
			return nil, err
		}
	}
	return &tweet, nil
}

func (tg *tweetGorm) CreateTweet(tweet *domain.Tweet) error {
	err := tg.db.Create(tweet).Error
	if err != nil {
		return err
	}
	return nil
}

func (tg *tweetGorm) DeleteTweet(tweet *domain.Tweet) error {
	// Delete the tweet.
	err := tg.db.Delete(tweet, "id = ?", tweet.ID).Error
	if err != nil {
		return err
	}
	// Delete its direct replies (not cascading further).
	err = tg.db.Delete(tweet, "replies_to_id = ?", tweet.ID).Error
	if err != nil {
		return err
	}
	// Delete its direct retweets (not cascading further).
	err = tg.db.Delete(tweet, "retweets_id = ?", tweet.ID).Error
	if err != nil {
		return err
	}
	// Delete its likes.
	var like domain.Like
	err = tg.db.Delete(&like, "tweet_id = ?", tweet.ID).Error
	if err != nil {
		return err
	}
	return nil
}