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
// If it does not, then this expression becomes invalid and won't compile.
var _ domain.TweetService = &TweetService{}

// Create runs validations needed for creating new Tweet database records.
func (tv *tweetValidator) Create(tweet *domain.Tweet) error {
	err := runTweetValFns(tweet,
		tv.userIdValid,
		tv.repliedToTweetExists,
		tv.retweetedTweetExists,
		// TODO: notAlreadyRetweeted... OR retweetValid, checking for these two an the one above.
		// TODO: retweetedIsNoRetweet
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

// userIdValid ensures that the userId is not empty.
func (tv *tweetValidator) userIdValid(tweet *domain.Tweet) error {
	if tweet.UserID <= 0 {
		return errs.UserIdValid
	}
	return nil
}

// TODO: Add comment.
// TODO: Make this more relevant to the user?
func (tg *tweetGorm) AllWithOffset(offset int) ([]domain.Tweet, error) {
	var feed []domain.Tweet
	// TODO: Limit? Random order? Dynamic offset for faster initial load and endless scroll?
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

// TODO: Add comment.
func (tg *tweetGorm) ByUserID(userId int) ([]domain.Tweet, error) {
	var tweets []domain.Tweet
	err := tg.db.
		Where("user_id = ?", userId).
		Preload("User").
		Preload("RepliesTo.User").
		//Preload("RepliesTo.RetweetsTweet.User").
		Preload("RetweetsTweet.User").
		Preload("RetweetsTweet.RepliesTo.User").
		Order("created_at desc").
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// TODO: Rename this to OrignalsAndRetweets and add comments.
func (tg *tweetGorm) OriginalsByUserID(userId int) ([]domain.Tweet, error) {
	var tweets []domain.Tweet
	err := tg.db.
		Where("user_id = ?", userId).
		Where("replies_to_id IS NULL").
		//Where("retweets_id IS NULL").
		Preload("User").
		Preload("RetweetsTweet.User").
		Preload("RetweetsTweet.RepliesTo.User").
		Order("created_at desc").
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// TODO: Add comment.
func (tg *tweetGorm) ImageTweetsByUserID(userId int) ([]domain.Tweet, error) {
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
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// TODO: Add comment.
func (tg *tweetGorm) LikedTweetsByUserID(userId int) ([]domain.Tweet, error) {
	var tweets []domain.Tweet
	err := tg.db.
		Joins("JOIN likes ON likes.tweet_id=tweets.id").
		Where("likes.user_id = ?", userId).
		Where("likes.deleted_at IS NULL").
		Preload("User").
		Order("created_at desc").
		Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	return tweets, nil
}

// TODO: Add comment.
func (tg *tweetGorm) CountRetweets(id int) (int, error) {
	var count int64
	err := tg.db.Model(&domain.Tweet{}).Where("retweets_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// TODO: Add comment.
func (tg *tweetGorm) CountReplies(id int) (int, error) {
	var count int64
	err := tg.db.Model(&domain.Tweet{}).Where("replies_to_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// TODO: Add comment.
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
func (tg *tweetGorm) GetAuthRepliedBool(userId, tweetId int) bool {
	var count int64
	tg.db.Model(&domain.Tweet{}).Where("user_id = ? AND replies_to_id = ?", userId, tweetId).Count(&count)
	if count > 0 {
		return true
	}
	return false
}

// TODO: Add comment.
func (tg *tweetGorm) GetAuthRetweet(userId, tweetId int) *domain.Tweet {
	var retweet domain.Tweet
	err := tg.db.Where("user_id = ? AND retweets_id = ?", userId, tweetId).First(&retweet).Error
	if err != nil {
		return nil
	}
	return &retweet
}

// TODO: Add comment.
func (tg *tweetGorm) GetAuthLike(userId, tweetId int) *domain.Like {
	var like domain.Like
	err := tg.db.Where("user_id = ? AND tweet_id = ?", userId, tweetId).First(&like).Error
	if err != nil {
		return nil
	}
	return &like
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