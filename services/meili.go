package services

import (
	"errors"

	"github.com/meilisearch/meilisearch-go"
	"spellscan.com/card-loader/objects"
)

const cardsIndexName = "cards"

var ErrTaskFailed = errors.New("task failed")

type MeiliService interface {
	SaveAll(cards []*objects.Card) error
	UpdateIndexes() error
	DeleteAll() error
}

type meiliService struct {
	client *meilisearch.Client
}

func NewMeiliService(client *meilisearch.Client) MeiliService {
	return &meiliService{client: client}
}

func (m *meiliService) SaveAll(cards []*objects.Card) error {
	var searchCards []*objects.CardSearch

	for _, card := range cards {
		cardSearch := &objects.CardSearch{
			ID:  card.ID,
			Set: card.Set,
		}

		if card.PrintedName != "" {
			cardSearch.Name = card.PrintedName
		}

		if card.PrintedText == "" {
			cardSearch.Text = card.OracleText
		}

		searchCards = append(searchCards, cardSearch)
	}

	res, err := m.client.Index(cardsIndexName).AddDocuments(searchCards)

	if err != nil {
		return err
	}

	if res.Status == meilisearch.TaskStatusFailed {
		return ErrTaskFailed
	}

	return nil
}

func (m *meiliService) UpdateIndexes() error {
	resp, err := m.client.Index(cardsIndexName).UpdateFilterableAttributes(&[]string{
		"set",
	})

	if err != nil {
		return err
	}

	if resp.Status == meilisearch.TaskStatusFailed {
		return ErrTaskFailed
	}

	return nil
}

func (m *meiliService) DeleteAll() error {
	res, err := m.client.Index(cardsIndexName).DeleteAllDocuments()

	if err != nil {
		return err
	}

	if res.Status == meilisearch.TaskStatusFailed {
		return ErrTaskFailed
	}

	return nil
}
