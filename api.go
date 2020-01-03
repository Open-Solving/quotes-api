package main

import (
	"context"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	paginationPageHeader     = "X-Pagination-Page"
	paginationSizeHeader     = "X-Pagination-Size"
	paginationCountHeader    = "X-Pagination-Count"
	paginationPageQueryParam = "pagination-page"
	paginationSizeQueryParam = "pagination-size"
)

type QuoteDto struct {
	Id     string `json:"id"`
	Text   string `json:"text"`
	Source string `json:"source"`
}

type QuoteEntity struct {
	Id     primitive.ObjectID `bson:"_id"`
	Text   string             `bson:"text"`
	Source string             `bson:"source"`
}

func main() {
	svr := echo.New()

	dbCtx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	dbClient, err := mongo.Connect(dbCtx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatalf("Unable to create database connection: %s", err)
	}
	if err := dbClient.Ping(dbCtx, readpref.Primary()); err != nil {
		log.Fatalf("Unable to connect to database: %s", err)
	}

	svr.GET("/quotes", getQuotesHandler(dbClient))

	// Setup CORS

	corsConfig := middleware.CORSConfig{
		Skipper:       middleware.DefaultSkipper,
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet},
		ExposeHeaders: []string{paginationPageHeader, paginationSizeHeader, paginationCountHeader},
	}

	svr.Use(middleware.CORSWithConfig(corsConfig))

	log.Fatal(svr.Start(":8080"))
}

func getQuotesHandler(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Acquire database collection + context
		collection := client.Database("quotes").Collection("quotes")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		// Acquire pagination result
		paginationPage, err := strconv.Atoi(c.QueryParam(paginationPageQueryParam))
		if err != nil {
			paginationPage = 1
		}
		paginationSize, err := strconv.Atoi(c.QueryParam(paginationSizeQueryParam))
		if err != nil {
			paginationSize = 50
		}

		var quotes []QuoteEntity

		// Set pagination options
		var opts options.FindOptions
		index := int64((paginationPage - 1) * paginationSize)
		limit := int64(paginationSize)
		opts.Skip = &index
		opts.Limit = &limit

		// Count number of documents
		count, err := collection.CountDocuments(ctx, bson.M{}, nil)
		if err != nil {
			c.Logger().Errorf("Error while querying database")
			return c.NoContent(http.StatusInternalServerError)
		}

		// Update pagination headers
		c.Response().Header().Set(paginationPageHeader, strconv.Itoa(paginationPage))
		c.Response().Header().Set(paginationSizeHeader, strconv.Itoa(paginationSize))
		c.Response().Header().Set(paginationCountHeader, strconv.FormatInt(count, 10))

		cur, err := collection.Find(ctx, bson.M{}, &opts)
		if err != nil {
			c.Logger().Errorf("Error while querying database")
			return c.NoContent(http.StatusInternalServerError)
		}

		if err := cur.All(ctx, &quotes); err != nil {
			c.Logger().Errorf("Error while querying database")
			return c.NoContent(http.StatusInternalServerError)
		}

		quoteDtos := make([]QuoteDto, len(quotes))
		for index, quote := range quotes {
			quoteDtos[index].Id = quote.Id.Hex()
			quoteDtos[index].Text = quote.Text
			quoteDtos[index].Source = quote.Source
		}

		return c.JSON(http.StatusOK, quoteDtos)
	}
}
