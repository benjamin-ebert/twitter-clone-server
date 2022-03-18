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
	r.HandleFunc("/profile/update", s.requireAuth(s.handleUpdateProfile)).Methods("PUT")

	// Search for users.
	r.HandleFunc("/search/profiles/{term}", s.requireAuth(s.handleSearchProfiles)).Methods("GET")
}

// handleSearchProfiles handles the route "GET /search/profiles/{term}".
// It parses the search term from the url, runs a user search with it, and returns
// the resulting slice of users (or null if nothing was found).
func (s *Server) handleSearchProfiles(w http.ResponseWriter, r *http.Request) {
	// Declare a slice to hold the resulting profiles (User objects).
	var profiles []domain.User

	// Parse the search term from the url.
	searchTerm := mux.Vars(r)["term"]

	// Search the database for users that are relevant to the term.
	profiles = s.us.Search(searchTerm)

	// Return the result.
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
		authFollow, err := s.us.GetAuthFollow(authedUser.ID, userId)
		if err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		user.AuthFollow = authFollow
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

// handleUpdateProfile handles the route "PUT /profile/update".
// It reads user data from the json body and updates the user record in the database.
func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Parse the request's json body into a User object.
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid update data."))
		return
	}

	// Check if the authed user is allowed to update that user record.
	authedUser := s.getUserFromContext(r.Context())
	if authedUser.ID != user.ID {
		errs.ReturnError(w, r, errs.Errorf(errs.EUNAUTHORIZED, "You are not allowed to updated this user."))
		return
	}

	// Update the authenticated user with the data in the User object.
	err := s.us.Update(&user)
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

// SetUserAssociationCounts takes a pointer to a user object, counts its followes,
// followeds and tweets, and sets those numbers to the according fields.
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
