package config

import (
	"log/slog"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

func DbConnect(cfg *Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.DbDsn)

	if err != nil {
		slog.Error("Could not connect to database", "err", err)
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DbMaxConnections)

	return db, nil
}
