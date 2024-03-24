package models

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ImageUris struct {
	ID         string         `db:"id"`
	CardId     sql.NullString `db:"card_id"`
	CardFaceId sql.NullString `db:"card_face_id"`
	Small      string         `db:"small_uri"`
	Normal     string         `db:"normal_uri"`
	Large      string         `db:"large_uri"`
	Png        string         `db:"png_uri"`
	ArtCrop    string         `db:"art_crop_uri"`
	BorderCrop string         `db:"border_crop_uri"`
}

func (iu *ImageUris) Save(db *sqlx.DB) error {
	iu.ID = uuid.NewString()

	query := `
	INSERT INTO image_uris (id, 
		card_id, 
		card_face_id, 
		small_uri, 
		normal_uri, 
		large_uri, 
		png_uri, 
		art_crop_uri, 
		border_crop_uri)
	VALUES (:id, 
		:card_id, 
		:card_face_id, 
		:small_uri, 
		:normal_uri, 
		:large_uri, 
		:png_uri, 
		:art_crop_uri, 
		:border_crop_uri)
	`

	if _, err := db.NamedExec(query, iu); err != nil {
		return err
	}

	return nil
}
