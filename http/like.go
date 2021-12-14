package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
)

func (s *Server) registerLikeRoutes(r *mux.Router) {
	r.HandleFunc("/like/{tweet_id:[0-9]+}", s.requireAuth(s.handleCreateLike)).Methods("POST")
	r.HandleFunc("/unlike/{tweet_id:[0-9]+}", s.requireAuth(s.handleDeleteLike)).Methods("POST")
}

func (s *Server) handleCreateLike(w http.ResponseWriter, r *http.Request) {
	var like domain.Like

	params, found := mux.Vars(r)["tweet_id"]
	if found {
		tweetId, err := strconv.Atoi(params)
		if err != nil {
			fmt.Println("err converting string route param tweet_id to golang int: ", err)
		}
		like.TweetID = tweetId
	}

	user := s.getUserFromContext(r.Context())
	like.UserID = user.ID

	err := s.ls.Create(&like)
	if err != nil {
		fmt.Println("err creating new like: ", err)
	}

	err = json.NewEncoder(w).Encode(&like)
	if err != nil {
		fmt.Println("err returning like as json: ", err)
	}
}

func (s *Server) handleDeleteLike(w http.ResponseWriter, r *http.Request) {
	var like domain.Like

	params, found := mux.Vars(r)["tweet_id"]
	if found {
		tweetId, err := strconv.Atoi(params)
		if err != nil {
			fmt.Println("err converting string route param tweet_id to golang int: ", err)
		}
		like.TweetID = tweetId
	}

	user := s.getUserFromContext(r.Context())
	like.UserID = user.ID

	err := s.ls.Delete(&like)
	if err != nil {
		fmt.Println("err deleting like: ", err)
	}

	err = json.NewEncoder(w).Encode(&like)
	if err != nil {
		fmt.Println("err returning deleted like as json: ", err)
	}
}