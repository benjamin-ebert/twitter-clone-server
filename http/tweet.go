package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
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

	err = json.NewEncoder(w).Encode(&tweet)
	if err != nil {
		fmt.Println("err returning deleted tweet as json: ", err)
	}
}

func (s *Server) handleUploadTweetImages(w http.ResponseWriter, r *http.Request) {
	var tweet domain.Tweet

	idString, found := mux.Vars(r)["id"]
	if found {
		id, err := strconv.Atoi(idString)
		if err != nil {
			fmt.Println("err converting string route param id to golang int: ", err)
		}
		tweet.ID = id
	}

	err := r.ParseMultipartForm(storage.MaxUploadSize)
	if err != nil {
		fmt.Println("err parsing multipart form: ", err)
	}

	files := r.MultipartForm.File["images"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println("err opening file: ", err)
		}
		defer file.Close()
		var img domain.Image
		img.OwnerType = "tweet"
		img.OwnerID = tweet.ID
		img.File = file
		img.Filename = fileHeader.Filename
		err = s.is.Create(&img)
		if err != nil {
			fmt.Println("err storing image: ", err)
		}
	}
	return
}