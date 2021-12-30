package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
	"wtfTwitter/storage"
)

//registerTweetRoutes is a helper for registering all tweet routes.
func (s *Server) registerTweetRoutes(r *mux.Router) {
	// Create a new original tweet (not a retweet or reply).
	r.HandleFunc("/tweet", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Create a new reply tweet, replying to an existing tweet.
	r.HandleFunc("/reply/{replies_to_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Create a new retweet, retweeting an existing tweet.
	r.HandleFunc("/retweet/{retweets_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")

	// Delete a tweet (regardless which type of tweet).
	r.HandleFunc("/tweet/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteTweet)).Methods("DELETE")

	// Upload images for an existing tweet.
	r.HandleFunc("/tweet/images/upload/{id:[0-9]+}", s.requireAuth(s.handleUploadTweetImages)).Methods("POST")
}

// handleCreateTweet handles the routes:
// "POST /tweet", "POST /reply/:replies_to_id" and "POST /retweet/:retweets_id".
// It determines which type of tweet to create by reading the url parameters, and creates a new
// tweet record in the database. replies_to_id / retweets_id are IDs of an existing tweet.
func (s *Server) handleCreateTweet(w http.ResponseWriter, r *http.Request) {
	// Parse the request's json body into a tweet object.
	var tweet domain.Tweet
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid json body."))
		return
	}

	// Get the authed user's ID and set it as the new tweet's UserID.
	user := s.getUserFromContext(r.Context())
	tweet.UserID = user.ID

	// Parse the ID of the tweet replied to from the url.
	if repliesToId, err := strconv.Atoi(mux.Vars(r)["replies_to_id"]); repliesToId > 0 {
		if err != nil {
			errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
			return
		}
		tweet.RepliesToID = repliesToId
	}

	// Parse the ID of the tweet retweeted from the url.
	if retweetsId, err := strconv.Atoi(mux.Vars(r)["retweets_id"]); retweetsId > 0 {
		if err != nil {
			errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
			return
		}
		tweet.RetweetsID = retweetsId
	}

	// Create the tweet.
	err := s.ts.CreateTweet(&tweet)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the created tweet.
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleDeleteTweet handles the route "DELETE /tweet/delete/:id".
// It soft-deletes a tweet and all it's direct replies and retweets, not cascading further.
// It permanently deletes the tweet's likes and images.
func (s *Server) handleDeleteTweet(w http.ResponseWriter, r *http.Request) {
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
		errs.ReturnError(w, r, errs.Errorf(errs.EUNAUTHORIZED, "You are not allowed to delete this tweet."))
		return
	}

	// Delete the tweet (plus its direct replies, retweets and likes).
	err = s.ts.DeleteTweet(tweet)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Delete the tweets images.
	err = s.is.DeleteAll(domain.OwnerTypeTweet, tweet.ID)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the deleted tweet.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tweet); err != nil {
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
	err = r.ParseMultipartForm(storage.MaxUploadSize)
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