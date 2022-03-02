package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// registerLikeRoutes is a helper for registering all Like routes.
func (s *Server) registerLikeRoutes(r *mux.Router) {
	// Create a new like for a tweet (Like a tweet).
	r.HandleFunc("/like", s.requireAuth(s.handleCreateLike)).Methods("POST")

	// Delete an existing like of a tweet (Unlike a tweet).
	r.HandleFunc("/like/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteLike)).Methods("DELETE")
}

// handleCreateLike handles the route "POST /like/:tweet_id".
// It reads the tweet ID from the url and creates a new Like record in the database.
func (s *Server) handleCreateLike(w http.ResponseWriter, r *http.Request) {
	// Parse the request's json body into a Like object.
	var like domain.Like
	if err := json.NewDecoder(r.Body).Decode(&like); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid json body."))
		return
	}

	// Get the authed user's ID and set it as the new Like's UserID.
	user := s.getUserFromContext(r.Context())
	like.UserID = user.ID

	// Create a new Like database record.
	err := s.ls.Create(&like)
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
// TODO: Delete by ID, like tweets, then update comment.
func (s *Server) handleDeleteLike(w http.ResponseWriter, r *http.Request) {
	// Parse the like ID from the url.
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Fetch the like from the database.
	like, err := s.ls.ByID(id) // TODO: Implement this.
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Check if the like belongs to the authed user.
	user := s.getUserFromContext(r.Context())
	if like.UserID != user.ID {
		errs.ReturnError(w, r, errs.Errorf(errs.EUNAUTHORIZED, "You are not allowed to delete this like."))
		return
	}

	// Soft-delete the like.
	err = s.ls.Delete(like)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the soft-deleted like.
	// TODO: Only return this on successful deletion
	w.WriteHeader(http.StatusNoContent)
}