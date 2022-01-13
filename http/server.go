package http

import (
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"strconv"
	"wtfTwitter/crud"
	"wtfTwitter/domain"
)

// Server provides most of the http functionality of this app, namely routing,
// request handling, and middleware. It also performs authentication and
// authorization before handing things over to one of the database services.
type Server struct {
	router *mux.Router
	us domain.UserService
	ts domain.TweetService
	fs domain.FollowService
	ls domain.LikeService
	is domain.ImageService
	os domain.OAuthService
	github oauth2.Config
}

// NewServer returns a new instance of the server, registers all necessary
// routes and gives their handlers access to the app services passed in.
func NewServer(
	us *crud.UserService,
	ts *crud.TweetService,
	fs *crud.FollowService,
	ls *crud.LikeService,
	is *crud.ImageService,
	os *crud.OAuthService,
	github oauth2.Config,
	) *Server {

	// Construct a new Server with a gorilla router and the services passed in.
	s := &Server{
		router: mux.NewRouter(),
		us: us,
		ts: ts,
		fs: fs,
		ls: ls,
		is: is,
		os: os,
		github: github,
	}

	// Register routes of the auth system.
	s.registerAuthRoutes(s.router)
	s.registerOAuthRoutes(s.router)

	// Register routes of the crud system.
	s.registerTweetRoutes(s.router)
	s.registerFollowRoutes(s.router)
	s.registerLikeRoutes(s.router)
	s.registerImageRoutes(s.router)

	// Set up middleware that needs to run on every request.
	// TODO: Put the csrf auth key into config.
	// TODO: Set secure value depending on a config production variable.
	// Construct the CSRF protection middleware. A new CSRF tokens is issued when the client requests
	// /register or /login with a GET request (they visit the register- or the login-page).
	csrfMw := csrf.Protect([]byte("32-byte-long-auth-key"), csrf.Secure(false), csrf.Path("/"))
	s.router.Use(csrfMw, setContentTypeJSON, s.checkUser)
	return s
}

// The setContentTypeJSON middleware sets the content type to "application/json".
func setContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Run starts to listen and serve on the specified port.
func (s *Server) Run(port int) {
	log.Fatal(http.ListenAndServe("localhost:" + strconv.Itoa(port), s.router))
}