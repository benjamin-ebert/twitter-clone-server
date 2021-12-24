package domain

import (
	"fmt"
	"io"
	"net/url"
)

type Image struct {
	OwnerType string
	OwnerID   int
	Filename string
}

type ImageService interface {
	Create(ownerType string, ownerID int, r io.Reader, filename string) error
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
