package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) registerUserRoutes(r *mux.Router) {
	r.HandleFunc("/profile", s.requireAuth(s.handleProfile)).Methods("GET")
	r.HandleFunc("/user/{image_type}/upload", s.requireAuth(s.handleUploadUserImages)).Methods("POST")
}

// handleProfile really shouldn't be here, but in its own http/user.go file for showing and updating users
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	fruits := make(map[string]int)
	fruits["Apples"] = 25
	fruits["Oranges"] = 10

	// getUserFromContext uses ByRemember under the Hood
	// append images within ByRemember? but then it would get the images on every damn request
	// so maybe better get ByID, which is only called within handleProfile?
	// and then within us.ByID append the users images?
	u := s.getUserFromContext(r.Context())
	user, err := s.us.ByID(u.ID)
	if err != nil {
		fmt.Println("err retrieving user: ", err)
	}

	for i, t := range user.Tweets {
		images, err := s.is.ByOwner("tweet", t.ID)
		if err != nil {
			fmt.Println("err retrieving user tweet images: ", err)
		}
		user.Tweets[i].Images = images
	}

	json.NewEncoder(w).Encode(&user)
}

func (s *Server) handleUploadUserImages(w http.ResponseWriter, r *http.Request) {
	imgType, found := mux.Vars(r)["image_type"]
	if found {
		if imgType == "avatar" || imgType == "header" {
			user := s.getUserFromContext(r.Context())

			err := r.ParseMultipartForm(1 << 20)
			if err != nil {
				 fmt.Println("err parsing multipart form: ", err)
			}

			files := r.MultipartForm.File["image"]
			image, err := files[0].Open()
			if err != nil {
				 fmt.Println("err opening image: ", err)
			}
			defer image.Close()

			err = s.is.Create("user", user.ID, image, files[0].Filename)
			if err != nil {
				 fmt.Println("err storing image: ", err)
			}

			if imgType == "avatar" {
				user.Avatar = files[0].Filename
			} else {
				user.Header = files[0].Filename
			}

			err = s.us.UpdateUser(r.Context(), user)
			if err != nil {
				fmt.Println("err updating user image: ", err)
			}

			// delete the previous header / avatar after successful update
			// might as well just delete everything in the users dir except the two stored in the db
			userImgs, err := s.is.ByOwner("user", user.ID)
			for _, img := range userImgs {
				if img.Filename != user.Avatar && img.Filename != user.Header {
					s.is.Delete(&img)
				}
			}
		}
	}
	return
}
