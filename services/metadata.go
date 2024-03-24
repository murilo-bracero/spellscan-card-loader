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
	"spellscan.com/card-loader/models"
	"spellscan.com/card-loader/objects"
)

const scryfallBaseUrl = "https://api.scryfall.com/bulk-data"

var ErrScryfallNotAvailable = errors.New("scryfall is not available")

type MetadataService interface {
	GetLastJobResult() (*models.JobResult, error)
	GetRemoteBulkMetadata() (*objects.BulkMetadata, error)
	DownloadBulkFile(data *objects.BulkMetadata) error
	Save(bm *objects.BulkMetadata, start time.Time, end time.Time) error
}

type metadataService struct {
	db *sqlx.DB
}

func NewMetadataService(db *sqlx.DB) MetadataService {
	return &metadataService{db: db}
}

func (m *metadataService) GetLastJobResult() (*models.JobResult, error) {
	var jobResult models.JobResult
	err := m.db.Get(&jobResult, "SELECT * FROM job_results ORDER BY reference_date DESC LIMIT 1")

	if err == sql.ErrNoRows {
		return &models.JobResult{}, nil
	}

	if err != nil {
		return nil, err
	}

	return &jobResult, nil
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

func (m *metadataService) Save(bm *objects.BulkMetadata, start time.Time, end time.Time) error {
	jr := &models.JobResult{
		Size:          bm.Size,
		ReferenceDate: bm.UpdatedAt,
		Started:       start,
		Finished:      end,
	}

	if err := jr.Save(m.db); err != nil {
		return err
	}

	return nil
}
