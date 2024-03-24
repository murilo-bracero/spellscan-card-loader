package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"spellscan.com/card-loader/objects"
)

type objectType int

const cardObject objectType = 1
const cardFaceObject objectType = 2

type Card struct {
	ID              string         `db:"id"`
	Name            string         `db:"card_name"`
	Lang            string         `db:"lang"`
	ReleasedAt      string         `db:"released_at"`
	Layout          string         `db:"layout"`
	ImageStatus     string         `db:"image_status"`
	ImageUris       *ImageUris     `db:"-"`
	CardFaces       []*CardFace    `db:"card_faces"`
	ManaCost        string         `db:"mana_cost"`
	TypeLine        string         `db:"type_line"`
	PrintedText     string         `db:"printed_text"`
	Colors          pq.StringArray `db:"colors"`
	ColorIdentity   pq.StringArray `db:"color_identity"`
	Keywords        pq.StringArray `db:"keywords"`
	Reserved        bool           `db:"reserved"`
	Foil            bool           `db:"foil"`
	Nonfoil         bool           `db:"nonfoil"`
	Promo           bool           `db:"promo"`
	Variation       bool           `db:"variation"`
	Set             string         `db:"card_set"`
	Rarity          string         `db:"rarity"`
	FlavorText      string         `db:"flavor_text"`
	Artist          string         `db:"artist"`
	Frame           string         `db:"frame"`
	FullArt         bool           `db:"full_art"`
	Textless        bool           `db:"textless"`
	CollectorNumber string         `db:"collector_number"`
}

func (c *Card) Save(db *sqlx.DB) error {

	query := `
		INSERT INTO cards (id, card_name, lang, released_at, layout, image_status, 
			mana_cost, type_line, printed_text, colors, color_identity, keywords, 
			reserved, foil, nonfoil, promo, variation, card_set, rarity, flavor_text, 
			artist, frame, full_art, textless, collector_number)
		VALUES (:id, :card_name, :lang, :released_at, :layout, :image_status, 
			:mana_cost, :type_line, :printed_text, :colors, :color_identity, :keywords, 
			:reserved, :foil, :nonfoil, :promo, :variation, :card_set, :rarity, :flavor_text, 
			:artist, :frame, :full_art, :textless, :collector_number)
		ON CONFLICT (id) DO UPDATE
		SET card_name = EXCLUDED.card_name, lang = EXCLUDED.lang, released_at = EXCLUDED.released_at,
			layout = EXCLUDED.layout, image_status = EXCLUDED.image_status,
			mana_cost = EXCLUDED.mana_cost, type_line = EXCLUDED.type_line,
			printed_text = EXCLUDED.printed_text, colors = EXCLUDED.colors, color_identity = EXCLUDED.color_identity,
			keywords = EXCLUDED.keywords, reserved = EXCLUDED.reserved, foil = EXCLUDED.foil, nonfoil = EXCLUDED.nonfoil,
			promo = EXCLUDED.promo, variation = EXCLUDED.variation, card_set = EXCLUDED.card_set,
			rarity = EXCLUDED.rarity, flavor_text = EXCLUDED.flavor_text, artist = EXCLUDED.artist, frame = EXCLUDED.frame,
			full_art = EXCLUDED.full_art, textless = EXCLUDED.textless, collector_number = EXCLUDED.collector_number
		`

	if _, err := db.NamedExec(query, c); err != nil {
		return err
	}

	if err := c.ImageUris.Save(db); err != nil {
		return err
	}

	for _, cf := range c.CardFaces {
		if err := cf.Save(db); err != nil {
			return err
		}
	}

	return nil
}

func FromCardJson(card *objects.Card) *Card {
	carddb := &Card{
		ID:              card.ID,
		Name:            card.Name,
		Lang:            card.Lang,
		ReleasedAt:      card.ReleasedAt,
		Layout:          card.Layout,
		ImageStatus:     card.ImageStatus,
		ImageUris:       fromImageUrisJson(card.ID, &card.ImageUris, cardObject),
		CardFaces:       fromCardFacesJson(card.ID, card.CardFaces),
		ManaCost:        card.ManaCost,
		TypeLine:        card.TypeLine,
		PrintedText:     card.PrintedText,
		Colors:          card.Colors,
		ColorIdentity:   card.ColorIdentity,
		Keywords:        card.Keywords,
		Reserved:        card.Reserved,
		Foil:            card.Foil,
		Nonfoil:         card.Nonfoil,
		Promo:           card.Promo,
		Variation:       card.Variation,
		Set:             card.Set,
		Rarity:          card.Rarity,
		FlavorText:      card.FlavorText,
		Artist:          card.Artist,
		Frame:           card.Frame,
		FullArt:         card.FullArt,
		Textless:        card.Textless,
		CollectorNumber: card.CollectorNumber,
	}

	if card.PrintedName != "" {
		carddb.Name = card.PrintedName
	}

	if card.PrintedTypeLine != "" {
		carddb.TypeLine = card.PrintedTypeLine
	}

	if card.PrintedText == "" {
		carddb.PrintedText = card.OracleText
	}

	return carddb
}

func fromImageUrisJson(objectId string, imageUris *objects.ImageUris, ot objectType) *ImageUris {
	iu := &ImageUris{
		Small:      imageUris.Small,
		Normal:     imageUris.Normal,
		Large:      imageUris.Large,
		Png:        imageUris.Png,
		ArtCrop:    imageUris.ArtCrop,
		BorderCrop: imageUris.BorderCrop,
	}

	if ot == cardFaceObject {
		iu.CardFaceId = sql.NullString{String: objectId, Valid: true}
	}

	if ot == cardObject {
		iu.CardId = sql.NullString{String: objectId, Valid: true}
	}

	return iu
}

func fromCardFacesJson(cardId string, cardFaces []objects.CardFace) []*CardFace {
	var dbcf []*CardFace

	for _, raw := range cardFaces {
		entity := &CardFace{
			CardId:         cardId,
			Name:           raw.Name,
			ManaCost:       raw.ManaCost,
			TypeLine:       raw.TypeLine,
			PrintedText:    raw.OracleText,
			FlavorText:     raw.FlavorText,
			Colors:         raw.Colors,
			ColorIndicator: raw.ColorIndicator,
			ImageUris:      fromImageUrisJson("", &raw.ImageUris, cardFaceObject),
		}

		dbcf = append(dbcf, entity)
	}

	return dbcf
}
