package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"wtfTwitter/domain"
)

func (s *Server) registerOAuthRoutes(r *mux.Router) {
	r.HandleFunc("/oauth/index", s.handleOAuthIndex).Methods("GET")
	r.HandleFunc("/oauth/login", s.handleOAuthLogin).Methods("POST")
}

func (s *Server) handleOAuthIndex(w http.ResponseWriter, r *http.Request) {
	var m []OAuthMethod
	standard := OAuthMethod{Name: "Standard"}
	github := OAuthMethod{Name: "Github"}
	m = append(m, standard)
	m = append(m, github)
	json.NewEncoder(w).Encode(m)
}

type OAuthMethod struct {
	Name string
}

func (s *Server) handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Errorf("err reading request data: %w", err)
	}
	var oauth domain.OAuth
	oauth.Source = "standard"
	err = json.Unmarshal(data, &oauth)
	if err != nil {
		fmt.Errorf("err unmarshaling data read from request: %w", err)
	}
	err = s.OAuthService.CreateOAuth(r.Context(), &oauth)
	if err != nil {
		fmt.Errorf("err creating oauth: %w", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusCreated)
}
