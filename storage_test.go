package database_test

import (
	"embed"
	"path"
	"sanndy/database"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed test_assets/*
var assets embed.FS

func TestCreateStorage(t *testing.T) {
	assert := assert.New(t)
	db := Setup(t)
	tmpDir := t.TempDir()
	storage, err := database.CreateStorage(tmpDir, db)
	assert.Nil(err)
	assert.NotNil(storage)
	assert.DirExists(path.Join(tmpDir, "lg"))
	assert.DirExists(path.Join(tmpDir, "md"))
	assert.DirExists(path.Join(tmpDir, "sm"))
}

func TestTempImage(t *testing.T) {
	assert := assert.New(t)
	file, _ := assets.Open("test_assets/taiga1.JPG")
	img, err := database.TempImage(file)
	assert.Nil(err)
	assert.NotNil(img)
}

func TestSaveImage(t *testing.T) {
	assert := assert.New(t)
	file, _ := assets.Open("test_assets/taiga2.JPG")
	defer file.Close()
	storage := SetupStorage(t)
	err := storage.Save(file, []string{"old", "safe"})
	assert.Nil(err)
	file, _ = assets.Open("test_assets/taiga2.JPG")
	err = storage.Save(file, []string{"young", "safe"})
	assert.Nil(err)
}

func SetupStorage(tb testing.TB) database.IData {
	tb.Helper()
	tempDir := tb.TempDir()
	db := Setup(tb)
	storage, _ := database.CreateStorage(tempDir, db)
	return storage
}
