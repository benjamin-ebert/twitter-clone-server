package crud

import (
	"gorm.io/gorm"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode/utf8"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// TweetService manages Tweets.
// It implements the domain.TweetService interface.
type TweetService struct {
	tweetValidator
}

// tweetValidator runs validations on incoming Tweet data.
// On success, it passes the data on to tweetGorm.
// Otherwise, it returns the error of the validation that has failed.
type tweetValidator struct {
	tweetGorm
}

// tweetGorm runs CRUD operations on the database using incoming Tweet data.
// It assumes that data has been validated. On success, it returns nil.
// Otherwise, it returns the error of the operation that has failed.
type tweetGorm struct {
	db *gorm.DB
}

// NewTweetService returns an instance of TweetService.
func NewTweetService(db *gorm.DB) *TweetService {
	return &TweetService{
		tweetValidator{
			tweetGorm{
				db: db,
			},
		},
	}
}

// Ensure the TweetService struct properly implements the domain.TweetService interface.
// If it does not, then this expression becomes invalid and the program won't compile.
var _ domain.TweetService = &TweetService{}

// Create runs validations needed for creating new Tweet database records.
func (tv *tweetValidator) Create(tweet *domain.Tweet) error {
	err := runTweetValFns(tweet,
		tv.userIdValid,
		tv.repliedToTweetExists,
		tv.retweetedTweetExists,
		tv.retweetedTweetIsNoRetweet,
		tv.notAlreadyRetweeted,
		tv.contentMinLength,
		tv.contentMaxLength)
	if err != nil {
		return err
	}
	return tv.tweetGorm.Create(tweet)
}

// Delete runs validations needed for deleting existing Tweet database records.
func (tv *tweetValidator) Delete(tweet *domain.Tweet) error {
	err := runTweetValFns(tweet, tv.idValid)
	if err != nil {
		return err
	}
	return tv.tweetGorm.Delete(tweet)
}

// runTweetValFns runs any number of functions of type tweetValFn on the passed in Tweet object.
// If none of them returns an error, it returns nil. Otherwise, it returns the respective error.
func runTweetValFns(tweet *domain.Tweet, fns ...tweetValFn) error {
	for _, fn := range fns {
		if err := fn(tweet); err != nil {
			return err
		}
	}
	return nil
}

// A tweetValFn is any function that takes in a pointer to a domain.Tweet object and returns an error.
type tweetValFn = func(tweet *domain.Tweet) error

// contentMinLength makes sure that the Tweet's content is not empty...
// ...unless it's a Retweet, in which case empty content is expected.
func (tv *tweetValidator) contentMinLength(tweet *domain.Tweet) error {
	if tweet.RetweetsID == nil {
		contentStripped := strings.ReplaceAll(tweet.Content, " ", "")
		if contentStripped == "" {
			return errs.Errorf(errs.EINVALID, "Tweet content must not be empty.")
		}
	}
	return nil
}

// contentMaxLength makes sure that the Tweet's content does not exceed the maximum content length.
func (tv *tweetValidator) contentMaxLength(tweet *domain.Tweet) error {
	if utf8.RuneCountInString(tweet.Content) > 280 {
		return errs.Errorf(errs.EINVALID, "Tweet content max length is 280 characters.")
	}
	return nil
}

// idValid makes sure that the passed in ID of a Tweet to be deleted is greater than 0.
func (tv *tweetValidator) idValid(tweet *domain.Tweet) error {
	if tweet.ID <= 0 {
		return errs.IdInvalid
	}
	return nil
}

// notAlreadyRetweeted makes sure that the tweet to be retweeted has not yet been
// retweeted by the user already. A user can must only retweet a tweet once.
func (tv *tweetValidator) notAlreadyRetweeted(tweet *domain.Tweet) error {
	if tweet.RetweetsID != nil {
		err := tv.tweetGorm.db.First(&domain.Tweet{}, "user_id = ? AND retweets_id = ?", tweet.UserID, tweet.RetweetsID).Error
		if err == nil {
			return errs.Errorf(errs.EINVALID, "You already retweeted that tweet.")
		}
	}
	return nil
}

// repliedToTweetExists makes sure that the Tweet to be replied to actually exists.
// This check only runs if the incoming Tweet object has a valid ID in its RepliesToID field.
func (tv *tweetValidator) repliedToTweetExists(tweet *domain.Tweet) error {
	if tweet.RepliesToID != nil {
		err := tv.db.First(&domain.Tweet{}, "id = ?", tweet.RepliesToID).Error
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

// retweetedTweetExists makes sure that the Tweet to be retweeted actually exists.
// This check only runs if the incoming Tweet object has a valid ID in its RetweetsID field.
func (tv *tweetValidator) retweetedTweetExists(tweet *domain.Tweet) error {
	if tweet.RetweetsID != nil {
		err := tv.tweetGorm.db.First(&domain.Tweet{}, "id = ?", tweet.RetweetsID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errs.Errorf(errs.ENOTFOUND, "The retweeted tweet does not exist.")
			} else {
				return err
			}
		}
	}
	return nil
}

// retweetedTweetIsNoRetweet makes sure that the tweet to be retweeted isn't a retweet
// itself. Retweets must not be retweeted. Only originals and replies can be retweeted.
func (tv *tweetValidator) retweetedTweetIsNoRetweet(tweet *domain.Tweet) error {
	if tweet.RetweetsID != nil {
		var original domain.Tweet
		tv.tweetGorm.db.First(&original, "id = ?", tweet.RetweetsID)
		if original.RetweetsID != nil {
			return errs.Errorf(errs.EINVALID, "You cannot retweet a retweet.")
		}
	}
	return nil
}

// userIdValid ensures that the userId is not empty.
func (tv *tweetValidator) userIdValid(tweet *domain.Tweet) error {
	if tweet.UserID <= 0 {
		return errs.UserIdValid
	}
	return nil
}

// GetFeed loads 10 tweets to be displayed on the home feed. The frontend makes request
// for more tweets whenever the user reaches the bottom scrolling down. The offset
// is the number of tweets that are already visible in the feed, which means they don't
// need to be queried again. The tweets are loaded with their relevant associations.
func (tg *tweetGorm) GetFeed(offset int) ([]domain.Tweet, error) {
	var feed []domain.Tweet
	err := tg.db.
		Preload("User").
		Preload("RepliesTo.User").
		Preload("RetweetsTweet.User").
		Preload("RetweetsTweet.RepliesTo.User").
		Order("created_at desc").
		Offset(offset).
		Limit(10).
		Find(&feed).Error
	if err != nil {
		return nil, err
	}
	return feed, nil
}

// ByID retrieves a single Tweet by ID, along with its associated Replies and Retweets.
// If the record doesn't exist, it returns errs.ENOTFOUND. Otherwise, it returns nil.
func (tg *tweetGorm) ByID(id int) (*domain.Tweet, error) {
	var tweet domain.Tweet
	err := tg.db.
		Preload("User").
		Preload("Replies.User").
		First(&tweet, "id = ?", id).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.Errorf(errs.ENOTFOUND, "The tweet does not exist.")
		} else {
			return nil, err
		}
	}
	return &tweet, nil
}

// ByUserID finds the specified user's tweets, retweets and replies.
// It also takes an offset and uses a limit, because these tweets are loaded and
// displayed incrementally as people scroll further down the user's profile.
func (tg *tweetGorm) ByUserID(userId, offset int) ([]domain.Tweet, error) {
	var tweets []domain.Tweet
	err := tg.db.
		Where("user_id = ?", userId).
		Preload("User").
		Preload("RepliesTo.User").
		Preload("RetweetsTweet.User").
		Preload("RetweetsTweet.RepliesTo.User").
		Order("created_at desc").
		Offset(offset).
		Limit(10).
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// OriginalsByUserID finds the specified user's tweets and retweets.
// It also takes an offset and uses a limit, because these tweets are loaded and
// displayed incrementally as people scroll further down the user's profile.
func (tg *tweetGorm) OriginalsByUserID(userId, offset int) ([]domain.Tweet, error) {
	var tweets []domain.Tweet
	err := tg.db.
		Where("user_id = ?", userId).
		Where("replies_to_id IS NULL").
		Preload("User").
		Preload("RetweetsTweet.User").
		Preload("RetweetsTweet.RepliesTo.User").
		Order("created_at desc").
		Offset(offset).
		Limit(10).
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// ImageTweetsByUserID finds the specified user's tweets that contain images.
// It also takes an offset and uses a limit, because these tweets are loaded and
// displayed incrementally as people scroll further down the user's profile.
func (tg *tweetGorm) ImageTweetsByUserID(userId, offset int) ([]domain.Tweet, error) {
	files, err := ioutil.ReadDir(domain.ImagesBaseDir + "/" + domain.OwnerTypeTweet + "/")
	if err != nil {
		return nil, err
	}
	var imageTweetIds []int
	for _, f := range files {
		if f.IsDir() {
			id, err := strconv.Atoi(f.Name())
			if err != nil {
				return nil, err
			}
			imageTweetIds = append(imageTweetIds, id)
		}
	}
	var tweets []domain.Tweet
	err = tg.db.
		Where("user_id = ?", userId).
		Where("id IN ?", imageTweetIds).
		Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(10).
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// LikedTweetsByUserID finds all tweets that the user with the specified id likes.
// It also takes an offset and uses a limit, because these tweets are loaded and
// displayed incrementally as people scroll further down the user's profile.
func (tg *tweetGorm) LikedTweetsByUserID(userId, offset int) ([]domain.Tweet, error) {
	var tweets []domain.Tweet
	err := tg.db.
		Joins("JOIN likes ON likes.tweet_id=tweets.id").
		Where("likes.user_id = ?", userId).
		Preload("User").
		Preload("RepliesTo.User").
		Order("created_at desc").
		Offset(offset).
		Limit(10).
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// CountReplies takes a tweet ID, counts that tweet's Replies and returns the
// integer result and a nil error. If there is an error, it returns 0 and the error.
func (tg *tweetGorm) CountReplies(id int) (int, error) {
	var count int64
	err := tg.db.Model(&domain.Tweet{}).Where("replies_to_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// CountRetweets takes a tweet ID, counts that tweet's Retweets and returns the
// integer result  and a nil error. If there is an error, it returns 0 and the error.
func (tg *tweetGorm) CountRetweets(id int) (int, error) {
	var count int64
	err := tg.db.Model(&domain.Tweet{}).Where("retweets_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// CountLikes takes a tweet ID, counts that tweet's Likes and returns the
// integer result and a nil error. If there is an error, it returns 0 and the error.
func (tg *tweetGorm) CountLikes(id int) (int, error) {
	var count int64
	err := tg.db.Model(&domain.Like{}).Where("tweet_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetAuthRepliedBool takes a user ID and a tweet ID and returns a boolean expressing whether
// the given user has replied to the given tweet.
func (tg *tweetGorm) GetAuthRepliedBool(userId, tweetId int) (bool, error) {
	var authReply domain.Tweet
	err := tg.db.First(&authReply, "user_id = ? AND replies_to_id = ?", userId, tweetId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetAuthRetweet takes a user ID and a tweet ID. It looks for a retweet that belongs
// to the specified user, and that retweets the specified tweet.
// It returns a pointer to that retweet if it exists, otherwise it returns nil.
func (tg *tweetGorm) GetAuthRetweet(userId, tweetId int) (*domain.Tweet, error) {
	var retweet domain.Tweet
	err := tg.db.Where("user_id = ? AND retweets_id = ?", userId, tweetId).First(&retweet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &retweet, nil
}

// GetAuthLike takes a user ID and a tweet ID. It looks for a Like that belongs
// to the specified user, and that likes the specified tweet.
// It returns a pointer to that Like if it exists, otherwise it returns nil.
func (tg *tweetGorm) GetAuthLike(userId, tweetId int) (*domain.Like, error) {
	var like domain.Like
	err := tg.db.Where("user_id = ? AND tweet_id = ?", userId, tweetId).First(&like).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &like, nil
}

// Create stores the data from the Tweet object in a new database record.
func (tg *tweetGorm) Create(tweet *domain.Tweet) error {
	if err := tg.db.Create(tweet).Error; err != nil {
		return err
	}
	if err := tg.db.Preload("User").First(&tweet).Error; err != nil {
		return err
	}
	return nil
}

// Delete soft-deletes a Tweet record from the database, along with its associated
// Replies, Retweets (not cascading to delete their Replies / Retweets) and Likes.
func (tg *tweetGorm) Delete(tweet *domain.Tweet) error {
	err := tg.db.Select("Replies", "Retweets", "Likes").Delete(tweet).Error
	if err != nil {
		return err
	}
	return nil
}
