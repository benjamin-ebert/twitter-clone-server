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
	//userMw auth.UserMw
	//requireUserMw auth.RequireUserMw
}

func NewServer(us *database.UserService) *Server {
	//userMw := auth.UserMw{ UserService: us }
	//requireUserMw := auth.RequireUserMw{ UserMw: userMw }

	s := &Server{
		router: mux.NewRouter(),
		us: us,
		//userMw: userMw,
		//requireUserMw: requireUserMw,
	}

	{
		r := s.router.PathPrefix("/").Subrouter()
		s.registerAuthRoutes(r)
	}

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