package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/errs"
)

func (s *Server) registerUserRoutes(r *mux.Router) {
	r.HandleFunc("/profile/{user_id:[0-9]+}", s.requireAuth(s.handleProfile)).Methods("GET")
}

// handleProfile handles the route "GET /profile".
// It displays the requested user's basic data and original tweets.
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	// Parse the User ID from the url.
	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if userId <=0 || err != nil{
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Fetch the user from the database, along with their original tweets.
	// TODO: Only get the user here, get their tweets in a separate query.
	user, err := s.us.ByID(userId)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Get original tweets by user id
	// TODO: Handle err.
	originalTweets, err := s.ts.OriginalsByUserID(user.ID)
	user.Tweets = originalTweets

	// Get the images of the user's tweets.
	if err = s.GetTweetImages(user.Tweets); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	if err = s.CountAssociations(user.Tweets); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Get total tweet count
	// TODO: Handle err.
	tweetCount, err := s.ts.CountByUserID(user.ID)
	user.TweetCount = tweetCount

	// Get follower count
	// Get followeds count

	// Return the user.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&user); err != nil {
		errs.LogError(r, err)
		return
	}
}
