package http

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"wtfTwitter/auth"
	"wtfTwitter/database"
	"wtfTwitter/domain"
)

type Server struct {
	router *mux.Router

	OAuthService domain.OAuthService
	UserService domain.UserService
	userMw auth.User
	reqUserMw auth.RequireUser
}

func NewServer(us *database.UserService) *Server {
	s := &Server{
		router: mux.NewRouter(),
	}
	{
		r := s.router.PathPrefix("/").Subrouter()
		s.registerAuthRoutes(r)
		s.registerOAuthRoutes(r)
	}
	s.UserService = us
	userMw := auth.User{
		UserService: s.UserService,
	}
	reqUserMw := auth.RequireUser{
		User: userMw,
	}
	s.userMw = userMw
	s.reqUserMw = reqUserMw

	s.router.Use(setContentTypeJSON)
	return s
}

func setContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Run(server *Server) {
	log.Fatal(http.ListenAndServe("localhost:1111", s.userMw.Apply(s.router)))
}