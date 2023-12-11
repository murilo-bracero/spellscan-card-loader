package objects

import "time"

type BulkMetadataRoot struct {
	Object  string         `json:"object"`
	HasMore bool           `json:"has_more"`
	Data    []BulkMetadata `json:"data"`
}

type BulkMetadata struct {
	Object          string    `json:"object" db:"object"`
	ID              string    `json:"id" db:"id"`
	Type            string    `json:"type" db:"type"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	URI             string    `json:"uri" db:"uri"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	Size            int       `json:"size" db:"size"`
	DownloadURI     string    `json:"download_uri" db:"download_uri"`
	ContentType     string    `json:"content_type" db:"content_type"`
	ContentEncoding string    `json:"content_encoding" db:"content_encoding"`
}
