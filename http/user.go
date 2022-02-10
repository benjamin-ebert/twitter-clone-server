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

	// Fetch the user from the database.
	user, err := s.us.ByID(userId)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Get follower count.
	followerCount, err := s.fs.CountFollowers(user.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	user.FollowerCount = followerCount

	// Get followed count.
	followedCount, err := s.fs.CountFollowers(user.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	user.FollowedCount = followedCount

	// Get their total tweet count.
	tweetCount, err := s.ts.CountByUserID(user.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	user.TweetCount = tweetCount

	// Get their original tweets (to populate the default selected tab in the profile view).
	originals, err := s.ts.OriginalsByUserID(user.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	user.Tweets = originals

	// Get their original tweets' images from the filesystem.
	if err = s.GetTweetImages(user.Tweets); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Count their original tweets' replies, retweets and likes.
	if err = s.CountAssociations(user.Tweets); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the user.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&user); err != nil {
		errs.LogError(r, err)
		return
	}
}
