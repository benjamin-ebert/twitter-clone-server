package crud

import (
	"fmt"
	"strconv"
	"time"

	//"image/png"
	"wtfTwitter/errs"

	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"wtfTwitter/domain"
)

const (
)

// ImageService manages Images.
// It implements the domain.ImageService interface.
type ImageService struct{
	imageValidator
}

// imageValidator runs validations on incoming Image data.
// On success, it passes the data on to imageCrud.
// Otherwise, it returns the error of the validation that has failed.
type imageValidator struct{
	imageCrud
}

// imageCrud runs CRUD operations on the filesystem using incoming Image data.
// It assumes that data has been validated. On success, it returns nil.
// Otherwise, it returns the error of the operation that has failed.
type imageCrud struct{}

// NewImageService returns an instance of ImageService.
func NewImageService() *ImageService {
	return &ImageService{
		imageValidator{
			imageCrud{},
		},
	}
}

// Ensure the ImageService struct properly implements the domain.ImageService interface.
// If it does not, then this expression becomes invalid and won't compile.
var _ domain.ImageService = &ImageService{}

// Create runs validations needed for storing uploaded images in the filesystem.
func (iv *imageValidator) Create(img *domain.Image) error {
	err := runImageValFns(img,
		iv.extensionValid,
		iv.contentTypeValid,
		iv.contentTypeExtensionMatch,
		iv.belowMaxSize,
		iv.fileNameUnique,
	)
	if err != nil {
		return err
	}
	return iv.imageCrud.Create(img)
}

// runImageValFns runs any number of functions of type imageValFn on the passed in Image object.
func runImageValFns(img *domain.Image, fns ...imageValFn) error {
	for _, fn := range fns {
		if err := fn(img); err != nil {
			return err
		}
	}
	return nil
}

// A imageValFn is any function that takes in a pointer to a domain.Image object and returns an error.
type imageValFn func (img *domain.Image) error

// belowMaxSize makes sure that the image to be uploaded does not exceed MaxUploadSize.
func (iv *imageValidator) belowMaxSize(img *domain.Image) error {
	size, err := img.File.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if err = resetFilePointer(img); err != nil {
		return err
	}
	if size > domain.MaxUploadSize {
		return errs.Errorf(
			errs.EINVALID,
			"Image " + img.Filename + " exceeds upload size limit of " + strconv.FormatInt(domain.MaxUploadSize/1000000, 10) + "MB.",
			)
	}
	return nil
}

// contentTypeValid makes sure that the image to be uploaded is a valid jpeg or png file.
func (iv *imageValidator) contentTypeValid(img *domain.Image) error {
	buffer := make([]byte, 512)
	_, err := img.File.Read(buffer)
	if err != nil {
		return err
	}
	if err = resetFilePointer(img); err != nil {
		return err
	}
	contentType := http.DetectContentType(buffer)
	if contentType != "image/jpeg" && contentType != "image/png" {
		return errs.Errorf(
			errs.EINVALID,
			"Image " + img.Filename + " invalid content-type, must be image/jpeg or image/png.",
			)
	}
	img.ContentType = contentType
	return nil
}

// contentTypeExtensionMatch makes sure that the image's filename extension and content type match.
func (iv *imageValidator) contentTypeExtensionMatch(img *domain.Image) error {
	contentType := strings.TrimPrefix(img.ContentType, "image/")
	ext := strings.TrimPrefix(img.Extension, ".")
	if contentType != ext {
		return errs.Errorf(
			errs.EINVALID,
			"Image " + img.Filename + " content-type " + img.ContentType + " does not match extension " + img.Extension + ".",
			)
	}
	return nil
}

// extensionValid makes sure that the image to be uploaded has the extension .jpeg,
// .jpg or .png. If the extension is .jpg it will be renamed to .jpeg for consistency.
func (iv *imageValidator) extensionValid(img *domain.Image) error {
	ext := filepath.Ext(img.Filename)
	ext = strings.ToLower(ext)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return errs.Errorf(
			errs.EINVALID,
			"Image " + img.Filename + " invalid extension, must be .jpeg or .png",
			)
	}
	if ext == ".jpg" {
		ext = ".jpeg"
	}
	img.Extension = ext
	return nil
}

// fileNameUnique replaces the image's name with a unique string based on a unix timestamp.
func (iv *imageValidator) fileNameUnique(img *domain.Image) error {
	timestamp := time.Now().UnixMicro()
	img.Filename = strconv.FormatInt(timestamp, 10) + img.Extension
	return nil
}

// resetFilePointer sets the file pointer back to beginning of the file,
// so that subsequent reads can properly read from the beginning again.
func resetFilePointer(img *domain.Image) error {
	_, err := img.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}

// Create takes a domain.Image object, creates a path to store the image, creates a
// destination file inside that path, and copies the file data from the domain.Image object
// into the destination file. If the path already exists, that one will be used.
// Images are only stored in the filesystem and have no dedicated table in the database.
// The users table does have columns to store the paths to a user's avatar and header image though.
func (ic *imageCrud) Create(img *domain.Image) error {
	path, err := ic.mkImagePath(img.OwnerType, img.OwnerID)
	if err != nil {
		return err
	}
	// Create a destination file inside the path.
	dst, err := os.Create(path + img.Filename)
	if err != nil {
		return err
	}
	// Copy the file data from the domain.Image object into the destination path.
	_, err = io.Copy(dst, img.File)
	if err != nil {
		return err
	}
	return nil
}

// ByOwner takes an ownerType, which as of now is either an Image or a User, and an ownerID.
// It returns an array of domain.Image objects containing information about that owner's images.
func (ic *imageCrud) ByOwner(ownerType string, ownerID int) ([]domain.Image, error) {
	path := ic.imagePath(ownerType, ownerID)
	imgStrings, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	ret := make([]domain.Image, len(imgStrings))
	for i := range ret {
		imgStrings[i] = strings.Replace(imgStrings[i], path, "", 1)
		ret[i] = domain.Image{
			Filename: imgStrings[i],
			OwnerType: ownerType,
			OwnerID: ownerID,
			URL: path + imgStrings[i],
		}
	}
	return ret, nil
}

// Delete removes a specific image from the filesystem.
func (ic *imageCrud) Delete(i *domain.Image) error {
	return os.Remove(i.RelativePath())
}

// DeleteAll removes an entire directory containing images from the filesystem.
func (ic *imageCrud) DeleteAll(ownerType string, ownerID int) error {
	return os.RemoveAll(ic.imagePath(ownerType, ownerID))
}

// mkImagePath creates a filesystem path based on an image's ownerType and ownerID.
// This results in directories like: /images/user/1/ and /images/tweet/2/.
func (ic *imageCrud) mkImagePath(ownerType string, ownerID int) (string, error) {
	imagePath := ic.imagePath(ownerType, ownerID)
	err := os.MkdirAll(imagePath, 0755)
	if err != nil {
		return "", err
	}
	return imagePath, nil
}

// imagePath builds the name of a path based on the name of the base directory for images,
// an image's ownerType and its ownerID.
func (ic *imageCrud) imagePath(ownerType string, ownerID int) string {
	return fmt.Sprintf("%v/%v/%v/", domain.ImagesBaseDir, ownerType, ownerID)
}
