package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"time"
)

type mongoDatabase struct {
	client *mongo.Client
}

func (m *mongoDatabase) GetQuotes(pagination Pagination) ([]QuoteEntity, error) {
	collection, ctx := m.getQuotesCollection()

	var quotes []QuoteEntity

	// Set pagination options
	var opts options.FindOptions
	index := int64((pagination.Page - 1) * pagination.Size)
	limit := int64(pagination.Size)
	opts.Skip = &index
	opts.Limit = &limit

	cur, err := collection.Find(ctx, bson.M{}, &opts)
	if err != nil {
		return nil, err
	}

	if err := cur.All(ctx, &quotes); err != nil {
		return nil, err
	}

	return quotes, nil
}

func (m *mongoDatabase) CountQuotes(text string) (int64, error) {
	collection, ctx := m.getQuotesCollection()

	filter := bson.M{}

	if text != "" {
		filter["text"] = text
	}

	count, err := collection.CountDocuments(ctx, filter, nil)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (m *mongoDatabase) AddQuote(quote QuoteEntity) (QuoteEntity, error) {
	collection, ctx := m.getQuotesCollection()

	// Insert quote
	res, err := collection.InsertOne(ctx, QuoteEntity{
		Id:     primitive.NewObjectID(),
		Text:   quote.Text,
		Source: quote.Source,
	})

	if err != nil {
		return QuoteEntity{}, err
	}

	quoteId := res.InsertedID.(primitive.ObjectID)
	return QuoteEntity{
		Id:     quoteId,
		Text:   quote.Text,
		Source: quote.Source,
	}, nil
}

func (m *mongoDatabase) SetQuotes(quotes []QuoteEntity) ([]QuoteEntity, error) {
	collection, ctx := m.getQuotesCollection()

	// Delete old quotes
	if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
		return nil, err
	}

	// Insert new quotes
	quotesToCreate := make([]interface{}, len(quotes))
	for i, quote := range quotes {
		quotesToCreate[i] = QuoteEntity{
			Id:     primitive.NewObjectID(),
			Text:   quote.Text,
			Source: quote.Source,
		}
	}
	createdRes, err := collection.InsertMany(ctx, quotesToCreate)
	if err != nil {
		return nil, err
	}

	// Return list of created quotes with updated IDs
	for i := range quotes {
		quotes[i].Id = createdRes.InsertedIDs[i].(primitive.ObjectID)
	}

	return quotes, nil
}

func (m *mongoDatabase) RandomQuote() (QuoteEntity, error) {
	collection, ctx := m.getQuotesCollection()

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return QuoteEntity{}, err
	}

	if count == 0 {
		return QuoteEntity{}, ErrQuoteNotFound
	}

	index := rand.Int63n(count)

	// Set find options
	var opts options.FindOneOptions
	opts.Skip = &index

	var quote QuoteEntity
	if err := collection.FindOne(ctx, bson.M{}, &opts).Decode(&quote); err != nil {
		return QuoteEntity{}, err
	}

	return quote, nil
}

func (m *mongoDatabase) getQuotesCollection() (*mongo.Collection, context.Context) {
	// Acquire database collection + context
	collection := m.client.Database("quotes").Collection("quotes")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	return collection, ctx
}

func NewMongoDatabase(dsn string) (Database, error) {
	dbCtx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	dbClient, err := mongo.Connect(dbCtx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	return &mongoDatabase{
		client: dbClient,
	}, nil
}
