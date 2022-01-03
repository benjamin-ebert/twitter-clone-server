package http

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"wtfTwitter/database"
	"wtfTwitter/domain"
	"wtfTwitter/storage"
)

type Server struct {
	router *mux.Router
	us domain.UserService
	ts domain.TweetService
	fs domain.FollowService
	ls domain.LikeService
	is domain.ImageService
}

func NewServer(
	us *database.UserService,
	ts *database.TweetService,
	fs *database.FollowService,
	ls *database.LikeService,
	is *storage.ImageService,
	) *Server {

	s := &Server{
		router: mux.NewRouter(),
		us: us,
		ts: ts,
		fs: fs,
		ls: ls,
		is: is,
	}

	s.registerAuthRoutes(s.router)
	s.registerUserRoutes(s.router)
	//tweetRouter := s.router.PathPrefix("/tweet").Subrouter()
	//s.registerTweetRoutes(tweetRouter)
	s.registerTweetRoutes(s.router)
	s.registerLikeRoutes(s.router)
	s.registerFollowRoutes(s.router)

	s.router.Use(setContentTypeJSON, s.checkUser)
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