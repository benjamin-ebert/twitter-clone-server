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
	err := runTweetValFns(tweet,
		tv.idValid,
		tv.exists)
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

func (tv *tweetValidator) exists(tweet *domain.Tweet) error {
	err := tv.db.First(tweet, "id = ? AND user_id = ?", tweet.ID, tweet.UserID).Error
	if err != nil {
		return errs.NotFound
	}
	return nil
}

func (tv *tweetValidator) idValid(tweet *domain.Tweet) error {
	if tweet.ID <= 0 {
		return errs.IDInvalid
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
		var retweetedTweet domain.Tweet
		err := tv.tweetGorm.db.First(&retweetedTweet, "id = ?", tweet.RetweetsID).Error
		if err != nil {
			return errs.RetweetedTweetDoesNotExist
		}
		tweet.Content = retweetedTweet.Content
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

func (tg *tweetGorm) DeleteTweet(tweet *domain.Tweet) error {
	var reply domain.Tweet
	var retweet domain.Tweet
	// Delete direct replies to the tweet (not cascading further)
	err := tg.db.Delete(&reply, "replies_to_id = ?", tweet.ID).Error
	if err != nil {
		return err
	}
	// Delete direct retweets of the tweet (not cascading further)
	err = tg.db.Delete(&retweet, "retweets_id = ?", tweet.ID).Error
	if err != nil {
		return err
	}
	// Delete likes of the tweet
	var like domain.Like
	err = tg.db.Delete(&like, "tweet_id = ?", tweet.ID).Error
	if err != nil {
		return err
	}
	// Delete the tweet
	err = tg.db.Delete(tweet, "id = ? AND user_id = ?", tweet.ID, tweet.UserID).Error
	if err != nil {
		return err
	}
	return nil
}