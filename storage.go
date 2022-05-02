package database

import (
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path"

	_ "image/gif"

	"github.com/corona10/goimagehash"
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

type DataStore struct {
	Root   string
	DB     Store
	Server http.Handler
}

type IData interface {
	Save(io.Reader, []string) error
}

func CreateStorage(f string, db Store) (IData, error) {
	err := os.MkdirAll(path.Join(f, "lg"), 0777)
	if err != nil {
		return nil, fmt.Errorf("createLgDir: %w", err)
	}
	err = os.MkdirAll(path.Join(f, "md"), 0777)
	if err != nil {
		return nil, fmt.Errorf("createMdDir: %w", err)
	}
	err = os.MkdirAll(path.Join(f, "sm"), 0777)
	if err != nil {
		return nil, fmt.Errorf("createSmDir: %w", err)
	}
	for _, dir := range []string{"lg", "md", "sm"} {
		err = os.MkdirAll(path.Join(f, dir), 0777)
		if err != nil {
			return nil, fmt.Errorf("create%sDir: %w", dir, err)
		}
	}
	storage := DataStore{
		Root: f,
		DB:   db,
	}
	return storage, nil
}

func TempImage(r io.Reader) (*Image, error) {
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("CreateTempFile: %w", err)
	}
	img, ext, err := image.Decode(io.TeeReader(r, tempFile))
	if err != nil {
		return nil, fmt.Errorf("WriteCreateImage: %w", err)
	}
	dhash, err := goimagehash.DifferenceHash(img)
	if err != nil {
		return nil, fmt.Errorf("CreateHash: %w", err)
	}
	image := &Image{
		Path: tempFile.Name(),
		Size: int64(img.Bounds().Dx() * img.Bounds().Dx()),
		Hash: int64(dhash.GetHash()),
		Ext:  ext,
		Data: img,
	}
	return image, nil
}

func (s DataStore) Save(r io.Reader, tags []string) error {
	img, err := TempImage(r)
	if err != nil {
		return fmt.Errorf("SaveImage: %w", err)
	}
	prevImg, err := s.DB.ByHash(img.Hash)
	if !(err == nil || errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("FindImageDB: %w", err)
	}
	if err == nil && prevImg.Size >= img.Size {
		os.Remove(img.Path)
		return nil
	}
	err = s.NewImage(img)
	if err != nil {
		return fmt.Errorf("ErrNewImage: %w", err)
	}
	err = s.DB.SaveImage(img)
	if err != nil {
		return fmt.Errorf("SaveImgDB: %w", err)
	}
	err = s.DB.SaveImage(img)
	if err != nil {
		return fmt.Errorf("SaveImgDB: %w", err)
	}

	err = s.DB.SaveTags(tags)
	if err != nil {
		return fmt.Errorf("CreateTagsDB: %w", err)
	}
	err = s.DB.SaveImageTags(img.Hash, tags)
	if err != nil {
		return fmt.Errorf("SaveTagsDB: %w", err)
	}

	return nil
}

func (s DataStore) NewImage(img *Image) error {
	filename := fmt.Sprintf("%x.%s", uint64(img.Hash), img.Ext)
	lgImg := path.Join(s.Root, "lg", filename)
	mdImg := path.Join(s.Root, "md", filename)
	smImg := path.Join(s.Root, "sm", filename)
	os.Rename(img.Path, lgImg)
	if img.Data.Bounds().Dx() > 1080 {
		reimg := imaging.Resize(img.Data, 1080, 0, imaging.Box)
		mdFile, _ := os.Create(mdImg)
		jpeg.Encode(mdFile, reimg, nil)
	} else {
		os.Symlink(lgImg, mdImg)
	}
	if img.Data.Bounds().Dx() > 720 {
		reimg := imaging.Resize(img.Data, 720, 0, imaging.Box)
		smFile, _ := os.Create(smImg)
		jpeg.Encode(smFile, reimg, nil)
	} else {
		os.Symlink(lgImg, smImg)
	}
	img.Path = filename
	return nil
}
