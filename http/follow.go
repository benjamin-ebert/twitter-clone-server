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
	// Get ten users to be suggested to the authed user as potential follows.
	r.HandleFunc("/follow/suggestions/{user_id:[0-9]+}", s.requireAuth(s.handleGetSuggestions)).Methods("GET")

	// Create a new Follow.
	r.HandleFunc("/follow", s.requireAuth(s.handleCreateFollow)).Methods("POST")

	// Delete an existing Follow.
	r.HandleFunc("/follow/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteFollow)).Methods("DELETE")
}

func (s *Server) handleGetSuggestions(w http.ResponseWriter, r *http.Request) {
	// Parse the User ID from the url.
	userId, err := strconv.Atoi(mux.Vars(r)["user_id"])
	if userId <=0 || err != nil{
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}
	suggestions := s.fs.SuggestFollows(userId)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(suggestions); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleCreateFollow handles the route "POST /follow".
// It reads the followed_id from the json body, gets the authed user's id from context,
// sets their id as the follower_id, and creates a new Follow record in the database.
func (s *Server) handleCreateFollow(w http.ResponseWriter, r *http.Request) {
	// Parse the request's json body into a Follow object. It only contains the followed_id.
	var follow domain.Follow
	if err := json.NewDecoder(r.Body).Decode(&follow); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid json body."))
		return
	}

	// Get the authed user's ID and set it as the new Follow's FollowerID.
	user := s.getUserFromContext(r.Context())
	follow.FollowerID = user.ID

	// Create a new Follow database record.
	err := s.fs.Create(&follow)
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

// handleDeleteFollow handles the route "DELETE /follow/delete/:id".
// It reads id from the url and permanently deletes the respective follow record from the database.
func (s *Server) handleDeleteFollow(w http.ResponseWriter, r *http.Request) {
	// Parse the follow ID from the url.
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Fetch the follow from the database.
	follow, err := s.fs.ByID(id)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Check if the follow belongs to the authed user.
	user := s.getUserFromContext(r.Context())
	if follow.FollowerID != user.ID {
		errs.ReturnError(w, r, errs.Errorf(errs.EUNAUTHORIZED, "You are not allowed to delete this follow."))
		return
	}

	// Soft-delete the follow.
	err = s.fs.Delete(follow)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the soft-deleted follow.
	// TODO: Only return this on successful deletion
	w.WriteHeader(http.StatusNoContent)
}