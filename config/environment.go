package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DbDsn                   string
	DbMaxConnections        int
	MeiliApiKey             string
	MeiliUrl                string
	SkipDownload            bool
	UseReleaseDateReference bool
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Could not load .env file, using env variables instead")
	}

	return &Config{
		DbDsn:                   os.Getenv("DB_DSN"),
		DbMaxConnections:        parseIntVar("DB_MAX_CONNECTIONS"),
		MeiliApiKey:             os.Getenv("MEILI_API_KEY"),
		MeiliUrl:                os.Getenv("MEILI_URL"),
		SkipDownload:            boolOrFalse("SKIP_DOWNLOAD"),
		UseReleaseDateReference: boolOrFalse("USE_RELEASE_DATE_REFERENCE"),
	}
}

func boolOrFalse(variable string) bool {
	value, err := strconv.ParseBool(os.Getenv(variable))

	if err != nil {
		slog.Warn("Could not convert variable to bool", "variable", variable)
		return false
	}

	return value
}

func parseIntVar(variable string) int {
	value, err := strconv.Atoi(os.Getenv(variable))

	if err != nil {
		slog.Warn("Could not convert variable to bool", "variable", variable)
		return 0
	}

	return value
}
