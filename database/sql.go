package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
)

//go:embed schema.sql
var schema string

func CreateStore(db *sql.DB) (Store, error) {
	_, err := db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("createTable: %w", err)
	}
	imageDB := ImageStore{db}
	return imageDB, nil
}

func (db ImageStore) ByHash(h int64) (*Image, error) {
	img := &Image{}
	err := db.QueryRow("SELECT dhash, path, size FROM images WHERE dhash = ?", h).Scan(&img.Hash, &img.Path, &img.Size)
	if err != nil {
		return img, fmt.Errorf("findImageHash: %w", err)
	}
	return img, nil
}

func (db ImageStore) ByTags(tags []string, offset, limit int64, order SortOrder) ([]Image, error) {
	var sorter string
	if order == ASC {
		sorter = "ASC"
	} else {
		sorter = "DESC"
	}
	var placeholders []string
	var sTags []any
	for _, tag := range tags {
		placeholders = append(placeholders, "?")
		sTags = append(sTags, string(tag))
	}
	sTags = append(sTags, len(tags))
	sTags = append(sTags, limit)
	sTags = append(sTags, offset)

	placeholderExpr := strings.Join(placeholders, ",")
	tagLogic := `
		LEFT JOIN 
		image_tags AS it 
		ON i.dhash = it.dhash
		WHERE label IN (` + placeholderExpr + `)
		GROUP BY i.dhash
		HAVING COUNT(DISTINCT label) = ?`

	if len(tags) == 0 {
		tagLogic = ""
		sTags = []any{limit, offset}
	}

	rows, err := db.Query(`
		SELECT i.dhash, path, size 
		FROM images AS i `+tagLogic+`
		ORDER BY created_at `+sorter+`
		LIMIT ?
		OFFSET ?`, sTags...)
	if err != nil {
		return nil, fmt.Errorf("findImagesTags.ExecQuery: %w", err)
	}
	defer rows.Close()
	var images []Image
	for rows.Next() {
		var img Image
		rows.Scan(&img.Hash, &img.Path, &img.Size)
		images = append(images, img)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("findImagesTags.ScanRows: %w", err)
	}
	return images, nil
}

func (db ImageStore) SaveTags(tags []string) error {
	if tags == nil || len(tags) == 0 {
		return errors.New("no tags supplied")
	}
	var args []any
	var placeholders []string
	for _, tag := range tags {
		args = append(args, strings.ToLower(strings.TrimSpace(tag)))
		placeholders = append(placeholders, "(?)")
	}
	_, err := db.Exec(`
		INSERT INTO tags (label)
		VALUES `+strings.Join(placeholders, ",")+`
		ON CONFLICT (label) DO NOTHING`, args...)
	if err != nil {
		return fmt.Errorf("createTagFail: %w", err)
	}
	return nil
}
func (db ImageStore) SaveImage(img *Image) error {
	_, err := db.Exec(`
		INSERT INTO images (dhash, path, size) 
		VALUES (?, ?, ?)
		ON CONFLICT (dhash) DO UPDATE
		SET path = EXCLUDED.path, size = EXCLUDED.size`, img.Hash, img.Path, img.Size)
	if err != nil {
		return fmt.Errorf("createImageFail: %w", err)
	}
	return nil
}
func (db ImageStore) SaveImageTags(h int64, tags []string) error {
	if tags == nil || len(tags) == 0 {
		return errors.New("no tags supplied")
	}
	var args []any
	var placeholders []string
	for _, tag := range tags {
		args = append(args, h)
		args = append(args, strings.ToLower(strings.TrimSpace(tag)))
		placeholders = append(placeholders, "(?,?)")
	}
	_, err := db.Exec(`
		INSERT INTO image_tags (dhash, label) 
		VALUES `+strings.Join(placeholders, ",")+`
		ON CONFLICT DO NOTHING`, args...)
	if err != nil {
		return fmt.Errorf("createImageTagFail: %w", err)
	}
	return nil
}
