package http

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"wtfTwitter/database"
	"wtfTwitter/domain"
	"wtfTwitter/storage"
)

// Server provides most of the http functionality of this app, namely routing,
// request handling, and middleware. It also performs authentication and
// authorization before handing things over to one of the database services.
type Server struct {
	port int
	router *mux.Router
	us domain.UserService
	ts domain.TweetService
	fs domain.FollowService
	ls domain.LikeService
	is domain.ImageService
}

// NewServer returns a new instance of the server.
func NewServer(
	us *database.UserService,
	ts *database.TweetService,
	fs *database.FollowService,
	ls *database.LikeService,
	is *storage.ImageService,
	) *Server {

	// Construct a new Server with a gorilla router and the services passed in.
	s := &Server{
		router: mux.NewRouter(),
		us: us,
		ts: ts,
		fs: fs,
		ls: ls,
		is: is,
	}

	// Register all routes.
	s.registerAuthRoutes(s.router)
	s.registerUserRoutes(s.router)
	s.registerTweetRoutes(s.router)
	s.registerLikeRoutes(s.router)
	s.registerFollowRoutes(s.router)

	// Set up middleware that needs to run on every request.
	s.router.Use(setContentTypeJSON, s.checkUser)
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