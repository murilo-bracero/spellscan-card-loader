package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type JobResult struct {
	ID            string    `db:"id"`
	Size          int       `db:"size"`
	ReferenceDate time.Time `db:"reference_date"`
	Started       time.Time `db:"started"`
	Finished      time.Time `db:"finished"`
}

func (j *JobResult) Save(db *sqlx.DB) error {
	j.ID = uuid.NewString()

	query := `
	INSERT INTO job_results (id,
		size,
		reference_date,
		started,
		finished)
	VALUES (:id,
		:size,
		:reference_date,
		:started,
		:finished)
	`

	if _, err := db.NamedExec(query, j); err != nil {
		return err
	}

	return nil
}
