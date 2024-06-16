package config

import (
	"github.com/meilisearch/meilisearch-go"
)

func MeiliConnect(cfg *Config) *meilisearch.Client {
	return meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   cfg.MeiliUrl,
		APIKey: cfg.MeiliApiKey,
	})
}
