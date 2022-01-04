package domain

import (
	"fmt"
	"mime/multipart"
	"net/url"
)

const (
	// OwnerTypeTweet expresses that an Image belongs to a Tweet.
	OwnerTypeTweet = "tweet"
	// OwnerTypeUser expresses that an Image belongs to a User.
	OwnerTypeUser = "user"
	// ImagesBaseDir determines the general storage location of uploaded images.
	ImagesBaseDir = "images"
	// MaxUploadSize determines the maximum filesize of an image to be uploaded.
	MaxUploadSize int64 = 5 << 20 // 5 Megabyte
)

// Image represents an image to be uploaded. Images are only stored as files in the filesystem
// and have no dedicated table in the database. Images always have a polymorphic one-to-many
// relationship with an owner. The owner is the entity that the Image belongs to. As of now,
// that's either a Tweet or a User, depending on the Image's OwnerType. The exact record that
// the Image belongs to is determined by the OwnerID. Since Images are not stored in the database,
// the relationship must be created (and resolved) through the location of the stored image file
// in the filesystem:
// An Image belonging to the User with ID 1 will be stored in: images/user/1/unique_name.jpeg.
// An Image belonging to the Tweet with ID 2 will be stored in: images/tweet/2/unique_name.png.
// URL contains the relative path to an image stored in the filesystem, starting in ImagesBaseDir.
// File contains the actual image file that will be stored in the filesystem.
type Image struct {
	URL string `json:"url"`
	OwnerType string `json:"-"`
	OwnerID   int `json:"-"`
	File multipart.File `json:"-"`
	Filename string `json:"-"`
	Extension string `json:"-"`
	ContentType string `json:"-"`
	Size int64 `json:"-"`
}

// ImageService is a set of methods to manipulate and work with the Image model and respective image files.
type ImageService interface {
	Create(image *Image) error
	ByOwner(ownerType string, ownerID int) ([]Image, error)
	Delete(i *Image) error
	DeleteAll(ownerType string, ownerID int) error
}

// Path returns the path of an image stored in the filesystem.
func (i *Image) Path() string {
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

// RelativePath returns the relative path to an image stored in the filesystem.
func (i *Image) RelativePath() string {
	return fmt.Sprintf("%v/%v/%v/%v", ImagesBaseDir, i.OwnerType, i.OwnerID, i.Filename)
}
