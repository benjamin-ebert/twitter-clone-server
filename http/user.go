package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"wtfTwitter/domain"
	"wtfTwitter/security"
)

func (s *Server) registerAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/login", s.handleLogin).Methods("POST")
	r.HandleFunc("/logout", s.handleLogout).Methods("POST")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Errorf("err parsing register data from request body: %w", err)
	}
	var user domain.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		fmt.Errorf("err unmarshalling register data into user obj: %w", err)
	}
	err = s.UserService.CreateUser(r.Context(), &user)
	if err != nil {
		fmt.Errorf("err creating new user: %w", err)
	}
	err = s.signIn(w, r.Context(), &user)
	if err != nil {
		fmt.Errorf("err signing in new user: %w", err)
	}
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		fmt.Errorf("err returning new user as json: %w", err)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

// signIn is used to sign the given user in via cookies
func (s *Server) signIn(w http.ResponseWriter, ctx context.Context, user *domain.User) error {
	if user.Remember == "" {
		token, err := security.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = s.UserService.UpdateUser(ctx, user)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	return nil
}

