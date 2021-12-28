package http

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"wtfTwitter/domain"
	"wtfTwitter/storage"
)

func (s *Server) registerUserRoutes(r *mux.Router) {
	r.HandleFunc("/profile", s.requireAuth(s.handleProfile)).Methods("GET")
	r.HandleFunc("/user/{image_type}/upload", s.requireAuth(s.handleUploadUserImages)).Methods("POST")
	r.HandleFunc("/user/{image_type}/delete", s.requireAuth(s.handleDeleteUserImages)).Methods("DELETE")
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	fruits := make(map[string]int)
	fruits["Apples"] = 25
	fruits["Oranges"] = 10

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

			err := r.ParseMultipartForm(storage.MaxUploadSize)
			if err != nil {
				 fmt.Println("err parsing multipart form: ", err)
			}

			files := r.MultipartForm.File["image"]
			image, err := files[0].Open()
			if err != nil {
				 fmt.Println("err opening image: ", err)
			}
			defer image.Close()

			var img domain.Image
			img.OwnerType = "user"
			img.OwnerID = user.ID
			img.File = image
			img.Filename = files[0].Filename
			err = s.is.Create(&img)
			if err != nil {
				 fmt.Println("err storing image: ", err)
			}

			if imgType == "avatar" {
				user.Avatar = img.Filename
			} else {
				user.Header = img.Filename
			}

			err = s.us.UpdateUser(r.Context(), user)
			if err != nil {
				fmt.Println("err updating user image: ", err)
			}

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

func (s *Server) handleDeleteUserImages(w http.ResponseWriter, r *http.Request) {
	imgType, found := mux.Vars(r)["image_type"]
	// This doesn't look right. Make it cleaner.
	if found {
		if imgType == "avatar" || imgType == "header" {
			user := s.getUserFromContext(r.Context())

			var filename string
			if imgType == "avatar" {
				filename = user.Avatar
			} else {
				filename = user.Header
			}

			var img domain.Image
			img.OwnerType = "user"
			img.OwnerID = user.ID
			img.Filename = filename

			err := s.is.Delete(&img)
			if err != nil {
				fmt.Println("err deleting image: ", err)
			}

			if imgType == "avatar" {
				user.Avatar = ""
			} else {
				user.Header = ""
			}

			err = s.us.UpdateUser(r.Context(), user)
			if err != nil {
				fmt.Println("err updating user image: ", err)
			}
		}
	}
	return
}
