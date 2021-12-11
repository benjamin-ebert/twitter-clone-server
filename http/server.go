package http

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"wtfTwitter/database"
	"wtfTwitter/domain"
)

type Server struct {
	router *mux.Router
	us domain.UserService
	ts domain.TweetService
}

func NewServer(us *database.UserService, ts *database.TweetService) *Server {

	s := &Server{
		router: mux.NewRouter(),
		us: us,
		ts: ts,
	}

	authRouter := s.router.PathPrefix("/").Subrouter()
	s.registerAuthRoutes(authRouter)
	tweetRouter := s.router.PathPrefix("/tweet").Subrouter()
	s.registerTweetRoutes(tweetRouter)

	s.router.Use(setContentTypeJSON, s.authUser)
	return s
}

func setContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Run() {
	//log.Fatal(http.ListenAndServe("localhost:1111", s.userMw.Apply(s.router)))
	log.Fatal(http.ListenAndServe("localhost:1111", s.router))
}