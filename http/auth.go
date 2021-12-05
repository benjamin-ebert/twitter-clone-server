package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"wtfTwitter/auth"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

func (s *Server) registerAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/login", s.handleLogin).Methods("POST")
	r.HandleFunc("/logout", s.handleLogout).Methods("POST")
	r.HandleFunc("/profile", s.reqUserMw.ApplyFn(s.handleProfile)).Methods("GET")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("err reading register data from request body: ", err)
	}
	var user domain.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		fmt.Println("err unmarshalling register data into user obj: ", err)
	}
	err = s.UserService.CreateUser(r.Context(), &user)
	if err != nil {
		fmt.Println("err creating new user: ", err)
	}
	err = s.signIn(w, r.Context(), &user)
	if err != nil {
		fmt.Println("err signing in new user: ", err)
	}
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		fmt.Println("err returning new user as json: ", err)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("err reading login data from request body: ", err)
	}
	var cred domain.User
	json.Unmarshal(data, &cred)
	user, err := s.authenticate(cred.Email, cred.Password)
	if err != nil {
		switch err {
		case errs.NotFound:
			fmt.Println("email doesn't exist in our database", err)
		default:
			fmt.Println("err authenticating", err)
		}
		return
	}
	err = s.signIn(w, r.Context(), user)
	if err != nil {
		fmt.Println("err signing in after authentication", err)
		return
	}
	fmt.Println("Worked like a charm")
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	fruits := make(map[string]int)
	fruits["Apples"] = 25
	fruits["Oranges"] = 10

	json.NewEncoder(w).Encode(&fruits)
}

// signIn is used to sign the given user in via cookies
func (s *Server) signIn(w http.ResponseWriter, ctx context.Context, user *domain.User) error {
	if user.Remember == "" {
		token, err := auth.RememberToken()
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
	fmt.Println("COOKIE COOKIE COOKIE: ", cookie)
	return nil
}

func (s *Server) authenticate(email, password string) (*domain.User, error) {
	found, err := s.UserService.FindUserByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(found.PasswordHash), []byte(password + "blerz"))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, errs.PasswordIncorrect
		default:
			return nil, err
		}
	}
	return found, nil
}
