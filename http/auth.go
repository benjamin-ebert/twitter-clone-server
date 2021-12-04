package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"wtfTwitter/domain"
)

func (s *Server) registerAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", s.handleRegisterIndex).Methods("GET")
	r.HandleFunc("/register", s.handleRegister).Methods("POST")
}

func (s *Server) handleRegisterIndex(w http.ResponseWriter, r *http.Request) {
	var m []AuthMethod
	standard := AuthMethod{Name: "Standard"}
	github := AuthMethod{Name: "Github"}
	m = append(m, standard)
	m = append(m, github)
	json.NewEncoder(w).Encode(m)
}

type AuthMethod struct {
	Name string
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Errorf("err reading request data: %w", err)
	}
	var auth domain.Auth
	auth.Source = "standard"
	err = json.Unmarshal(data, &auth)
	if err != nil {
		fmt.Errorf("err unmarshaling data read from request: %w", err)
	}
	err = s.AuthService.CreateAuth(r.Context(), &auth)
	if err != nil {
		fmt.Errorf("err creating auth: %w", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusCreated)
}
