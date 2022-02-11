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
	// Create a new original tweet (not a retweet or reply).
	r.HandleFunc("/tweet", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Create a new reply tweet, replying to an existing tweet.
	r.HandleFunc("/reply/{replies_to_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Create a new retweet, retweeting an existing tweet.
	r.HandleFunc("/retweet/{retweets_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Delete a tweet (regardless which type of tweet).
	r.HandleFunc("/tweet/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteTweet)).Methods("DELETE")

	// Get one of the three possible subsets of tweets to be displayed on a user's profile.
	r.HandleFunc("/tweets/{subset}/{user_id:[0-9]+}", s.requireAuth(s.handleGetTweets)).Methods("GET")
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
		// TODO: Probably screw this.. just preload the user's likes from like service instead?
		var likes []domain.Like
		likes, err = s.ls.ByUserID(userId)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
		for _, like := range likes {
			tweets = append(tweets, like.Tweet)
		}
	default:
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid tweet subset."))
	}

	// Get the retrieved tweets' images from the filesystem.
	if err = s.GetTweetImages(tweets); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Count the retrieved tweets' replies, retweets and likes.
	if err = s.CountAssociations(tweets); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the tweets.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweets); err != nil {
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

	// If present, parse the ID of the Tweet replied to from the url.
	if repliesToId, err := strconv.Atoi(mux.Vars(r)["replies_to_id"]); repliesToId > 0 {
		if err != nil {
			errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
			return
		}
		tweet.RepliesToID = repliesToId
	}

	// If present, parse the ID of the Tweet retweeted from the url.
	if retweetsId, err := strconv.Atoi(mux.Vars(r)["retweets_id"]); retweetsId > 0 {
		if err != nil {
			errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
			return
		}
		tweet.RetweetsID = retweetsId
	}

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
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}

// GetTweetImages takes an array of Tweet objects, finds each Tweet's images
// in the filesystem and attaches the resulting Image array to the Tweet object.
// TODO: Put that into the crud/image.go package?
func (s *Server) GetTweetImages(tweets []domain.Tweet) error {
	for i, tweet := range tweets {
		images, err := s.is.ByOwner(domain.OwnerTypeTweet, tweet.ID)
		if err != nil {
			return err
		}
		tweets[i].Images = images
	}
	return nil
}

// CountAssociations takes a slice of Tweet objects, iterates over it and gets the
// number of replies, retweets and likes of each tweet.
// TODO: Put that into the crud/tweet.go?
func (s *Server) CountAssociations(tweets []domain.Tweet) error {
	for i, tweet := range tweets {
		repliesCount, err := s.ts.CountReplies(tweet.ID)
		if err != nil {
			return err
		}
		tweets[i].RepliesCount = repliesCount

		retweetsCount, err := s.ts.CountRetweets(tweet.ID)
		if err != nil {
			return err
		}
		tweets[i].RetweetsCount = retweetsCount

		likesCount, err := s.ts.CountLikes(tweet.ID)
		if err != nil {
			return err
		}
		tweets[i].LikesCount = likesCount
	}
	return nil
}