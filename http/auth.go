package http

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

const (
	ctxUserKey = "user"
)

// registerAuthRoutes is a helper for registering all authentication routes.
func (s *Server) registerAuthRoutes(r *mux.Router) {
	// Register a new user.
	r.HandleFunc("/register", s.handleRegister).Methods("POST")

	// Login an existing user.
	r.HandleFunc("/login", s.handleLogin).Methods("POST")

	// Logout a logged-in user.
	r.HandleFunc("/logout", s.requireAuth(s.handleLogout)).Methods("POST")

	// Display the home / login page.
	r.HandleFunc("/home", s.handleHome).Methods("GET")
}

// handleRegister creates a new user record in the database and signs the user in.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	// Parse the request's json body into a User object.
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid json body."))
		return
	}

	// Create a new user record in the database.
	err := s.us.Create(r.Context(), &user)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Sign the new user in (through a remember token and a cookie).
	err = s.signIn(w, r.Context(), &user)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the created user.
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(&user); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleLogin authenticates a user and signs them in on success.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse the request's json body (email and password) into a User object.
	var user domain.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid json body."))
		return
	}

	// Authenticate the user.
	authedUser, err := s.us.Authenticate(user.Email, user.Password)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Sign the user in (through a remember token and a cookie).
	err = s.signIn(w, r.Context(), authedUser)
	if err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Return the logged-in user.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&authedUser); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleLogout logs a user out by updating their remember token and invalidating their cookie.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get the authed user from the request's context.
	user := s.getUserFromContext(r.Context())

	// Create a new remember token and replace the user's current one with it.
	token, _ := s.us.MakeRememberToken()
	user.Remember = token

	// Update the user's record in the database.
	if err := s.us.Update(r.Context(), user); err != nil {
		errs.ReturnError(w, r, err)
		return
	}
	
	// Create a new http.Cookie that has an empty remember_token and expires immediately.
	cookie := http.Cookie{
		Name: "remember_token",
		Value: "",
		Expires: time.Now(),
		HttpOnly: true,
	}

	// Add the new cookie to the response.
	http.SetCookie(w, &cookie)

	// Return a success message.
	response := make(map[string]string)
	response["message"] = "Successfully logged out."
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		errs.LogError(r, err)
		return
	}
}

// handleHome displays the home / login page.
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]string)
	response["message"] = "Welcome home."
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		errs.LogError(r, err)
		return
	}
}

// signIn signs a given user in through a cookie and a remember token.
func (s *Server) signIn(w http.ResponseWriter, ctx context.Context, user *domain.User) error {
	// If the user doesn't have a remember token, create a new one, 
	// assign it to them and update their database record.
	if user.Remember == "" {
		token, err := s.us.MakeRememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		err = s.us.Update(ctx, user)
		if err != nil {
			return err
		}
	}

	// Create a new http.Cookie containing the user's (updated) remember token.
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}
	
	// Add the cookie to the response. From now the remember token is passed back and forth via cookie.
	http.SetCookie(w, &cookie)

	// Clear the remember token from the user object in memory for security reasons.
	// Only the remember token in the cookie and the hashed remember token in the database are left.
	user.Remember = ""

	return nil
}


// The checkUser middleware reads an incoming request's cookie, checks if its remember token
// matches a user database record, and on success attaches that user to the request context.
// Subsequent request handlers can read the current user from the request's context. If the
// cookie's remember token did not match a user record, the request's context does not change.
// checkUser always returns the next request handler (usually that's the requireAuth middleware).
func (s *Server) checkUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request's cookie named remember_token.
		cookie, err := r.Cookie("remember_token")
		// If the cookie can't be read / does not exist, return the subsequent request handler.
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Look for a user database record matching the cookie's remember token value.
		user, err := s.us.ByRemember(cookie.Value)
		// If such a record does not exist, return the subsequent request handler.
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Create a context.Context for the request.
		ctx := r.Context()

		// Put the found user into the request's context.
		ctx = s.setUserInContext(ctx, user)

		// Attach the context to the request.
		r = r.WithContext(ctx)

		// Return subsequent request handler.
		next.ServeHTTP(w, r)
	})
}

// The requireAuth middleware prevents unauthenticated users from accessing things
// that require authentication. It does that by trying to read the authenticated user
// from the request's context. If it fails, it redirects to the home / login page.
// Otherwise, it returns the subsequent authed-users-only handler.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the authed user from the request's context.
		if user := s.getUserFromContext(r.Context()); user == nil {
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		// Return the subsequent request handler.
		next.ServeHTTP(w, r)
	}
}

// setUserInContext takes a context and a user and puts the user into the context.
func (s *Server) setUserInContext(ctx context.Context, user *domain.User) context.Context {
	// Within the context set a user-key that other functions can access
	// to get the user-value. Set the user as the context's user-value.
	return context.WithValue(ctx, ctxUserKey, user)
}

// getUserFromContext takes a context, reads a user from it, and returns the user.
func (s *Server) getUserFromContext(ctx context.Context) *domain.User {
	// Try to read the value of the context's user-key into temp.
	if temp := ctx.Value(ctxUserKey); temp != nil {
		// Assert that the type of the value stored in temp is identical to type *domain.User.
		if user, ok := temp.(*domain.User); ok {
			return user
		}
	}
	return nil
}