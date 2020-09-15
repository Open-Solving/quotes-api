package database

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

// todo make this abstract & split mongo bson in separate struct
type QuoteEntity struct {
	Id     primitive.ObjectID `bson:"_id" json:"id"`
	Text   string             `bson:"text" json:"text"`
	Source string             `bson:"source" json:"source"`
}

type Pagination struct {
	Page int
	Size int
}

type Database interface {
	GetQuotes(pagination Pagination) ([]QuoteEntity, error)
	CountQuotes(text string) (int64, error)
	AddQuote(quote QuoteEntity) (QuoteEntity, error)
	SetQuotes(quotes []QuoteEntity) ([]QuoteEntity, error)
}

func GetDatabase(dsn string) (Database, error) {
	if strings.HasPrefix(dsn, "mongodb://") {
		return NewMongoDatabase(dsn)
	}
	if strings.HasPrefix(dsn, "file://") {
		return NewFileDatabase(strings.TrimPrefix(dsn, "file://"))
	}
	return nil, fmt.Errorf("no database driver found for dsn %s", dsn)
}
