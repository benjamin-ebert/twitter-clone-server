package http

import (
	"crypto/rand"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"strconv"
	"wtfTwitter/crud"
	"wtfTwitter/domain"
)

// Server provides routing, request handling and middleware. It contains all the route
// declarations and their respective handler functions. It also performs authentication
// and authorization before calling one of the crud services to do some actual work.
type Server struct {
	isProd    bool
	clientUrl string
	router    *mux.Router
	github    *oauth2.Config
	// A single field for every service isn't necessary here, since the services could be
	// accessed through the passed in crud.Services object like so: s.service.User.Create(...).
	// However, having those single fields nicely shortens the call: s.us.Create(...).
	us domain.UserService
	os domain.OAuthService
	ts domain.TweetService
	fs domain.FollowService
	ls domain.LikeService
	is domain.ImageService
}

// NewServer returns a new instance of the server, registers all necessary
// routes and gives their handlers access to the crud services passed in.
func NewServer(
	isProd bool,
	clientUrl string,
	github *oauth2.Config,
	services *crud.Services,
) *Server {

	// Construct a new Server with a gorilla router and the services passed in.
	s := &Server{
		isProd:    isProd,
		clientUrl: clientUrl,
		router:    mux.NewRouter(),
		github:    github,
		us:        services.User,
		os:        services.OAuth,
		ts:        services.Tweet,
		fs:        services.Follow,
		ls:        services.Like,
		is:        services.Image,
	}

	r := s.router.PathPrefix("/api").Subrouter()

	// Register routes of the auth system.
	s.registerAuthRoutes(r)
	s.registerOAuthRoutes(r)

	// Register routes of the crud system.
	s.registerUserRoutes(r)
	s.registerTweetRoutes(r)
	s.registerFollowRoutes(r)
	s.registerLikeRoutes(r)
	s.registerImageRoutes(r)

	// Set up routes for serving images.
	imageHandler := http.FileServer(http.Dir("./images/"))
	s.router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Construct the CSRF protection middleware.
	// A new CSRF tokens is issued when the client makes a GET request to /csrf.
	csrfAuthKey := make([]byte, 32)
	if _, err := rand.Read(csrfAuthKey); err != nil {
		panic(err)
	}
	// TODO: Figure out SameSite for the remember token cookie...
	// ...since firefox warns it will be rejected if set to None like it is now.
	csrfMw := csrf.Protect(csrfAuthKey, csrf.Secure(s.isProd), csrf.Path("/"), csrf.SameSite(csrf.SameSiteStrictMode))

	// Set up middleware that needs to run on every request.
	s.router.Use(csrfMw, setContentTypeJSON, s.checkUser)

	// Return the pointer to the Server object.
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
	// 0.0.0.0 instead of localhost, so the proxy for the Angular SPA works properly.
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), s.router))
}
