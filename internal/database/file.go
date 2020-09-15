package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type fileDatabase struct {
	quotes []QuoteEntity
}

func (f *fileDatabase) GetQuotes(pagination Pagination) ([]QuoteEntity, error) {
	startIndex := (pagination.Page - 1) * pagination.Size
	total := len(f.quotes)

	if startIndex >= total {
		return []QuoteEntity{}, nil
	}

	// clamp count if needed
	count := pagination.Size
	if startIndex+pagination.Size >= total {
		count = total - startIndex
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
	return QuoteEntity{}, fmt.Errorf("not implemented")
}

func (f *fileDatabase) SetQuotes(quotes []QuoteEntity) ([]QuoteEntity, error) {
	return nil, fmt.Errorf("not implemented")
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
		quotes: quotes,
	}, nil
}
