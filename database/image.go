package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"wtfTwitter/domain"
)

var _ domain.ImageService = &ImageService{}

func NewImageService() *ImageService {
	return &ImageService{}
}

type ImageService struct{}

// Create needs validation? size, extension, unique name, content type
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

func (is *ImageService) ByOwner(ownerType string, ownerID int) ([]domain.Image, error) {
	path := is.imagePath(ownerType, ownerID)
	imgStrings, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	ret := make([]domain.Image, len(imgStrings))
	for i := range ret {
		imgStrings[i] = strings.Replace(imgStrings[i], path, "", 1)
		ret[i] = domain.Image{
			OwnerType: ownerType,
			OwnerID: ownerID,
			Filename: imgStrings[i],
		}
	}
	return ret, nil
}

func (is *ImageService) Delete(i *domain.Image) error {
	return os.Remove(i.RelativePath())
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
