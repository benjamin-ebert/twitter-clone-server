package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

//registerTweetRoutes is a helper for registering all tweet routes.
func (s *Server) registerTweetRoutes(r *mux.Router) {
	// Get the authed user's feed.
	r.HandleFunc("/feed/{offset:[0-9]+}", s.requireAuth(s.handleGetFeed)).Methods("GET")

	// Get a specific tweet by id, with its associated user and replies.
	r.HandleFunc("/tweet/{id:[0-9]+}", s.requireAuth(s.handleGetTweet)).Methods("GET")

	// Get one of the three possible subsets of tweets to be displayed on a user's profile.
	// The subsets are: all tweets of the user, the user's original tweets (not a retweet or reply),
	// or tweets of other users that the user has liked.
	r.HandleFunc("/tweets/{subset}/{user_id:[0-9]+}", s.requireAuth(s.handleGetTweets)).Methods("GET")

	// Create a new tweet / retweet / reply. Which one it is, is decided implicitly based on
	// the value of tweet's retweets_id / replies_to_id fields.
	r.HandleFunc("/tweet", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Delete a tweet.
	r.HandleFunc("/tweet/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteTweet)).Methods("DELETE")
}

func (s *Server) handleGetFeed(w http.ResponseWriter, r *http.Request) {
	// Parse the offset from the url.
	offset, err := strconv.Atoi(mux.Vars(r)["offset"])
	if offset < 0 || err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid offset value."))
		return
	}

	// Get the authenticated user.
	authedUser := s.getUserFromContext(r.Context())

	// Get the user's feed.
	var feed []domain.Tweet
	feed, err = s.ts.Index(offset)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Loop over the tweets in the feed and get their images and relevant relations.
	for i, _ := range feed {
		// Get the retrieved feed' images from the filesystem.
		if err = s.SetTweetImages(&feed[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet.
		if err = s.SetTweetAssociationCounts(&feed[i]); err != nil{
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
		s.SetTweetUserAssociationBools(authedUser.ID, &feed[i])
	}

	// Return the feed.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(feed); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleGetTweets gets one of three possible subsets of tweets to be displayed on a user's profile,
// depending on the value of the subset url parameter. Possible values are "all", "with_images" and "liked".
// "all" gets all tweets created by the user. "with_images" gets the user's tweets that contain images.
// "liked" gets all the tweets the user has liked in the past (usually created by other users).
func (s *Server) handleGetTweets(w http.ResponseWriter, r *http.Request) {
	// Parse the user id from the url.
	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if userId <= 0 || err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Get the authenticated user.
	authedUser := s.getUserFromContext(r.Context())

	// Get the tweets according to the value of the subset url parameter.
	subset := mux.Vars(r)["subset"]
	var tweets []domain.Tweet
	switch subset {
	case "all":
		tweets, err = s.ts.ByUserID(userId)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
	case "with_images":
		tweets, err = s.ts.ImageTweetsByUserID(userId)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
	case "liked":
		tweets, err = s.ts.LikedTweetsByUserID(userId)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
	default:
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid tweet subset."))
		return
	}

	// TODO: Add comment.
	for i, _ := range tweets {
		// Get the retrieved tweets' images from the filesystem.
		if err = s.SetTweetImages(&tweets[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet.
		if err = s.SetTweetAssociationCounts(&tweets[i]); err != nil{
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
		s.SetTweetUserAssociationBools(authedUser.ID, &tweets[i])
	}

	// Return the tweets.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweets); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleGetTweet gets one specific tweet, along with its relevant associations and helper data.
func (s *Server) handleGetTweet(w http.ResponseWriter, r *http.Request) {
	// Parse the tweet ID from the url.
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Fetch the tweet from the database.
	tweet, err := s.ts.ByID(id)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Get the authenticated user.
	authedUser := s.getUserFromContext(r.Context())

	// Get the counts of replies, retweets and likes of the tweet.
	if err = s.SetTweetAssociationCounts(tweet); err != nil{
		errs.ReturnError(w, r, err)
		return
	}
	// Get the retrieved tweets' images from the filesystem.
	if err = s.SetTweetImages(tweet); err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
	s.SetTweetUserAssociationBools(authedUser.ID, tweet)

	// TODO: Add comment.
	for i, _ := range tweet.Replies {
		// Get the retrieved tweets' replies' images from the filesystem.
		if err = s.SetTweetImages(&tweet.Replies[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet's replies.
		if err = s.SetTweetAssociationCounts(&tweet.Replies[i]); err != nil{
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
		s.SetTweetUserAssociationBools(authedUser.ID, &tweet.Replies[i])
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleCreateTweet handles the routes:
// "POST /tweet", "POST /reply/:replies_to_id" and "POST /retweet/:retweets_id".
// It determines which type of tweet to create by reading the url parameters, and creates a new
// tweet record in the database. replies_to_id / retweets_id are IDs of an existing tweet.
// TODO: Update comment.
func (s *Server) handleCreateTweet(w http.ResponseWriter, r *http.Request) {
	// Parse the request's json body into a Tweet object.
	var tweet domain.Tweet
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid json body."))
		return
	}

	// Get the authed user's ID and set it as the new Tweet's UserID.
	user := s.getUserFromContext(r.Context())
	tweet.UserID = user.ID

	// Create a new Tweet database record.
	err := s.ts.Create(&tweet)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the created Tweet.
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleDeleteTweet handles the route "DELETE /tweet/delete/:id".
// It soft-deletes a tweet and all it's direct replies and retweets, not cascading further.
// It permanently deletes the tweet's likes and images.
// TODO: Update comment.
func (s *Server) handleDeleteTweet(w http.ResponseWriter, r *http.Request) {
	// Parse the tweet ID from the url.
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Fetch the tweet from the database.
	tweet, err := s.ts.ByID(id)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Check if the tweet belongs to the authed user.
	user := s.getUserFromContext(r.Context())
	if tweet.UserID != user.ID {
		errs.ReturnError(w, r, errs.Errorf(errs.EUNAUTHORIZED, "You are not allowed to delete this tweet."))
		return
	}

	// Soft-delete the tweet and its direct replies and retweets.
	err = s.ts.Delete(tweet)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Permanently delete the tweet's images from disk.
	err = s.is.DeleteAll(domain.OwnerTypeTweet, tweet.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Permanently delete the tweet's replies' images from disk.
	for _, reply := range tweet.Replies {
		if err := s.is.DeleteAll(domain.OwnerTypeTweet, reply.ID); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
	}

	// Permanently delete the tweet's retweets' images from disk.
	for _, retweet := range tweet.Retweets {
		if err := s.is.DeleteAll(domain.OwnerTypeTweet, retweet.ID); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
	}

	// Return the soft-deleted tweet.
	// TODO: Only return this on successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// SetTweetImages takes a pointer to a tweet, finds its images
// in the filesystem and attaches the resulting Image slice to it.
// TODO: Put that into the crud/image.go package?
func (s *Server) SetTweetImages(tweet *domain.Tweet) error {
	images, err := s.is.ByOwner(domain.OwnerTypeTweet, tweet.ID)
	if err != nil {
		return err
	}
	tweet.Images = images
	if tweet.RepliesTo != nil {
		if err = s.SetTweetImages(tweet.RepliesTo); err != nil {
			return err
		}
	}
	if tweet.RetweetsTweet != nil {
		if err = s.SetTweetImages(tweet.RetweetsTweet); err != nil {
			return err
		}
	}
	return nil
}

// SetTweetAssociationCounts a pointer to a tweet, counts its replies, retweets
// and likes and sets those numbers to the according fields.
func (s *Server) SetTweetAssociationCounts(tweet *domain.Tweet) error {
	repliesCount, err := s.ts.CountReplies(tweet.ID)
	if err != nil {
		return err
	}
	tweet.RepliesCount = repliesCount

	retweetsCount, err := s.ts.CountRetweets(tweet.ID)
	if err != nil {
		return err
	}
	tweet.RetweetsCount = retweetsCount

	likesCount, err := s.ts.CountLikes(tweet.ID)
	if err != nil {
		return err
	}
	tweet.LikesCount = likesCount

	if tweet.RepliesTo != nil {
		if err = s.SetTweetAssociationCounts(tweet.RepliesTo); err != nil {
			return err
		}
	}

	if tweet.RetweetsTweet != nil {
		if err = s.SetTweetAssociationCounts(tweet.RetweetsTweet); err != nil {
			return err
		}
	}

	return nil
}

// SetTweetUserAssociationBools takes way too long to say out loud, but as programmer I don't talk
// much anyway. I wanted to be precise here. Takes the ID of the authenticated user and a pointer
// to a tweet. Then checks if the user likes / has replied / retweeted the tweet, and sets the
// boolean value for the respective field on the tweet. If the tweet is a reply or a retweet,
// it recursively does the same to the parent tweet.
// TODO: Update comment, Rename method.
func (s *Server) SetTweetUserAssociationBools(authUserId int, tweet *domain.Tweet) {
	tweet.AuthReplied = s.ts.GetAuthRepliedBool(authUserId, tweet.ID)
	tweet.AuthLike = s.ts.GetAuthLike(authUserId, tweet.ID)
	tweet.AuthRetweet = s.ts.GetAuthRetweet(authUserId, tweet.ID)

	if tweet.RepliesTo != nil {
		s.SetTweetUserAssociationBools(authUserId, tweet.RepliesTo)
	}
	
	if tweet.RetweetsTweet != nil {
		s.SetTweetUserAssociationBools(authUserId, tweet.RetweetsTweet)
	}
}
