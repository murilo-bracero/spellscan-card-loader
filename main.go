package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"
	"spellscan.com/card-loader/models"
	"spellscan.com/card-loader/objects"
	"spellscan.com/card-loader/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Could not load .env file, using env variables instead")
	}

	meiliClient := getMeiliClient()

	meiliService := services.NewMeiliService(meiliClient)

	db, err := sqlx.Connect("pgx", os.Getenv("DB_DSN"))

	if err != nil {
		slog.Error("Could not connect to database", "err", err)
		os.Exit(1)
	}

	db.SetMaxOpenConns(1)

	metadataService := services.NewMetadataService(db)

	localBulkData, err := metadataService.GetLocalBulkMetadata()

	if err != nil {
		slog.Error("Could not get bulk metadata from database", "err", err)
		os.Exit(1)
	}

	remoteBulkData, err := metadataService.GetRemoteBulkMetadata()

	if err != nil {
		slog.Error("Could not get bulk metadata from remote server", "err", err)
		os.Exit(1)
	}

	if remoteBulkData.Size == localBulkData.Size {
		slog.Info("Same data, nothing to do", "size", localBulkData.Size)
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

	clearCardFaces(db)

	start := time.Now()

	slog.Info("Started insertion job", "start", start)

	cardsChannel := make(chan *objects.Card)

	go sendCardsToChannel(cardsChannel)

	var cards []*objects.Card

	var wg sync.WaitGroup

	for card := range cardsChannel {
		cards = append(cards, card)
		saveCard(db, card, &wg)

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

	slog.Info("Ended insertion job", "duration", time.Now().Unix()-start.Unix())

	if err := meiliService.UpdateIndexes(); err != nil {
		slog.Error("Could not update meili filter attributes", "error", err)
	}

	if err := metadataService.Save(remoteBulkData); err != nil {
		slog.Error("Could not save bulk metadata in database", "err", err)
		os.Exit(1)
	}
}

func saveCard(db *sqlx.DB, card *objects.Card, wg *sync.WaitGroup) {
	if !hasSupportedLanguage(card.Lang) || card.Digital {
		return
	}

	if card.Layout == "art_series" {
		return
	}

	entity := models.FromCardJson(card)

	wg.Add(1)
	go saveOnDb(db, entity, wg)

	slog.Info("Saved", "id", card.ID, "name", card.Name, "type", card.TypeLine, "set", card.Set)
}

func saveOnDb(db *sqlx.DB, card *models.Card, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := card.Save(db); err != nil {
		slog.Error("Could not save card in database", "cardId", card.ID, "err", err.Error())
		os.Exit(1)
	}
}

func clearCardFaces(db *sqlx.DB) {
	db.Exec("DELETE FROM card_faces")
}

func getMeiliClient() *meilisearch.Client {
	return meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   os.Getenv("MEILI_URL"),
		APIKey: os.Getenv("MEILI_API_KEY"),
	})
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

func hasSupportedLanguage(lang string) bool {
	supportedLanguages := []string{"en", "pt", "sp", "fr", "de", "it", "la"}

	for _, sl := range supportedLanguages {
		if sl == lang {
			return true
		}
	}
	return false
}
