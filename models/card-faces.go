package models

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type CardFace struct {
	ID             string         `db:"id"`
	CardId         string         `db:"card_id"`
	Name           string         `db:"card_name"`
	ManaCost       string         `db:"mana_cost"`
	TypeLine       string         `db:"type_line"`
	PrintedText    string         `db:"printed_text"`
	FlavorText     string         `db:"flavor_text"`
	Colors         pq.StringArray `db:"colors"`
	ColorIndicator pq.StringArray `db:"color_indicator"`
	ImageUris      *ImageUris     `db:"-"`
}

func (cf *CardFace) Save(db *sqlx.DB) error {
	cf.ID = uuid.NewString()
	cf.ImageUris.CardFaceId = sql.NullString{String: cf.ID, Valid: true}

	query := `
	INSERT INTO card_faces (id, 
		card_id,
		card_name,
		mana_cost,
		type_line,
		printed_text,
		flavor_text,
		colors,
		color_indicator)
	VALUES (:id, 
		:card_id, 
		:card_name, 
		:mana_cost, 
		:type_line, 
		:printed_text, 
		:flavor_text, 
		:colors, 
		:color_indicator)
	ON CONFLICT DO NOTHING
	`

	if _, err := db.NamedExec(query, cf); err != nil {
		return err
	}

	return cf.ImageUris.Save(db)
}
