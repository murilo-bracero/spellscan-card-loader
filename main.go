package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/meilisearch/meilisearch-go"
	"spellscan.com/card-loader/models"
	"spellscan.com/card-loader/objects"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Could not load .env file, using env variables instead.")
	}

	meiliClient := getMeiliClient()

	ctx := context.Background()

	db, err := sqlx.Connect("pgx", os.Getenv("DB_DSN"))

	if err != nil {
		panic(err)
	}

	localBulkData, err := getLocalBulkMetadata(ctx, db)

	if err != nil {
		panic(err)
	}

	remoteBulkData, err := getRemoteBulkMetadata()

	if err != nil {
		panic(err)
	}

	if remoteBulkData.Size == localBulkData.Size {
		log.Printf("Same data, nothing to do")
		return
	}

	if err != nil {
		panic(err)
	}

	if err := downloadBulkFile(remoteBulkData); err != nil {
		_, err := db.Query("DELETE FROM bulk_metadata")
		panic(err)
	}

	if err := eraseMeili(meiliClient); err != nil {
		panic(err)
	}

	cardsChannel := make(chan *objects.Card)

	clearCardFaces(db)

	start := time.Now()

	slog.Info("Started insertion job", "start", start)

	go sendCardsToChannel(cardsChannel)

	for card := range cardsChannel {
		if !hasSupportedLanguage(card.Lang) || card.Digital {
			continue
		}

		cardDb := models.FromCardJson(card)

		if err := cardDb.Save(db); err != nil {
			log.Println(card.Keywords)
			panic(err)
		}

		if err := saveOnMeili(meiliClient, card); err != nil {
			panic(err)
		}

		slog.Info("Saved", "id", card.ID, "name", card.Name, "type", card.TypeLine, "set", card.Set)
	}

	slog.Info("Ended insertion job", "end", time.Now().Unix()-start.Unix())

	if err := meiliUpdateIndexes(meiliClient); err != nil {
		slog.Error("Could not update meili filter attributes: ", "error", err)
	}
}

func meiliUpdateIndexes(client *meilisearch.Client) error {
	resp, err := client.Index("cards").UpdateFilterableAttributes(&[]string{
		"set",
	})

	if err != nil {
		return err
	}

	if resp.Status == meilisearch.TaskStatusFailed {
		return errors.New("task failed")
	}

	return nil
}

func eraseMeili(client *meilisearch.Client) error {
	res, err := client.Index("cards").DeleteAllDocuments()

	if err != nil {
		return err
	}

	if res.Status == meilisearch.TaskStatusFailed {
		return errors.New("task failed")
	}

	return nil
}

func saveOnMeili(client *meilisearch.Client, card *objects.Card) error {
	cardSearch := &objects.CardSearch{
		ID:   card.ID,
		Set:  card.Set,
		Text: card.PrintedText,
	}

	if card.PrintedName == "" {
		cardSearch.Name = card.Name
	} else {
		cardSearch.Name = card.PrintedName
	}

	res, err := client.Index("cards").AddDocuments(cardSearch)

	if err != nil {
		return err
	}

	if res.Status == meilisearch.TaskStatusFailed {
		return errors.New("task failed")
	}

	return nil
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

func getLocalBulkMetadata(ctx context.Context, db *sqlx.DB) (*objects.BulkMetadata, error) {

	var bulkMetadata objects.BulkMetadata
	err := db.Get(&bulkMetadata, "SELECT * FROM bulk_metadata ORDER BY updated_at DESC LIMIT 1")

	if err == sql.ErrNoRows {
		return &objects.BulkMetadata{}, nil
	}

	if err != nil {
		return nil, err
	}

	return &bulkMetadata, nil
}

func getRemoteBulkMetadata() (*objects.BulkMetadata, error) {
	url := "https://api.scryfall.com/bulk-data"

	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("scryfall is not available")
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

func downloadBulkFile(data *objects.BulkMetadata) error {
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

	res, err := http.Get(data.DownloadURI)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("scryfall API is not available")
	}

	defer res.Body.Close()

	_, err = io.Copy(out, res.Body)

	return err
}

func sendCardsToChannel(c chan *objects.Card) {
	f, err := os.Open("./tmp/bulk_data.json")

	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(f)

	_, err = dec.Token()

	if err != nil {
		panic(err)
	}

	for dec.More() {
		var card objects.Card
		err := dec.Decode(&card)
		if err != nil {
			panic(err)
		}

		c <- &card
	}

	_, err = dec.Token()

	if err != nil {
		panic(err)
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
