package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"spellscan.com/card-loader/objects"
)

const scryfallBaseUrl = "https://api.scryfall.com/bulk-data"

var ErrScryfallNotAvailable = errors.New("scryfall is not available")

type MetadataService interface {
	GetLocalBulkMetadata() (*objects.BulkMetadata, error)
	GetRemoteBulkMetadata() (*objects.BulkMetadata, error)
	DownloadBulkFile(data *objects.BulkMetadata) error
	Save(bm *objects.BulkMetadata) error
}

type metadataService struct {
	db *sqlx.DB
}

func NewMetadataService(db *sqlx.DB) MetadataService {
	return &metadataService{db: db}
}

func (m *metadataService) GetLocalBulkMetadata() (*objects.BulkMetadata, error) {
	var bulkMetadata objects.BulkMetadata
	err := m.db.Get(&bulkMetadata, "SELECT * FROM bulk_metadata ORDER BY updated_at DESC LIMIT 1")

	if err == sql.ErrNoRows {
		return &objects.BulkMetadata{}, nil
	}

	if err != nil {
		return nil, err
	}

	return &bulkMetadata, nil
}

func (m *metadataService) GetRemoteBulkMetadata() (*objects.BulkMetadata, error) {
	res, err := http.Get(scryfallBaseUrl)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, ErrScryfallNotAvailable
	}

	var root objects.BulkMetadataRoot
	err = json.NewDecoder(res.Body).Decode(&root)
	if err != nil {
		return nil, err
	}

	var bulkMetadata objects.BulkMetadata

	for _, object := range root.Data {
		if object.Type == "all_cards" {
			bulkMetadata = object
		}
	}

	return &bulkMetadata, nil
}

func (m *metadataService) DownloadBulkFile(data *objects.BulkMetadata) error {
	if os.Getenv("SKIP_DOWNLOAD") == "true" {
		slog.Info("Skipping Download")
		return nil
	}

	if err := os.Mkdir("tmp", 0700); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	out, err := os.Create("./tmp/bulk_data.json")

	if err != nil {
		return err
	}

	defer out.Close()

	start := time.Now()

	slog.Info("Starting downloading bulk data", "start", start)

	res, err := http.Get(data.DownloadURI)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return ErrScryfallNotAvailable
	}

	defer res.Body.Close()

	_, err = io.Copy(out, res.Body)

	slog.Info("Finished download bulk data", "duration", time.Now().Unix()-start.Unix())

	return err
}

func (m *metadataService) Save(bm *objects.BulkMetadata) error {
	query := `
		INSERT INTO bulk_metadata (object, id, type, updated_at, uri, name, description, size, download_uri, content_type, content_encoding)
		VALUES (:object, :id, :type, :updated_at, :uri, :name, :description, :size, :download_uri, :content_type, :content_encoding)
		`

	if _, err := m.db.NamedExec(query, bm); err != nil {
		return err
	}

	return nil
}
