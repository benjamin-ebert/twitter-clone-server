package http

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

const providerGithub = "github"
const cookieOAuthState = "oauth_state"

// registerOAuthRoutes is a helper for registering all oauth routes.
func (s *Server) registerOAuthRoutes(r *mux.Router) {
	// Set a user up for authentication with Github.
	r.HandleFunc("/oauth/github/connect", s.handleOAuthGithub).Methods("GET")

	// Handle the user coming back from Github. Finish the oauth authentication process.
	r.HandleFunc("/oauth/github/callback", s.handleOAuthGithubCallback).Methods("GET")
}

// handleOAuthGithub handles the route "GET /oauth/github/connect".
// It creates a state token, gives it to the user via cookie and sends them over to
// Github, where they can authorize this app to access their Github account.
func (s *Server) handleOAuthGithub(w http.ResponseWriter, r *http.Request) {
	// Generate a CSRF token called "state".
	state := csrf.Token(r)

	// Put it into a cookie, so we can verify it's the same when the user comes back.
	cookie := http.Cookie{
		Name:     cookieOAuthState,
		Value:    state,
		HttpOnly: true,
	}

	// Set the cookie on the user.
	http.SetCookie(w, &cookie)

	// Build the url to redirect the user to the OAuth provider.
	url := s.github.AuthCodeURL(state)

	// Redirect the user to the provider.
	http.Redirect(w, r, url, http.StatusFound)
}

// handleOAuthGithubCallback handles the route "GET /oauth/github/callback".
// After authorization at Github, the user gets redirected here. Github will attach
// multiple parameters to this url. This method parses those parameters, verifies
// the state from the user's state cookie, creates an oauth object, determines what
// to do with it, and identifies an existing user or creates a new one in our database.
// On success, it signs them in through the regular auth system (remember token + cookie).
func (s *Server) handleOAuthGithubCallback(w http.ResponseWriter, r *http.Request) {
	// By now the user has been over at Github and authorized our app's access to
	// their Github account. Github then sends them back here, to the callback
	// route /oauth/github/callback, and attaches a bunch of url parameters to it.
	// Parse those url parameters first.
	if err := r.ParseForm(); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Extract the state parameter sent back by Github.
	state := r.FormValue("state")

	// Get the state from the user's cookie.
	cookie, err := r.Cookie(cookieOAuthState)

	// Verify that the state sent back by Github is the same I've set in the user's cookie.
	if err != nil {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Something went wrong."))
		return
	} else if cookie == nil || cookie.Value != state {
		errs.ReturnError(w, r, errs.Errorf(errs.EINVALID, "Invalid state provided."))
		return
	}

	// Delete the cookie.
	cookie.Value = ""
	cookie.Expires = time.Now()

	// Set the cookie on the user.
	http.SetCookie(w, cookie)

	// Grab the authorization code that Github has sent back as a url param.
	code := r.FormValue("code")

	// Exchange it for a token.
	token, err := s.github.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a new GitHub API client.
	client := github.NewClient(oauth2.NewClient(r.Context(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.AccessToken},
	)))

	// Fetch information about the user that's currently authenticated at Github.
	// Require that we at least receive a user ID from them.
	githubUser, _, err := client.Users.Get(r.Context(), "")
	if err != nil {
		errs.ReturnError(w, r, fmt.Errorf("cannot fetch github user: %s", err))
		return
	} else if githubUser.ID == nil {
		errs.ReturnError(w, r, fmt.Errorf("user ID not returned by GitHub, cannot authenticate user"))
		return
	}

	// If Github provides the user's name and email, store it to be able to
	// link together multiple OAuth providers in the future (Facebook, Google, etc).
	var name string
	if githubUser.Name != nil {
		name = *githubUser.Name
	} else if githubUser.Login != nil {
		name = *githubUser.Login
	}
	var email string
	if githubUser.Email != nil {
		email = *githubUser.Email
	}

	// Create an OAuth object with an associated user.
	oauth := &domain.OAuth{
		Provider:       providerGithub,
		ProviderUserID: strconv.FormatInt(*githubUser.ID, 10),
		TokenType:      token.TokenType,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		User: domain.User{
			Name:  name,
			Email: email,
		},
	}
	if !token.Expiry.IsZero() {
		oauth.Expiry = token.Expiry
	}

	// Take the oauth object, find or create an associated user for it,
	// and sign that user in through the regular auth system.
	if err := s.oauthSignIn(w, r, oauth); err != nil {
		errs.ReturnError(w, r, err)
		return
	}

	// Redirect them to their profile for now. That will change in the future but
	// for now it's helpful, since the profile will only show upon successful signIn.
	//http.Redirect(w, r, "/profile", http.StatusFound)
	http.Redirect(w, r, "http://localhost:4200", http.StatusFound)
}

// oauthSignIn takes in a pointer to an oauth object, finds or creates an associated
// user for it, and signs that user in through the regular auth system. How it works:
// 1. First check if an oauth record with that provider_user_id and that provider exists.
// If yes, we have a user in our database who has previously signed in with Github.
// Update their Github oauth record with the new token and sign the user in.
// 2. If such an oauth record doesn't exist, the user either didn't sign in with Github
// before, or they don't exist in our database at all. Look for a user with the email
// from the oauth object to find out.
// 3. If such a user exists, they haven't signed in with Github before. Create a new
// oauth record and associate it with the user. Then sign them in.
// 4. If such a user doesn't exit, this is a new account registration. Create a new
// user record and a new Github oauth record, and associate them with each other.
// Then sign the user in.
func (s *Server) oauthSignIn(w http.ResponseWriter, r *http.Request, oauth *domain.OAuth) error {
	// The user who will eventually be signed in with the oauth.
	var authedUser *domain.User

	// Check if there is an oauth record having that Provider and that ProviderUserID.
	existingOAuth, err := s.os.ByProviderUserId(oauth.Provider, oauth.ProviderUserID)

	// If yes, that means a user exists in our database, and they have previously signed in with Github.
	if existingOAuth != nil && err == nil {

		// Update the oauth record's token data.
		existingOAuth.AccessToken = oauth.AccessToken
		existingOAuth.TokenType = oauth.TokenType
		existingOAuth.RefreshToken = oauth.RefreshToken
		existingOAuth.Expiry = oauth.Expiry
		if err := s.os.Update(existingOAuth); err != nil {
			return err
		}

		// Get the user by ID, using the UserID of the existing oauth record.
		existingUser, err := s.us.ByID(existingOAuth.UserID)
		if err != nil {
			return err
		}

		// Set the found user to be the one that will be signed in.
		authedUser = existingUser

		// If there's no oauth record with that Provider and that ProviderUserID...
	} else if existingOAuth == nil && err == gorm.ErrRecordNotFound {

		// ...look for a user with the email address returned by Github.
		existingUser, err := s.us.ByEmail(oauth.User.Email)

		// If a user was found, that means they have previously signed in, but not with Github.
		if existingUser != nil && err == nil {

			// Attach the found user to the oauth object and create the oauth record in the database.
			// TODO: Implement oauth for another provider in order to test this properly.
			oauth.User = *existingUser
			oauth.UserID = existingUser.ID
			if err := s.os.Create(oauth); err != nil {
				return fmt.Errorf("cannot create oauth: %s", err)
			}

			// Set the found user to be the one that will be signed in.
			authedUser = existingUser

			// If looking for a user with that email returned an error...
		} else {

			// ...and the error is RecordNotFound, that means they are here for the first time.
			if err == gorm.ErrRecordNotFound {

				// Create a new user with the info from Github and NoPasswordNeeded (more on that below).
				oauth.User.NoPasswordNeeded = true
				if err := s.us.Create(&oauth.User); err != nil {
					return err
				}

				// Attach their ID to the oauth object. Create a new oauth record in the database.
				oauth.UserID = oauth.User.ID
				if err := s.os.Create(oauth); err != nil {
					return fmt.Errorf("cannot create oauth: %s", err)
				}

				// Set the newly created user to be the one that will be signed in.
				authedUser = &oauth.User

				// If looking for a user with that email returns any other error...
			} else {
				// ...something went wrong internally.
				return err
			}
		}

		// If looking for an oauth record with that Provider and ProviderUserID returns any other error...
	} else {
		// ...something went wrong internally.
		return err
	}

	// By now authedUser should hold an actual user from our database.
	// If yes, sign them in. If not, return an error EINVALID with a message.
	// Signing a user in requires updating the user's remember token in the database.
	// The UserService runs password validations before every user-create or update.
	// Because the user is being signed in with oauth, not with email and password,
	// their Password field will be empty. That would normally cause the password
	// validations to fail. Only the field NoPasswordNeeded set to true will make them pass.
	// That is done inside a validation that runs before the password validations.
	// It checks if there is an oauth record associated with the user, if no password
	// hash is supplied, and sets NoPasswordNeeded to true if that's the case.
	// The NoPasswordNeeded field is never stored in the user's database record.
	if authedUser != nil {
		err = s.signIn(w, r.Context(), authedUser)
		if err != nil {
			return err
		}
	} else {
		return errs.Errorf(errs.EINVALID, "Failed to sign you in with that method. Please try a different one.")
	}

	// Return the nil error upon successful signIn.
	return nil
}
