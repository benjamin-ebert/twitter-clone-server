package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// registerLikeRoutes is a helper for registering all like routes.
func (s *Server) registerLikeRoutes(r *mux.Router) {
	// Create a new like for a tweet (Like a tweet).
	r.HandleFunc("/like/{tweet_id:[0-9]+}", s.requireAuth(s.handleCreateLike)).Methods("POST")
	// Delete an existing like of a tweet (Unlike a tweet).
	r.HandleFunc("/unlike/{tweet_id:[0-9]+}", s.requireAuth(s.handleDeleteLike)).Methods("DELETE")
}

// handleCreateLike handles the route "POST /like/:tweet_id".
// It reads the tweet ID from the url and creates a new like record in the database.
func (s *Server) handleCreateLike(w http.ResponseWriter, r *http.Request) {
	// Parse the ID of the liked tweet from the url.
	tweetId, err := strconv.Atoi(mux.Vars(r)["tweet_id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid tweet id."))
		return
	}

	// Get the authed user form the request context.
	user := s.getUserFromContext(r.Context())

	// Create a new Like object.
	like := &domain.Like{
		UserID: user.ID,
		TweetID: tweetId,
	}

	// Create the new Like.
	err = s.ls.Create(like)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the created Like.
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(like); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleDeleteLike handles the route "POST /unlike/:tweet_id".
// It reads the tweet ID from the url and permanently deletes the respective like record from the database.
func (s *Server) handleDeleteLike(w http.ResponseWriter, r *http.Request) {
	// Parse the ID of the liked tweet from the url.
	tweetId, err := strconv.Atoi(mux.Vars(r)["tweet_id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid tweet id."))
		return
	}

	// Get the authed user from the request context.
	user := s.getUserFromContext(r.Context())

	// Create a Like object.
	like := &domain.Like{
		UserID: user.ID,
		TweetID: tweetId,
	}

	// Permanently delete the Like.
	err = s.ls.Delete(like)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return status code 204 to indicate successful deletion.
	w.WriteHeader(http.StatusNoContent)
}