package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"wtfTwitter/domain"
)

func (s *Server) registerTweetRoutes(r *mux.Router) {
	r.HandleFunc("/create", s.requireAuth(s.handleCreateTweet)).Methods("POST")
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
	err = s.ts.CreateTweet(&tweet)
	if err != nil {
		fmt.Println("err creating new tweet: ", err)
	}

	err = json.NewEncoder(w).Encode(&tweet)
	if err != nil {
		fmt.Println("err returning tweet as json: ", err)
	}
}