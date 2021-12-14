package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
)

func (s *Server) registerTweetRoutes(r *mux.Router) {
	r.HandleFunc("/tweet", s.requireAuth(s.handleCreateTweet)).Methods("POST")
	r.HandleFunc("/reply/{replies_to_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")
	r.HandleFunc("/retweet/{retweets_id:[0-9]+}", s.requireAuth(s.handleCreateTweet)).Methods("POST")
	r.HandleFunc("/tweet/delete/{id:[0-9]+}", s.requireAuth(s.handleDeleteTweet)).Methods("DELETE")
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
			 fmt.Println("err converting string route param replies_to to golang int: ", err)
		 }
		 tweet.RepliesToID = repliesToId
	}

	retweetsIdString, found := mux.Vars(r)["retweets_id"]
	if found {
		retweetsId, err := strconv.Atoi(retweetsIdString)
		if err != nil {
			fmt.Println("err converting string route param replies_to to golang int: ", err)
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
			fmt.Println("err converting string route param replies_to to golang int: ", err)
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