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
		tv.contentMinLength,
		tv.contentMaxLength,
		tv.repliedToTweetExists,
		tv.retweetedTweetExists)
	if err != nil {
		return err
	}
	return tv.tweetGorm.CreateTweet(tweet)
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
		return errs.ContentTooShort
	}
	return nil
}

func (tv *tweetValidator) contentMaxLength(tweet *domain.Tweet) error {
	if utf8.RuneCountInString(tweet.Content) > 280 {
		return errs.ContentTooLong
	}
	return nil
}

func (tv *tweetValidator) repliedToTweetExists(tweet *domain.Tweet) error {
	if tweet.RepliesToID > 0 {
		err := tv.db.First(&domain.Tweet{ID: tweet.RepliesToID}).Error
		if err != nil {
			return errs.RepliedToTweetDoesNotExist
		}
	}
	return nil
}

func (tv *tweetValidator) retweetedTweetExists(tweet *domain.Tweet) error {
	if tweet.RetweetsID > 0 {
		err := tv.db.First(&domain.Tweet{ID: tweet.RetweetsID}).Error
		if err != nil {
			return errs.RetweetedTweetDoesNotExist
		}
	}
	return nil
}

func (tg *tweetGorm) CreateTweet(tweet *domain.Tweet) error {
	err := tg.db.Create(tweet).Error
	if err != nil {
		return err
	}
	return nil
}