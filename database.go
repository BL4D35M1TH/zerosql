package database

import (
	"database/sql"
	"image"
)

type SortOrder int

type ErrCode string
type Error struct {
	Code    ErrCode
	Message string
}

const (
	ASC          = SortOrder(1)
	DESC         = SortOrder(-1)
	ErrDuplicate = ErrCode("duplicate")
	ErrUnique    = ErrCode("unique")
	ErrNotFound  = ErrCode("notFound")
)

type Image struct {
	Hash int64
	Path string
	Size int64
	Ext  string
	Data image.Image
}

type ImageProvider interface {
	ByHash(int64) (*Image, error)
	ByTags(tags []string, offset, limit int64, order SortOrder) ([]Image, error)
}

type ImageConsumer interface {
	SaveTags([]string) error
	SaveImage(*Image) error
	SaveImageTags(int64, []string) error
}

type ImageStore struct {
	*sql.DB
}

type Store interface {
	ImageConsumer
	ImageProvider
}
