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
	r.HandleFunc("/tweets/{subset}/{user_id:[0-9]+}/{offset}", s.requireAuth(s.handleGetTweets)).Methods("GET")

	// Create a new tweet / retweet / reply. Which one it is, is determined implicitly
	// by the value of tweet's retweets_id / replies_to_id fields.
	r.HandleFunc("/tweet", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Delete a tweet.
	r.HandleFunc("/tweet/{id:[0-9]+}", s.requireAuth(s.handleDeleteTweet)).Methods("DELETE")
}

// handleGetFeed loads a limited number of tweets to be displayed in the home feed.
// Limited, because the frontend doesn't want all tweets at once, but loads more
// tweets as the user scrolls further down. Infinite scroll they call it.
// The offset parameter indicates how many tweets the frontend already has, and therefore
// how many tweets we can skip in the database query. It defaults to 0 on initial load.
// The limit is hard-coded to 10 in the database query, so every load is pretty darn fast.
func (s *Server) handleGetFeed(w http.ResponseWriter, r *http.Request) {
	// Parse the offset from the url.
	offset, err := strconv.Atoi(mux.Vars(r)["offset"])
	if offset < 0 || err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid offset value."))
		return
	}

	// Get the authenticated user.
	authedUser := s.getUserFromContext(r.Context())

	// Get 10 more tweets of the user's feed.
	var feed []domain.Tweet
	feed, err = s.ts.GetFeed(offset)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Loop over those tweets and get their images and associations with the user.
	for i, _ := range feed {
		// Get the retrieved feed' images from the filesystem.
		if err = s.SetTweetImages(&feed[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet.
		if err = s.SetTweetAssociationCounts(&feed[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
		if err = s.SetUserTweetAssociationData(authedUser.ID, &feed[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
	}

	// Return the ten tweets.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(feed); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleGetTweets gets one of four possible subsets of tweets to be displayed on a
// user's profile, depending on the value of the subset url parameter. Possible values
// are "original", "all", "with_images" and "liked". original means retweets and original
// tweets of the user. all means all tweets of the user, including replies. with_images
// means only those tweets of the user that contain images. liked means all tweets the
// user likes. As with the home feed, these four "profile feeds" are loaded using infinite
// scroll, hence the offset url parameter. See handleGetFeed for details on that.
func (s *Server) handleGetTweets(w http.ResponseWriter, r *http.Request) {
	// Parse the user id from the url.
	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if userId <= 0 || err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Parse the offset from the url.
	offset, err := strconv.Atoi(mux.Vars(r)["offset"])
	if offset < 0 || err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid offset value."))
		return
	}

	// Get the authenticated user.
	authedUser := s.getUserFromContext(r.Context())

	// Get the tweets according to the value of the subset url parameter.
	subset := mux.Vars(r)["subset"]
	var tweets []domain.Tweet
	switch subset {
	case "original":
		tweets, err = s.ts.OriginalsByUserID(userId, offset)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
	case "all":
		tweets, err = s.ts.ByUserID(userId, offset)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
	case "with_images":
		tweets, err = s.ts.ImageTweetsByUserID(userId, offset)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
	case "liked":
		tweets, err = s.ts.LikedTweetsByUserID(userId, offset)
		if err != nil {
			errs.ReturnError(w, r, err)
		}
	default:
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid tweet subset."))
		return
	}

	// Get the tweets' images and association data.
	for i, _ := range tweets {
		// Get the retrieved tweets' images from the filesystem.
		if err = s.SetTweetImages(&tweets[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet.
		if err = s.SetTweetAssociationCounts(&tweets[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
		if err = s.SetUserTweetAssociationData(authedUser.ID, &tweets[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
	}

	// Return the tweets.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweets); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleGetTweet gets one specific tweet, along with its replies, images and relevant
// association data.
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
	if err = s.SetTweetAssociationCounts(tweet); err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	// Get the retrieved tweets' images from the filesystem.
	if err = s.SetTweetImages(tweet); err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
	if err = s.SetUserTweetAssociationData(authedUser.ID, tweet); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Get the tweet's replies' images and association data.
	for i, _ := range tweet.Replies {
		// Get the retrieved tweets' replies' images from the filesystem.
		if err = s.SetTweetImages(&tweet.Replies[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Get the counts of replies, retweets and likes of the tweet's replies.
		if err = s.SetTweetAssociationCounts(&tweet.Replies[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		// Determine if the authenticated user has retweeted / replied to / liked the tweet or not.
		if err = s.SetUserTweetAssociationData(authedUser.ID, &tweet.Replies[i]); err != nil {
			errs.ReturnError(w, r, err)
			return
		}
	}

	// Return the tweet.
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleCreateTweet handles the routes: "POST /tweet"
// It reads the tweet data from the posted JSON object, and the user data from the request
// context, and creates a new tweet record in the database. If the posted JSON has values
// in the  replies_to_id or retweets_id fields, the new tweet will be a reply / retweet.
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

// handleDeleteTweet handles the route "DELETE /tweet/:id".
// It soft-deletes a tweet and all it's direct replies and retweets, not cascading further.
// It permanently deletes the tweet's and the tweet's replies' images from the filesystem.
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

	// Return Http Status 204 to indicate successful deletion.
	w.WriteHeader(http.StatusNoContent)
}

// SetTweetImages takes a pointer to a tweet, finds its images
// in the filesystem and attaches the resulting Image slice to it.
// If the tweet is a retweet or a reply and therefore has a "parent" tweet,
// it recursively does the same to the tweet that it retweets / replies to.
func (s *Server) SetTweetImages(tweet *domain.Tweet) error {

	// Get the tweet's images.
	images, err := s.is.ByOwner(domain.OwnerTypeTweet, tweet.ID)
	if err != nil {
		return err
	}
	tweet.Images = images

	// If the tweet is a Reply, get the original tweet's images.
	if tweet.RepliesTo != nil {
		if err = s.SetTweetImages(tweet.RepliesTo); err != nil {
			return err
		}
	}

	// If the tweet is a Retweet, get the original tweet's images.
	if tweet.RetweetsTweet != nil {
		if err = s.SetTweetImages(tweet.RetweetsTweet); err != nil {
			return err
		}
	}

	return nil
}

// SetTweetAssociationCounts takes a pointer to a tweet, counts its replies, retweets
// and likes and sets those numbers to the according fields.
// If the tweet is a retweet or a reply and therefore has a "parent" tweet,
// it recursively does the same to the tweet that it retweets / replies to.
func (s *Server) SetTweetAssociationCounts(tweet *domain.Tweet) error {
	// Get the tweet's count of Replies.
	repliesCount, err := s.ts.CountReplies(tweet.ID)
	if err != nil {
		return err
	}
	tweet.RepliesCount = repliesCount

	// Get the tweet's count of Retweets.
	retweetsCount, err := s.ts.CountRetweets(tweet.ID)
	if err != nil {
		return err
	}
	tweet.RetweetsCount = retweetsCount

	// Get the tweet's count of Likes.
	likesCount, err := s.ts.CountLikes(tweet.ID)
	if err != nil {
		return err
	}
	tweet.LikesCount = likesCount

	// If the tweet is a Reply, repeat the above with the original tweet.
	if tweet.RepliesTo != nil {
		if err = s.SetTweetAssociationCounts(tweet.RepliesTo); err != nil {
			return err
		}
	}

	// If the tweet is a Retweet, repeat the above with the original tweet.
	if tweet.RetweetsTweet != nil {
		if err = s.SetTweetAssociationCounts(tweet.RetweetsTweet); err != nil {
			return err
		}
	}

	return nil
}

// SetUserTweetAssociationData takes way too long to say out loud, but as a programmer
// I don't talk much anyway. It takes the ID of the authenticated user and a pointer
// to a tweet, and checks if the user likes / has replied / retweeted the tweet.
// If the authed user liked or retweeted the tweet, it finds that particular like /
// retweet, and attaches it to the tweet. A boolean value indicates if the authed user
// has replied to the tweet. If the tweet is a retweet or a reply and therefore has
// a "parent" tweet, it recursively does the same to the tweet it retweets / replies to.
func (s *Server) SetUserTweetAssociationData(authUserId int, tweet *domain.Tweet) error {

	// Get the boolean indicating if the authed user has replied to the tweet.
	authReplied, err := s.ts.GetAuthRepliedBool(authUserId, tweet.ID)
	if err != nil {
		return err
	}
	tweet.AuthReplied = authReplied

	// Get the authed user's Like that likes the tweet.
	authLike, err := s.ts.GetAuthLike(authUserId, tweet.ID)
	if err != nil {
		return err
	}
	tweet.AuthLike = authLike

	// Get the authed user's Retweet that retweets the tweet.
	authRetweet, err := s.ts.GetAuthRetweet(authUserId, tweet.ID)
	if err != nil {
		return err
	}
	tweet.AuthRetweet = authRetweet

	// If the tweet is a Reply, repeat the above with the original tweet.
	if tweet.RepliesTo != nil {
		if err = s.SetUserTweetAssociationData(authUserId, tweet.RepliesTo); err != nil {
			return err
		}
	}

	// If the tweet is a Retweet, repeat the above with the original tweet.
	if tweet.RetweetsTweet != nil {
		if err = s.SetUserTweetAssociationData(authUserId, tweet.RetweetsTweet); err != nil {
			return err
		}
	}

	return nil
}
