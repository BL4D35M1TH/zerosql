package database_test

import (
	"database/sql"
	"sanndy/database"
	"testing"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

//go:embed sample.sql
var sample string

func TestCreation(t *testing.T) {
	assert := assert.New(t)
	sqlDB, _ := sql.Open("sqlite3", ":memory:")
	t.Cleanup(func() { sqlDB.Close() })
	db, err := database.CreateStore(sqlDB)
	assert.Nil(err)
	assert.NotNil(db)
	db, err = database.CreateStore(sqlDB)
	assert.Nil(err)
	assert.NotNil(db)
}

func TestByHash(t *testing.T) {
	assert := assert.New(t)

	db := Setup(t)
	expected := database.Image{
		Hash: 12345,
		Path: "temp/png.jpeg",
		Size: 169,
	}
	image, err := db.ByHash(12345)
	assert.Nil(err)
	assert.Equal(expected, *image)
}

func TestByTags(t *testing.T) {
	assert := assert.New(t)
	db := Setup(t)
	expected := []int64{54321, 67890}
	images, err := db.ByTags([]string{"safe", "old"}, 0, 10, database.DESC)
	assert.Nil(err)
	assert.Len(images, 2)
	assert.Equal(expected[0], images[0].Hash)
	images, err = db.ByTags([]string{"safe", "old"}, 0, 1, database.DESC)
	assert.Nil(err)
	assert.Len(images, 1)
	images, err = db.ByTags([]string{"safe", "old"}, 1, 1, database.DESC)
	assert.Nil(err)
	assert.Len(images, 1)
	assert.Equal(expected[1], images[0].Hash)
	images, err = db.ByTags([]string{"safe", "old"}, 0, 1, database.ASC)
	assert.Nil(err)
	assert.Len(images, 1)
	assert.Equal(expected[0], images[0].Hash)
	images, err = db.ByTags([]string{}, 0, 10, database.ASC)
	assert.Nil(err)
	assert.Len(images, 3)
}

func TestCreateTag(t *testing.T) {
	assert := assert.New(t)
	db := Setup(t)
	err := db.SaveTags([]string{"boobs"})
	assert.Nil(err)
	err = db.SaveTags([]string{"boobs"})
	assert.Nil(err)
}

func TestCreateImage(t *testing.T) {
	assert := assert.New(t)
	db := Setup(t)
	img := &database.Image{
		Hash: 45678,
		Path: "temp/nevernude.jpeg",
		Size: 4269,
	}
	err := db.SaveImage(img)
	assert.Nil(err)
	err = db.SaveImage(img)
	assert.Nil(err)
}

func TestCreateImageTag(t *testing.T) {
	assert := assert.New(t)
	db := Setup(t)
	err := db.SaveImageTags(12345, []string{"old"})
	assert.Nil(err)
	err = db.SaveImageTags(12345, []string{"old"})
	assert.Nil(err)
}

func Setup(tb testing.TB) database.Store {
	tb.Helper()
	sqlDB, _ := sql.Open("sqlite3", ":memory:")
	db, _ := database.CreateStore(sqlDB)
	sqlDB.Exec(sample)
	return db
}
