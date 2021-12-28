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
	return &ImageService{}
}

type ImageService struct{}

func (is *ImageService) Create(img *domain.Image) error {
	err := runImageValFns(img,
		is.fileNameNotEmpty,
		is.extensionValid,
		is.contentTypeValid,
		is.contentTypeExtensionMatch,
		is.belowMaxSize,
		is.fileNameUnique,
	)
	if err != nil {
		return err
	}

	path, err := is.mkImagePath(img.OwnerType, img.OwnerID)
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

type imageValFn func (img *domain.Image) error
func runImageValFns(img *domain.Image, fns ...imageValFn) error {
	for _, fn := range fns {
		if err := fn(img); err != nil {
			return err
		}
	}
	return nil
}

func (is *ImageService) belowMaxSize(img *domain.Image) error {
	size, err := img.File.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if err = resetReaderPosition(img); err != nil {
		return err
	}
	if size > MaxUploadSize {
		return errs.MaxUploadSizeExceeded
	}
	return nil
}

func (is *ImageService) contentTypeValid(img *domain.Image) error {
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
		return errs.ContentTypeInvalid
	}
	img.ContentType = contentType
	return nil
}

func (is *ImageService) contentTypeExtensionMatch(img *domain.Image) error {
	cType := img.ContentType
	cType = strings.TrimPrefix(cType, "image/")
	ext := img.Extension
	ext = strings.TrimPrefix(ext, ".")
	if cType != ext {
		return errs.ContentTypeExtensionMismatch
	}
	return nil
}

func (is *ImageService) extensionValid(img *domain.Image) error {
	ext := filepath.Ext(img.Filename)
	ext = strings.ToLower(ext)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return errs.ExtensionInvalid
	}
	if ext == ".jpg" {
		ext = ".jpeg"
	}
	img.Extension = ext
	return nil
}

func (is *ImageService) fileNameNotEmpty(img *domain.Image) error {
	if img.Filename == "" {
		return errs.FilenameRequired
	}
	return nil
}

func (is *ImageService) fileNameUnique(img *domain.Image) error {
	timestamp := time.Now().UnixMicro()
	img.Filename = strconv.FormatInt(timestamp, 10) + img.Extension
	return nil
}

func resetReaderPosition(img *domain.Image) error {
	// reset reader position back to beginning of the file, so that it can be properly read again
	_, err := img.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}
