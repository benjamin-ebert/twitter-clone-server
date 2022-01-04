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