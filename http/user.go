package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
	"wtfTwitter/storage"
)

// registerUserRoutes is a helper for registering all user routes.
func (s *Server) registerUserRoutes(r *mux.Router) {
	// View the authed user and his relationships.
	r.HandleFunc("/profile", s.requireAuth(s.handleProfile)).Methods("GET")

	// Upload the authed user's avatar or header image.
	r.HandleFunc("/user/{image_type}/upload", s.requireAuth(s.handleUploadUserImages)).Methods("POST")

	// Delete the authed user's avatar or header image.
	r.HandleFunc("/user/{image_type}/delete", s.requireAuth(s.handleDeleteUserImages)).Methods("DELETE")
}

// handleProfile handles the route "GET /profile".
// It displays the authed user's data and relationships.
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	// Get the authed user from the request context.
	u := s.getUserFromContext(r.Context())

	// Fetch the user from the database, to get all his tweets and likes.
	user, err := s.us.ByID(u.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Get the images of the user's tweets.
	for i, tweet := range user.Tweets {
		images, err := s.is.ByOwner(domain.OwnerTypeTweet, tweet.ID)
		if err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		user.Tweets[i].Images = images
	}

	// Return the user.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&user); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleUploadUserImages handles the route "POST /user/:image_type/upload".
// It reads an uploaded image, determines if it's an avatar or a header, stores it on disk
// and updates the user's avatar / header path in the database. On success, it deletes
// the user's previous avatar / header image from disk and returns the updated user.
func (s *Server) handleUploadUserImages(w http.ResponseWriter, r *http.Request) {
	// Parse the image type from the url.
	imgType := mux.Vars(r)["image_type"]
	if imgType != "avatar" && imgType != "header" {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid image type, must be 'avatar' or 'header'."))
		return
	}

	// Parse the data to be uploaded.
	err := r.ParseMultipartForm(storage.MaxUploadSize)
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, errs.ErrorMessage(err)))
		return
	}

	// Get the authed user from the request context.
	user := s.getUserFromContext(r.Context())

	// Open the image.
	imageHeader := r.MultipartForm.File["image"][0]
	image, err := imageHeader.Open()
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	defer image.Close()

	// Parse it into an Image object.
	img := &domain.Image{
		OwnerType: domain.OwnerTypeUser,
		OwnerID: user.ID,
		File: image,
		Filename: imageHeader.Filename,
	}

	// Save the image to disk (includes validation / normalization).
	err = s.is.Create(img)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Update the user's avatar/header field to be the Filename of the newly stored image.
	if imgType == "avatar" {
		user.Avatar = img.Filename
	} else {
		user.Header = img.Filename
	}

	// Update the user record in the database.
	err = s.us.Update(r.Context(), user)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Delete any old avatar/header images of the user.
	userImages, err := s.is.ByOwner(domain.OwnerTypeUser, user.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	for _, img := range userImages {
		if img.Filename != user.Avatar && img.Filename != user.Header {
			err := s.is.Delete(&img)
			if err != nil {
				errs.ReturnError(w, r, err)
				return
			}
		}
	}

	// Return the user.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&user); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleDeleteUserImages handles the route "DELETE /user/:image_type/delete".
// It determines which user image should be deleted (avatar or header), deletes the file
// from disk and updates the avatar / header path in the database to be empty.
// On success, it returns the updated user.
func (s *Server) handleDeleteUserImages(w http.ResponseWriter, r *http.Request) {
	// Parse the image type from the url.
	imgType := mux.Vars(r)["image_type"]
	if imgType != "avatar" && imgType != "header" {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid image type, must be 'avatar' or 'header'."))
		return
	}

	// Get the user from the request context.
	user := s.getUserFromContext(r.Context())

	// Get the user's Avatar or Header image filename, based on which one will be deleted.
	var filename string
	if imgType == "avatar" {
		filename = user.Avatar
	} else {
		filename = user.Header
	}

	// Create an Image object.
	 img := &domain.Image{
		 OwnerType: domain.OwnerTypeUser,
		 OwnerID: user.ID,
		 Filename: filename,
	 }

	 // Delete the image.
	 err := s.is.Delete(img)
	 if err != nil {
		  errs.ReturnError(w, r, err)
		  return
	 }

	 // Set the user's Avatar or Header field empty, based on which image has been deleted.
	 if imgType == "avatar" {
		  user.Avatar = ""
	 } else {
		  user.Header = ""
	 }

	 // Update the user's database record.
	 err = s.us.Update(r.Context(), user)
	 if err != nil {
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
