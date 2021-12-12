package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"wtfTwitter/domain"
)

func (s *Server) registerFollowRoutes(r *mux.Router) {
	r.HandleFunc("/follow/{followed_id}", s.requireAuth(s.handleCreateFollow)).Methods("POST")
	r.HandleFunc("/unfollow/{followed_id}", s.requireAuth(s.handleDeleteFollow)).Methods("DELETE")
}

func (s *Server) handleCreateFollow(w http.ResponseWriter, r *http.Request) {
	var follow domain.Follow

	routeParams := mux.Vars(r)
	followedId, err := strconv.Atoi(routeParams["followed_id"])
	if err != nil {
		fmt.Println("err converting string route param followed_id to golang int: ", err)
	}
	follow.FollowedID = followedId

	follower := s.getUserFromContext(r.Context())
	follow.FollowerID = follower.ID

	err = s.fs.Create(&follow)
	if err != nil {
		fmt.Println("err creating new follow: ", err)
	}

	err = json.NewEncoder(w).Encode(&follow)
	if err != nil {
		fmt.Println("err returning follow as json: ", err)
	}
}

func (s *Server) handleDeleteFollow(w http.ResponseWriter, r *http.Request) {
	var follow domain.Follow

	routeParams := mux.Vars(r)
	followedId, err := strconv.Atoi(routeParams["followed_id"])
	if err != nil {
		fmt.Println("err converting string route param followed_id to golang int: ", err)
	}
	follow.FollowedID = followedId

	follower := s.getUserFromContext(r.Context())
	follow.FollowerID = follower.ID

	err = s.fs.Delete(&follow)
	if err != nil {
		fmt.Println("err deleting follow: ", err)
	}

	err = json.NewEncoder(w).Encode(&follow)
	if err != nil {
		fmt.Println("err returning follow as json: ", err)
	}
}