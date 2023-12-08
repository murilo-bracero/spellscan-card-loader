package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"spellscan.com/card-loader/objects"
)

func main() {
	client := getRedisConn()

	ctx := context.Background()

	localBulkData, err := getLocalBulkMetadata(ctx, client)

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

	client.JSONSet(ctx, "bulk-data", "$", remoteBulkData)

	if err := downloadBulkFile(remoteBulkData); err != nil {
		client.JSONClear(ctx, "bulk-data", "$")
		panic(err)
	}

	cardsChannel := make(chan *objects.Card)

	go sendCardsToChannel(cardsChannel)

	for card := range cardsChannel {
		if !hasSupportedLanguage(card.Lang) {
			continue
		}

		id := fmt.Sprintf("card:%s", card.ID)

		status := client.JSONSet(ctx, id, "$", card)

		if status.Err() != nil {
			panic(status.Err())
		}

		slog.Info("Saved", "id", id, "name", card.Name, "type", card.TypeLine, "set", card.SetID)
	}
}

func getRedisConn() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func getLocalBulkMetadata(ctx context.Context, client *redis.Client) (*objects.BulkMetadata, error) {
	raw := client.JSONGet(ctx, "bulk-data")

	if raw.Val() == "" {
		return &objects.BulkMetadata{}, nil
	}

	var bulkMetadata []objects.BulkMetadata
	err := json.Unmarshal([]byte(raw.Val()), &bulkMetadata)

	return &bulkMetadata[len(bulkMetadata)-1], err
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
