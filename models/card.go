package models

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"spellscan.com/card-loader/objects"
)

type Card struct {
	ID            string         `db:"id"`
	Name          string         `db:"name"`
	Lang          string         `db:"lang"`
	ReleasedAt    string         `db:"released_at"`
	Layout        string         `db:"layout"`
	HighresImage  bool           `db:"highres_image"`
	ImageStatus   string         `db:"image_status"`
	ImageUris     *ImageUris     `db:"image_uris"`
	CardFaces     []*CardFace    `db:"card_faces"`
	ManaCost      string         `db:"mana_cost"`
	TypeLine      string         `db:"type_line"`
	PrintedText   string         `db:"printed_text"`
	Colors        pq.StringArray `db:"colors"`
	ColorIdentity pq.StringArray `db:"color_identity"`
	Keywords      pq.StringArray `db:"keywords"`
	Reserved      bool           `db:"reserved"`
	Foil          bool           `db:"foil"`
	Nonfoil       bool           `db:"nonfoil"`
	Oversized     bool           `db:"oversized"`
	Promo         bool           `db:"promo"`
	Variation     bool           `db:"variation"`
	Set           string         `db:"set"`
	Rarity        string         `db:"rarity"`
	FlavorText    string         `db:"flavor_text"`
	Artist        string         `db:"artist"`
	Frame         string         `db:"frame"`
	FullArt       bool           `db:"full_art"`
	Textless      bool           `db:"textless"`
}

func (c *Card) Save(db *sqlx.DB) error {

	query := `
		INSERT INTO card (id, name, lang, released_at, layout, highres_image, image_status, 
			mana_cost, type_line, printed_text, colors, color_identity, keywords, 
			reserved, foil, nonfoil, oversized, promo, variation, set, rarity, flavor_text, 
			artist, frame, full_art, textless)
		VALUES (:id, :name, :lang, :released_at, :layout, :highres_image, :image_status, 
			:mana_cost, :type_line, :printed_text, :colors, :color_identity, :keywords, 
			:reserved, :foil, :nonfoil, :oversized, :promo, :variation, :set, :rarity, :flavor_text, 
			:artist, :frame, :full_art, :textless)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name, lang = EXCLUDED.lang, released_at = EXCLUDED.released_at,
			layout = EXCLUDED.layout, highres_image = EXCLUDED.highres_image, image_status = EXCLUDED.image_status,
			mana_cost = EXCLUDED.mana_cost, type_line = EXCLUDED.type_line,
			printed_text = EXCLUDED.printed_text, colors = EXCLUDED.colors, color_identity = EXCLUDED.color_identity,
			keywords = EXCLUDED.keywords, reserved = EXCLUDED.reserved, foil = EXCLUDED.foil, nonfoil = EXCLUDED.nonfoil,
			oversized = EXCLUDED.oversized, promo = EXCLUDED.promo, variation = EXCLUDED.variation, set = EXCLUDED.set,
			rarity = EXCLUDED.rarity, flavor_text = EXCLUDED.flavor_text, artist = EXCLUDED.artist, frame = EXCLUDED.frame,
			full_art = EXCLUDED.full_art, textless = EXCLUDED.textless
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

type ImageUris struct {
	CardId     string `db:"card_id"`
	Small      string `db:"small"`
	Normal     string `db:"normal"`
	Large      string `db:"large"`
	Png        string `db:"png"`
	ArtCrop    string `db:"art_crop"`
	BorderCrop string `db:"border_crop"`
}

func (iu *ImageUris) Save(db *sqlx.DB) error {
	query := `
	INSERT INTO image_uris (card_id, small, normal, large, png, art_crop, border_crop)
	VALUES (:card_id, :small, :normal, :large, :png, :art_crop, :border_crop)
	ON CONFLICT (card_id) DO UPDATE
	SET card_id = EXCLUDED.card_id, 
		small = EXCLUDED.small, 
		normal = EXCLUDED.normal,
		large = EXCLUDED.large, 
		png = EXCLUDED.png, 
		art_crop = EXCLUDED.art_crop,
		border_crop = EXCLUDED.border_crop
	`

	if _, err := db.NamedExec(query, iu); err != nil {
		return err
	}

	return nil
}

type CardFace struct {
	ID              string         `db:"id"`
	CardId          string         `db:"card_id"`
	Name            string         `db:"name"`
	ManaCost        string         `db:"mana_cost"`
	TypeLine        string         `db:"type_line"`
	PrintedText     string         `db:"printed_text"`
	FlavorText      string         `db:"flavor_text"`
	Colors          pq.StringArray `db:"colors"`
	ColorIndicator  pq.StringArray `db:"color_indicator"`
	ImageSmall      string         `db:"image_small"`
	ImageNormal     string         `db:"image_normal"`
	ImageLarge      string         `db:"image_large"`
	ImagePng        string         `db:"image_png"`
	ImageArtCrop    string         `db:"image_art_crop"`
	ImageBorderCrop string         `db:"image_border_crop"`
}

func (cf *CardFace) Save(db *sqlx.DB) error {
	cf.ID = uuid.NewString()

	query := `
	INSERT INTO card_faces (id, card_id ,name ,mana_cost ,type_line ,printed_text ,flavor_text ,colors ,color_indicator ,image_small ,
		image_normal ,image_large ,image_png ,image_art_crop ,image_border_crop)
	VALUES (:id, :card_id, :name, :mana_cost, :type_line, :printed_text, :flavor_text, :colors, :color_indicator, :image_small, :image_normal, 
		:image_large, :image_png, :image_art_crop, :image_border_crop)
	ON CONFLICT (id) DO UPDATE
	SET card_id = EXCLUDED.card_id,
		name = EXCLUDED.name,
		mana_cost = EXCLUDED.mana_cost,
		type_line = EXCLUDED.type_line,
		printed_text = EXCLUDED.printed_text,
		flavor_text = EXCLUDED.flavor_text,
		colors = EXCLUDED.colors,
		color_indicator = EXCLUDED.color_indicator,
		image_small = EXCLUDED.image_small,
		image_normal = EXCLUDED.image_normal,
		image_large = EXCLUDED.image_large,
		image_png = EXCLUDED.image_png,
		image_art_crop = EXCLUDED.image_art_crop,
		image_border_crop = EXCLUDED.image_border_crop
	`

	if _, err := db.NamedExec(query, cf); err != nil {
		return err
	}

	return nil
}

func FromCardJson(card *objects.Card) *Card {
	carddb := &Card{
		ID:            card.ID,
		Name:          card.Name,
		Lang:          card.Lang,
		ReleasedAt:    card.ReleasedAt,
		Layout:        card.Layout,
		HighresImage:  card.HighresImage,
		ImageStatus:   card.ImageStatus,
		ImageUris:     fromImageUrisJson(card.ID, &card.ImageUris),
		CardFaces:     fromCardFacesJson(card.ID, card.CardFaces),
		ManaCost:      card.ManaCost,
		TypeLine:      card.TypeLine,
		PrintedText:   card.PrintedText,
		Colors:        card.Colors,
		ColorIdentity: card.ColorIdentity,
		Keywords:      card.Keywords,
		Reserved:      card.Reserved,
		Foil:          card.Foil,
		Nonfoil:       card.Nonfoil,
		Oversized:     card.Oversized,
		Promo:         card.Promo,
		Variation:     card.Variation,
		Set:           card.Set,
		Rarity:        card.Rarity,
		FlavorText:    card.FlavorText,
		Artist:        card.Artist,
		Frame:         card.Frame,
		FullArt:       card.FullArt,
		Textless:      card.Textless,
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

func fromImageUrisJson(cardId string, imageUris *objects.ImageUris) *ImageUris {
	return &ImageUris{
		CardId:     cardId,
		Small:      imageUris.Small,
		Normal:     imageUris.Normal,
		Large:      imageUris.Large,
		Png:        imageUris.Png,
		ArtCrop:    imageUris.ArtCrop,
		BorderCrop: imageUris.BorderCrop,
	}
}

func fromCardFacesJson(cardId string, cardFaces []objects.CardFace) []*CardFace {
	var dbcf []*CardFace

	for _, raw := range cardFaces {
		dbcf = append(dbcf, &CardFace{
			CardId:          cardId,
			Name:            raw.Name,
			ManaCost:        raw.ManaCost,
			TypeLine:        raw.TypeLine,
			PrintedText:     raw.OracleText,
			FlavorText:      raw.FlavorText,
			Colors:          raw.Colors,
			ColorIndicator:  raw.ColorIndicator,
			ImageSmall:      raw.ImageUris.Small,
			ImageNormal:     raw.ImageUris.Normal,
			ImageLarge:      raw.ImageUris.Large,
			ImagePng:        raw.ImageUris.Png,
			ImageArtCrop:    raw.ImageUris.ArtCrop,
			ImageBorderCrop: raw.ImageUris.BorderCrop,
		})
	}

	return dbcf
}
