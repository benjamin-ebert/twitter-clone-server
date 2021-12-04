package http

import (
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"wtfTwitter/domain"
)

type Server struct {
	router *mux.Router

	AuthService domain.AuthService
	UserService domain.UserService
}

func NewServer() *Server {
	s := &Server{
		router: mux.NewRouter(),
	}
	{
		r := s.router.PathPrefix("/").Subrouter()
		s.registerAuthRoutes(r)
	}
	s.router.Use(setContentTypeJSON)
	return s
}

func setContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, `{"db": "not yet set up"`)
		if err != nil {
			return 
		}
		next.ServeHTTP(w, r)
	})
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	io.WriteString(w, `{"alive": true}`)
}

func (s *Server) Run(server *Server) {
	s.router.HandleFunc("/health", HealthCheckHandler)

	log.Fatal(http.ListenAndServe("localhost:1111", s.router))
}