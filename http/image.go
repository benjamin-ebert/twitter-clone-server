package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/crud"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// registerUserRoutes is a helper for registering all user routes.
func (s *Server) registerImageRoutes(r *mux.Router) {
	// Upload the authed user's avatar or header image.
	r.HandleFunc("/upload/user/{image_type}", s.requireAuth(s.handleUploadUserImages)).Methods("POST")

	// Delete the authed user's avatar or header image.
	r.HandleFunc("/delete/user/{image_type}", s.requireAuth(s.handleDeleteUserImages)).Methods("DELETE")

	// Upload images for an existing tweet.
	r.HandleFunc("/upload/tweet/{id:[0-9]+}", s.requireAuth(s.handleUploadTweetImages)).Methods("POST")
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
	err := r.ParseMultipartForm(crud.MaxUploadSize)
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

// handleUploadTweetImages handles the route "POST /tweet/images/upload/:id"
// It reads up to 4 uploaded images for a tweet and stores them on disk.
// Their storage location determines which tweet they belong to. They are not stored in the database.
func (s *Server) handleUploadTweetImages(w http.ResponseWriter, r *http.Request) {
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
		errs.ReturnError(w, r, errs.Errorf(errs.EUNAUTHORIZED, "You are not allowed to edit this tweet."))
		return
	}

	// Parse the data to be uploaded.
	err = r.ParseMultipartForm(crud.MaxUploadSize)
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, errs.ErrorMessage(err)))
		return
	}

	// Check if the image count is max 4.
	files := r.MultipartForm.File["images"]
	if len(files) > 4 {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Too many images, not more than 4 allowed."))
		return
	}

	// Delete all existing images of the tweet. Not necessary if the API call comes from
	// the frontend app, since the GUI won't allow users to update existing tweets.
	// It's to prevent potential non-GUI API calls from uploading infinite images.
	err = s.is.DeleteAll(domain.OwnerTypeTweet, tweet.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Process the images.
	for _, fileHeader := range files {
		// Open the image.
		file, err := fileHeader.Open()
		if err != nil {
			errs.ReturnError(w, r, err)
			return
		}
		defer file.Close()
		// Parse it into an Image object.
		img := &domain.Image{
			OwnerType: domain.OwnerTypeTweet,
			OwnerID: id,
			File: file,
			Filename: fileHeader.Filename,
		}
		// Save the image to disk (includes validation / normalization).
		err = s.is.Create(img)
		if err != nil {
			errs.ReturnError(w, r, err)
			return
		}
	}

	// Fetch the tweet's images.
	images, err := s.is.ByOwner(domain.OwnerTypeTweet, id)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	tweet.Images = images

	// Return the tweet with its images.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}
