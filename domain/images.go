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
	OwnerType string
	OwnerID   int
	File multipart.File
	Filename string
	Extension string
	ContentType string
	Bytes []byte
	Size int64
}

type ImageService interface {
	Create(image *Image) error
	ByOwner(ownerType string, ownerID int) ([]Image, error)
	Delete(i *Image) error
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
