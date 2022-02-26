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
	// Get a specific tweet by id, with its associated user and replies.
	r.HandleFunc("/tweet/{id:[0-9]+}", s.requireAuth(s.handleGetTweet)).Methods("GET")

	// Get one of the three possible subsets of tweets to be displayed on a user's profile.
	// The subsets are: all tweets of the user, the user's original tweets (not a retweet or reply),
	// or tweets of other users that the user has liked.
	r.HandleFunc("/tweets/{subset}/{user_id:[0-9]+}", s.requireAuth(s.handleGetTweets)).Methods("GET")

	// Create a new tweet.
	r.HandleFunc("/tweet", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Delete a tweet.
	r.HandleFunc("/tweet/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteTweet)).Methods("DELETE")
}

// handleGetTweets gets one of three possible subsets of tweets to be displayed on a user's profile,
// depending on the value of the subset url parameter. Possible values are "all", "with_images" and "liked".
// "all" gets all tweets created by the user. "with_images" gets the user's tweets that contain images.
// "liked" gets all the tweets the user has liked in the past (usually created by other users).
func (s *Server) handleGetTweets(w http.ResponseWriter, r *http.Request) {
	// Parse the requested tweet sub set from the url. Return error if parameter invalid.
	subset := mux.Vars(r)["subset"]
	if subset != "all" && subset != "with_images" && subset != "liked" {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid tweet sub set, must be 'all', 'with_images' or 'liked'."))
		return
	}

	// Parse the user id from the url.
	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if userId <=0 || err != nil{
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Get the authenticated user.
	authedUser := s.getUserFromContext(r.Context())

	// Get the tweets according to the value of the subset url parameter.
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
	}

	// TODO: Add comment.
	for i, _ := range tweets {
		// Get the retrieved tweets' images from the filesystem.
		if err = s.GetTweetImages(&tweets[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet.
		if err = s.CountAssociations(&tweets[i]); err != nil{
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user likes the tweet or not.
		s.GetAuthLikesBool(authedUser.ID, &tweets[i])
		// Determine if the authenticated user has replied to the tweet or not.
		s.GetAuthRepliedBool(authedUser.ID, &tweets[i])
	}

	// Return the tweets.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweets); err != nil {
		errs.LogError(r, err)
		return
	}
}

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
	if err = s.CountAssociations(tweet); err != nil{
		errs.ReturnError(w, r, err)
		return
	}
	// Get the retrieved tweets' images from the filesystem.
	if err = s.GetTweetImages(tweet); err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	// Determine if the authenticated user likes the tweet or not.
	s.GetAuthLikesBool(authedUser.ID, tweet)
	// Determine if the authenticated user has replied to the tweet or not.
	s.GetAuthRepliedBool(authedUser.ID, tweet)

	// TODO: Add comment.
	for i, _ := range tweet.Replies {
		// Get the retrieved tweets' replies' images from the filesystem.
		if err = s.GetTweetImages(&tweet.Replies[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet's replies.
		if err = s.CountAssociations(&tweet.Replies[i]); err != nil{
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user likes the reply or not.
		s.GetAuthLikesBool(authedUser.ID, &tweet.Replies[i])
		// Determine if the authenticated user has replied to the tweet or not.
		s.GetAuthRepliedBool(authedUser.ID, &tweet.Replies[i])
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
	// TODO: Really? Better return 204 and empty response. Maybe for all delete operations.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}

// GetTweetImages takes a pointer to a tweet, finds its images
// in the filesystem and attaches the resulting Image slice to it.
// TODO: Put that into the crud/image.go package?
// TODO: Add parameter to pick if the images for related tweets should also be queried?
func (s *Server) GetTweetImages(tweet *domain.Tweet) error {
	images, err := s.is.ByOwner(domain.OwnerTypeTweet, tweet.ID)
	if err != nil {
		return err
	}
	tweet.Images = images
	if tweet.RepliesTo != nil {
		if err = s.GetTweetImages(tweet.RepliesTo); err != nil {
			return err
		}
	}
	return nil
}

// CountAssociations a pointer to a tweet, counts its replies, retweets
// and likes and sets those numbers to the according fields.
func (s *Server) CountAssociations(tweet *domain.Tweet) error {
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
		if err = s.CountAssociations(tweet.RepliesTo); err != nil {
			return err
		}
	}

	return nil
}

// GetAuthLikesBool takes a tweet and the ID of the authenticated user, checks if
// the user likes the tweet or not, and sets the according field on the tweet.
// TODO: Put that and the other helpers into the crud/tweet.go?
func (s *Server) GetAuthLikesBool(authUserId int, tweet *domain.Tweet) {
	tweet.AuthLikes = s.ls.CheckAuthLikes(authUserId, tweet.ID)
	if tweet.RepliesTo != nil {
		s.GetAuthLikesBool(authUserId, tweet.RepliesTo)
	}
}

// GetAuthRepliedBool takes a tweet and the ID of the authenticated user, checks if
// the user has replied the tweet or not, and sets the according field on the tweet.
func (s *Server) GetAuthRepliedBool(authUserId int, tweet *domain.Tweet) {
	tweet.AuthReplied = s.ts.CheckAuthReplied(authUserId, tweet.ID)
	if tweet.RepliesTo != nil {
		s.GetAuthRepliedBool(authUserId, tweet.RepliesTo)
	}
}