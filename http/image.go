package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) registerImageRoutes(r *mux.Router) {
	r.HandleFunc("/images/upload/{owner_type}/{owner_id:[0-9]+}", s.requireAuth(s.handleUploadImage)).Methods("POST")
}

// do one route for every use case - avatar, header and []tweet ?
// one function for every route
// put those routes into http/user.go and http/tweet.go?
// each function then calls basic upload handler?
func (s *Server) handleUploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		fmt.Println("err parsing multipart form: ", err)
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			fmt.Println("err opening file: ", err)
		}
		defer file.Close()
		err = s.is.Create("tweet", 1, file, f.Filename)
		if err != nil {
			fmt.Println("err storing image: ", err)
		}
	}
	return
}