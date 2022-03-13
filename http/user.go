package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

func (s *Server) registerUserRoutes(r *mux.Router) {
	// Get the profile data of a specific user.
	r.HandleFunc("/profile/{user_id:[0-9]+}", s.requireAuth(s.handleGetProfile)).Methods("GET")

	// Update the user's data.
	// TODO: Put this into auth.go.
	r.HandleFunc("/profile/update", s.requireAuth(s.handleUpdateProfile)).Methods("PUT")

	// Search for users.
	r.HandleFunc("/search/profiles/{term}", s.requireAuth(s.handleSearchProfiles)).Methods("GET")
}

// handleSearchProfiles handles the route "GET /search/profiles/{term}".
// It parses the search term from the url, runs a user search with it, and returns
// the resulting slice of users (or null if nothing was found).
func (s *Server) handleSearchProfiles(w http.ResponseWriter, r *http.Request) {
	var profiles []domain.User
	searchTerm := mux.Vars(r)["term"]
	profiles = s.us.Search(searchTerm)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(profiles); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleGetProfile handles the route "GET /profile".
// It displays the requested user's basic data and original tweets.
func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Parse the User ID from the url.
	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if userId <= 0 || err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Fetch the user from the database.
	user, err := s.us.ByID(userId)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Check if the authed user is following that user.
	authedUser := s.getUserFromContext(r.Context())
	if authedUser.ID != userId {
		user.AuthFollow = s.us.GetAuthFollow(authedUser.ID, userId)
	}

	// Get the number of tweets, followers and followeds of the user.
	if err = s.SetUserAssociationCounts(user); err != nil {
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

// TODO: Add comments.
// TODO: Put this into auth.go.
func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid update data."))
		return
	}

	err := s.us.Update(r.Context(), &user)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Get the number of tweets, followers and followeds of the user.
	if err = s.SetUserAssociationCounts(&user); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the updated User.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		errs.LogError(r, err)
		return
	}
}

func (s *Server) SetUserAssociationCounts(user *domain.User) error {
	// Get follower count.
	followerCount, err := s.us.CountFollowers(user.ID)
	if err != nil {
		return err
	}
	user.FollowerCount = followerCount

	// Get followed count.
	followedCount, err := s.us.CountFolloweds(user.ID)
	if err != nil {
		return err
	}
	user.FollowedCount = followedCount

	// Get their total tweet count.
	tweetCount, err := s.us.CountTweets(user.ID)
	if err != nil {
		return err
	}
	user.TweetCount = tweetCount

	return nil
}
