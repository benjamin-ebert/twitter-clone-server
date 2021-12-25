package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) registerUserRoutes(r *mux.Router) {
	r.HandleFunc("/profile", s.requireAuth(s.handleProfile)).Methods("GET")
}

// handleProfile really shouldn't be here, but in its own http/user.go file for showing and updating users
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	fruits := make(map[string]int)
	fruits["Apples"] = 25
	fruits["Oranges"] = 10

	// getUserFromContext uses ByRemember under the Hood
	// append images within ByRemember? but then it would get the images on every damn request
	// so maybe better get ByID, which is only called within handleProfile?
	// and then within us.ByID append the users images?
	u := s.getUserFromContext(r.Context())
	user, err := s.us.ByID(u.ID)
	if err != nil {
		fmt.Println("err retrieving user: ", err)
	}

	for i, t := range user.Tweets {
		images, err := s.is.ByTweetID(t.ID)
		if err != nil {
			fmt.Println("err retrieving user tweet images: ", err)
		}
		user.Tweets[i].Images = images
	}

	json.NewEncoder(w).Encode(&user)
}
