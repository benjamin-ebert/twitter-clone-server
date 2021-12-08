package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"time"
	"wtfTwitter/auth"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

func (s *Server) registerAuthRoutes(r *mux.Router) {
	r.HandleFunc("/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/login", s.handleLogin).Methods("POST")
	r.HandleFunc("/logout", s.handleLogout).Methods("POST")
	//r.HandleFunc("/profile", s.requireUserMw.ApplyFn(s.handleProfile)).Methods("GET")
	r.HandleFunc("/profile", s.requireAuth(s.handleProfile)).Methods("GET")
	r.HandleFunc("/home", s.handleHome).Methods("GET")
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
	err = s.us.CreateUser(r.Context(), &user)
	if err != nil {
		fmt.Println("err creating new user: ", err)
	}
	err = s.signIn(w, r.Context(), &user)
	if err != nil {
		fmt.Println("err signing in new user: ", err)
	}

	response := make(map[string]string)
	response["message"] = "successfully registered"
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		fmt.Println("err returning success message as json: ", err)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("err reading login data from request body: ", err)
	}
	var user domain.User
	json.Unmarshal(data, &user)
	authUser, err := s.authenticate(user.Email, user.Password)
	if err != nil {
		switch err {
		case errs.NotFound:
			fmt.Println("email doesn't exist in our database", err)
		default:
			fmt.Println("err authenticating", err)
		}
		return
	}
	err = s.signIn(w, r.Context(), authUser)
	if err != nil {
		fmt.Println("err signing in after authentication", err)
		return
	}
	fmt.Println("Worked like a charm")

	response := make(map[string]string)
	response["message"] = "successfully logged in"
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		fmt.Println("err returning success message as json: ", err)
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name: "remember_token",
		Value: "",
		Expires: time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	user := s.getUserFromContext(r.Context())
	token, _ := auth.MakeRememberToken()
	user.Remember = token
	err := s.us.UpdateUser(r.Context(), user)
	if err != nil {
		fmt.Println("err updating user with new remember token: ", err)
	}
	//http.Redirect(w, r, "/profile", http.StatusFound)
	response := make(map[string]string)
	response["message"] = "successfully logged out"
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		fmt.Println("err returning success message as json: ", err)
	}
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	fruits := make(map[string]int)
	fruits["Apples"] = 25
	fruits["Oranges"] = 10

	user := s.getUserFromContext(r.Context())

	json.NewEncoder(w).Encode(&user)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]string)
	response["message"] = "welcome home"
	json.NewEncoder(w).Encode(&response)
}

// signIn is used to sign the given user in via cookies
func (s *Server) signIn(w http.ResponseWriter, ctx context.Context, user *domain.User) error {
	if user.Remember == "" {
		token, err := auth.MakeRememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = s.us.UpdateUser(ctx, user)
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
	found, err := s.us.FindUserByEmail(email)
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

func (s *Server) authUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := s.us.FindUserByRemember(cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		//ctx = context.WithValue(ctx, "user", user)
		ctx = s.setUserInContext(ctx, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.getUserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) setUserInContext(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, "user", user)
}

func (s *Server) getUserFromContext(ctx context.Context) *domain.User {
	if temp := ctx.Value("user"); temp != nil {
		if user, ok := temp.(*domain.User); ok {
			return user
		}
	}
	return nil
}