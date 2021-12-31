package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// registerFollowRoutes is a helper for registering all Follow routes.
func (s *Server) registerFollowRoutes(r *mux.Router) {
	// Create a new Follow (Follow a user).
	r.HandleFunc("/follow/{user_id:[0-9]+}", s.requireAuth(s.handleCreateFollow)).Methods("POST")

	// Delete an existing Follow (Unfollow a user).
	r.HandleFunc("/unfollow/{user_id:[0-9]+}", s.requireAuth(s.handleDeleteFollow)).Methods("DELETE")
}

// handleCreateFollow handles the route "POST /follow/:user_id".
// It reads the ID of the user to be followed from the url and creates
// a new Follow record in the database.
func (s *Server) handleCreateFollow(w http.ResponseWriter, r *http.Request) {
	// Parse the ID of the user to be followed from the url.
	followedId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid user id."))
		return
	}

	// Get the authed user form the request context.
	user := s.getUserFromContext(r.Context())

	// Create a Follow object.
	follow := &domain.Follow{
		FollowedID: followedId,
		FollowerID: user.ID,
	}

	// Create a new Follow database record.
	err = s.fs.Create(follow)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the created Follow.
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(follow); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleDeleteFollow handles the route "DELETE /unfollow/:user_id".
// It reads the ID of the followed user from the url and permanently deletes
// the respective Follow record from the database.
func (s *Server) handleDeleteFollow(w http.ResponseWriter, r *http.Request) {
	// Parse the ID of the followed user from the url.
	followedId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid user id."))
		return
	}

	// Get the authed user form the request context.
	user := s.getUserFromContext(r.Context())

	// Create a Follow object.
	follow := &domain.Follow{
		FollowedID: followedId,
		FollowerID: user.ID,
	}

	// Permanently delete the Follow database record.
	err = s.fs.Delete(follow)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return status code 204 to indicate successful deletion.
	w.WriteHeader(http.StatusNoContent)
}