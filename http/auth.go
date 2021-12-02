package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) registerAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", s.handleRegister).Methods("GET")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var m []Method
	standard := Method{Name: "Standard"}
	github := Method{Name: "Github"}
	m = append(m, standard)
	m = append(m, github)
	json.NewEncoder(w).Encode(m)
}

type Method struct {
	Name string
}