package database

import (
	"fmt"
	"io"
	"os"
	"wtfTwitter/domain"
)

var _ domain.ImageService = &ImageService{}

func NewImageService() *ImageService {
	return &ImageService{}
}

type ImageService struct{}

func (is *ImageService) Create(ownerType string, ownerID int, r io.Reader, filename string) error {
	path, err := is.mkImagePath(ownerType, ownerID)
	if err != nil {
		return err
	}
	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (is *ImageService) mkImagePath(ownerType string, ownerID int) (string, error) {
	imagePath := is.imagePath(ownerType, ownerID)
	err := os.MkdirAll(imagePath, 0755)
	if err != nil {
		return "", err
	}
	return imagePath, nil
}

func (is *ImageService) imagePath(ownerType string, ownerID int) string {
	return fmt.Sprintf("images/%v/%v/", ownerType, ownerID)
}
