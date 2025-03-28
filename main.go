package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"spellscan.com/card-loader/config"
	"spellscan.com/card-loader/models"
	"spellscan.com/card-loader/objects"
	"spellscan.com/card-loader/services"
)

type semaphore chan struct{}

func (s semaphore) acquire() {
	s <- struct{}{}
}

func (s semaphore) release() {
	<-s
}

const max_semaphore = 100

func main() {
	cfg := config.LoadConfig()

	meiliClient := config.MeiliConnect(cfg)

	meiliService := services.NewMeiliService(meiliClient)

	db, err := config.DbConnect(cfg)

	if err != nil {
		os.Exit(1)
	}

	metadataService := services.NewMetadataService(db, cfg)

	jobResult, err := metadataService.GetLastJobResult()

	if err != nil {
		slog.Error("Could not get bulk metadata from database", "err", err)
		os.Exit(1)
	}

	remoteBulkData, err := metadataService.GetRemoteBulkMetadata()

	if err != nil {
		slog.Error("Could not get bulk metadata from remote server", "err", err)
		os.Exit(1)
	}

	if remoteBulkData.Size == jobResult.Size {
		slog.Info("Same data, nothing to do", "size", jobResult.Size)
		return
	}

	if err := metadataService.DownloadBulkFile(remoteBulkData); err != nil {
		slog.Error("Could not download bulk metadata from remote server", "err", err)
		os.Exit(1)
	}

	if err := meiliService.DeleteAll(); err != nil {
		slog.Error("Could not delete data from meilisearch", "err", err)
		os.Exit(1)
	}

	var releaseDateReference time.Time

	if cfg.UseReleaseDateReference {
		rows, err := db.Query("SELECT max(released_at) FROM public.cards")

		if err != nil {
			slog.Warn("Could not fetch max release date from database", "error", err)
		}

		if rows.Next() {
			if err := rows.Scan(&releaseDateReference); err != nil {
				slog.Warn("Could not fetch max release date from database", "error", err)
			}
		}
	}

	start := time.Now()

	slog.Info("Started insertion job", "start", start)

	cardsChannel := make(chan *objects.Card)

	go sendCardsToChannel(cardsChannel)

	var cards []*objects.Card

	wg := new(sync.WaitGroup)

	s := make(semaphore, max_semaphore)

	for card := range cardsChannel {
		if !isCardValid(card, &releaseDateReference) {
			continue
		}

		cards = append(cards, card)

		wg.Add(1)
		go saveCard(db, card, wg, &s)
		s.acquire()

		if len(cards) == 100 {
			err := meiliService.SaveAll(cards)

			if err != nil {
				slog.Error("Could not save cards in meilisearch", "err", err)
				os.Exit(1)
			}

			cards = nil
		}
	}

	if len(cards) != 0 {
		err := meiliService.SaveAll(cards)

		if err != nil {
			slog.Error("Error while processing remanescent cards", "err", err)
			os.Exit(1)
		}
		cards = nil
	}

	wg.Wait()

	end := time.Now()
	slog.Info("Ended insertion job", "duration", end.Unix()-start.Unix())

	if err := meiliService.UpdateIndexes(); err != nil {
		slog.Error("Could not update meili filter attributes", "error", err)
		os.Exit(1)
	}

	if err := metadataService.Save(remoteBulkData, start, end); err != nil {
		slog.Error("Could not save job result in database", "err", err)
		os.Exit(1)
	}
}

func saveCard(db *sqlx.DB, card *objects.Card, wg *sync.WaitGroup, s *semaphore) {
	entity := models.FromCardJson(card)

	if err := entity.Save(db); err != nil {
		slog.Error("Could not save card in database", "cardId", card.ID, "err", err.Error())
		os.Exit(1)
	}

	slog.Info("Saved", "cardId", card.ID)

	s.release()
	wg.Done()
}

func isCardValid(card *objects.Card, releaseDateReference *time.Time) bool {
	if card.Digital {
		return false
	}

	if !hasSupportedLanguage(card.Lang) {
		return false
	}

	if !hasSupportedLayout(card.Layout) {
		return false
	}

	releasedAt, _ := time.Parse(time.DateOnly, card.ReleasedAt)

	if releasedAt.After(time.Now()) {
		return false
	}

	if releaseDateReference != nil {
		if releasedAt.Unix() < releaseDateReference.Unix() {
			return false
		}
	}

	return true
}

func sendCardsToChannel(c chan *objects.Card) {
	f, err := os.Open("./tmp/bulk_data.json")

	if err != nil {
		slog.Error("Could not open temp folder with bulk data json file", "err", err)
		os.Exit(1)
	}

	dec := json.NewDecoder(f)

	_, err = dec.Token()

	if err != nil {
		slog.Error("Could not decode token of bulk data json file", "err", err)
		os.Exit(1)
	}

	for dec.More() {
		var card objects.Card
		err := dec.Decode(&card)
		if err != nil {
			slog.Error("Could not decode json slice into card", "err", err)
			os.Exit(1)
		}

		c <- &card
	}

	_, err = dec.Token()

	if err != nil {
		slog.Error("Could not decode token of bulk data json file", "err", err)
		os.Exit(1)
	}

	close(c)
}

func hasSupportedLayout(layout string) bool {
	unsupportedLayouts := []string{"token", "emblem", "augment", "host", "vanguard", "reversible_card", "scheme", "art_series", "double_faced_token"}

	for _, sl := range unsupportedLayouts {
		if sl == layout {
			return false
		}
	}
	return true
}

func hasSupportedLanguage(lang string) bool {
	supportedLanguages := []string{"en", "pt", "sp", "fr", "de", "it", "la"}

	for _, sl := range supportedLanguages {
		if sl == lang {
			return true
		}
	}
	return false
}
