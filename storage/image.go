package storage

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

const MaxUploadSize int64 = 5 << 20 // 5 Megabyte

var _ domain.ImageService = &ImageService{}

func NewImageService() *ImageService {
	return &ImageService{
		imageValidator{
			imageCrud{},
		},
	}
}

type ImageService struct{
	imageValidator
}

type imageValidator struct{
	imageCrud
}

type imageCrud struct{}

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

type imageValFn func (img *domain.Image) error
func runImageValFns(img *domain.Image, fns ...imageValFn) error {
	for _, fn := range fns {
		if err := fn(img); err != nil {
			return err
		}
	}
	return nil
}

func (iv *imageValidator) belowMaxSize(img *domain.Image) error {
	size, err := img.File.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if err = resetReaderPosition(img); err != nil {
		return err
	}
	if size > MaxUploadSize {
		return errs.Errorf(
			errs.EINVALID,
			"Image " + img.Filename + " exceeds upload size limit of " + strconv.FormatInt(MaxUploadSize/1000000, 10) + "MB.",
			)
	}
	return nil
}

func (iv *imageValidator) contentTypeValid(img *domain.Image) error {
	buffer := make([]byte, 512)
	_, err := img.File.Read(buffer)
	if err != nil {
		return err
	}
	if err = resetReaderPosition(img); err != nil {
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

func (iv *imageValidator) fileNameUnique(img *domain.Image) error {
	timestamp := time.Now().UnixMicro()
	img.Filename = strconv.FormatInt(timestamp, 10) + img.Extension
	return nil
}

// resetReaderPosition back to beginning of the file, so that subsequent reads will work.
func resetReaderPosition(img *domain.Image) error {
	_, err := img.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}

func (ic *imageCrud) Create(img *domain.Image) error {
	path, err := ic.mkImagePath(img.OwnerType, img.OwnerID)
	if err != nil {
		return err
	}
	dst, err := os.Create(path + img.Filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, img.File)
	if err != nil {
		return err
	}
	return nil
}

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
			URL: path + imgStrings[i],
		}
	}
	return ret, nil
}

func (ic *imageCrud) Delete(i *domain.Image) error {
	return os.Remove(i.RelativePath())
}

func (ic *imageCrud) mkImagePath(ownerType string, ownerID int) (string, error) {
	imagePath := ic.imagePath(ownerType, ownerID)
	err := os.MkdirAll(imagePath, 0755)
	if err != nil {
		return "", err
	}
	return imagePath, nil
}

func (ic *imageCrud) imagePath(ownerType string, ownerID int) string {
	return fmt.Sprintf("images/%v/%v/", ownerType, ownerID)
}
