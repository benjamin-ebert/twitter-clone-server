package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
	"wtfTwitter/storage"
)

func (s *Server) registerTweetRoutes(r *mux.Router) {
	r.HandleFunc("/tweet", s.requireAuth(s.handleCreateTweet)).Methods("POST")
	r.HandleFunc("/reply/{replies_to_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")
	r.HandleFunc("/retweet/{retweets_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")
	r.HandleFunc("/tweet/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteTweet)).Methods("DELETE")
	r.HandleFunc("/tweet/images/upload/{id:[0-9]+}", s.requireAuth(s.handleUploadTweetImages)).Methods("POST")
}

func (s *Server) handleCreateTweet(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("err reading tweet data from request body: ", err)
	}

	var tweet domain.Tweet
	err = json.Unmarshal(data, &tweet)
	if err != nil {
		fmt.Println("err unmarshalling tweet data into tweet obj: ", err)
	}

	user := s.getUserFromContext(r.Context())
	tweet.UserID = user.ID

	repliesToIdString, found := mux.Vars(r)["replies_to_id"]
	if found {
		 repliesToId, err := strconv.Atoi(repliesToIdString)
		 if err != nil {
			 fmt.Println("err converting string route param replies_to_id to golang int: ", err)
		 }
		 tweet.RepliesToID = repliesToId
	}

	retweetsIdString, found := mux.Vars(r)["retweets_id"]
	if found {
		retweetsId, err := strconv.Atoi(retweetsIdString)
		if err != nil {
			fmt.Println("err converting string route param retweets_id to golang int: ", err)
		}
		tweet.RetweetsID = retweetsId
	}

	err = s.ts.CreateTweet(&tweet)
	if err != nil {
		fmt.Println("err creating new tweet: ", err)
	}

	err = json.NewEncoder(w).Encode(&tweet)
	if err != nil {
		fmt.Println("err returning tweet as json: ", err)
	}
}

func (s *Server) handleDeleteTweet(w http.ResponseWriter, r *http.Request) {
	var tweet domain.Tweet

	deleteIdString, found := mux.Vars(r)["id"]
	if found {
		deleteId, err := strconv.Atoi(deleteIdString)
		if err != nil {
			fmt.Println("err converting string route param id to golang int: ", err)
		}
		tweet.ID = deleteId
	}

	user := s.getUserFromContext(r.Context())
	tweet.UserID = user.ID

	err := s.ts.DeleteTweet(&tweet)
	if err != nil {
		fmt.Println("err deleting tweet: ", err)
	}

	// THIS RUNS EVEN IF IT FAILED TO DELETE THE TWEET BECAUSE THE USER IS NOT THE OWNER. FIX!
	images, err := s.is.ByOwner(domain.OwnerTypeTweet, tweet.ID)
	if err != nil {
		fmt.Println("err retrieving tweet images")
	}
	for _, img := range images {
		err := s.is.Delete(&img)
		if err != nil {
			fmt.Println("err deleting tweet image")
		}
	}

	err = json.NewEncoder(w).Encode(&tweet)
	if err != nil {
		fmt.Println("err returning deleted tweet as json: ", err)
	}
}

func (s *Server) handleUploadTweetImages(w http.ResponseWriter, r *http.Request) {
	// Parse tweet ID from the url.
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid Id format."))
		return
	}

	// Fetch tweet from the database.
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
		 // Parse it to an Image object
		 img := &domain.Image{
			 OwnerType: domain.OwnerTypeTweet,
			 OwnerID: id,
			 File: file,
			 Filename: fileHeader.Filename,
		 }
		 // Save the image to disk (includes validation / normalization)
		 err = s.is.Create(img)
		 if err != nil {
			 errs.ReturnError(w, r, err)
			 return
		 }
	}

	// Fetch all the tweet's images
	images, err := s.is.ByOwner(domain.OwnerTypeTweet, id)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	tweet.Images = images

	// Return the tweet with its images.
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(tweet); err != nil {
		errs.LogError(r, err)
		return
	}
}