package domain

import (
	"fmt"
	"mime/multipart"
	"net/url"
)

const (
	OwnerTypeTweet = "tweet"
	OwnerTypeUser = "user"
)

type Image struct {
	URL string `json:"url"`
	OwnerType string `json:"-"`
	OwnerID   int `json:"-"`
	File multipart.File `json:"-"`
	Filename string `json:"-"`
	Extension string `json:"-"`
	ContentType string `json:"-"`
	Bytes []byte `json:"-"`
	Size int64 `json:"-"`
}

type ImageService interface {
	Create(image *Image) error
	ByOwner(ownerType string, ownerID int) ([]Image, error)
	Delete(i *Image) error
	DeleteAll(ownerType string, ownerID int) error
}

func (i *Image) Path() string {
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("images/%v/%v/%v", i.OwnerType, i.OwnerID, i.Filename)
}
