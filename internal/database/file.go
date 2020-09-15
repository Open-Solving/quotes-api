package database

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"os"
)

type fileDatabase struct {
	quotes   []QuoteEntity
	filePath string
}

func (f *fileDatabase) GetQuotes(pagination Pagination) ([]QuoteEntity, error) {
	startIndex := (pagination.Page - 1) * pagination.Size
	total := len(f.quotes)

	if startIndex >= total {
		return []QuoteEntity{}, nil
	}

	// clamp count if needed
	count := startIndex + pagination.Size
	if count >= total {
		count = total - 1
	}

	var quotes []QuoteEntity
	for i := startIndex; i < count; i++ {
		quotes = append(quotes, f.quotes[i])
	}

	return quotes, nil
}

func (f *fileDatabase) CountQuotes(text string) (int64, error) {
	if text == "" {
		return int64(len(f.quotes)), nil
	}

	count := 0
	for _, quote := range f.quotes {
		if quote.Text == text {
			count++
		}
	}

	return int64(count), nil
}

func (f *fileDatabase) AddQuote(quote QuoteEntity) (QuoteEntity, error) {
	quote.Id = primitive.NewObjectID()

	f.quotes = append(f.quotes, quote)

	if err := f.synchronize(); err != nil {
		return QuoteEntity{}, err
	}

	return quote, nil
}

func (f *fileDatabase) SetQuotes(quotes []QuoteEntity) ([]QuoteEntity, error) {
	for i, _ := range quotes {
		quotes[i].Id = primitive.NewObjectID()
	}

	f.quotes = quotes

	if err := f.synchronize(); err != nil {
		return nil, err
	}

	return quotes, nil
}

func NewFileDatabase(dsn string) (Database, error) {
	f, err := os.Open(dsn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var quotes []QuoteEntity
	if err := json.Unmarshal(b, &quotes); err != nil {
		return nil, err
	}

	return &fileDatabase{
		quotes:   quotes,
		filePath: dsn,
	}, nil
}

func (f *fileDatabase) synchronize() error {
	b, err := json.Marshal(f.quotes)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(f.filePath, b, 0640)
}
