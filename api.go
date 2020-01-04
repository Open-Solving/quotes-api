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

	defaultPaginationSize = 50
	maxPaginationSize     = 100

	authorizationHeader = "Authorization"
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

	// Setup custom logger
	svr.Logger = createLogger()

	dbCtx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	dbClient, err := mongo.Connect(dbCtx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatalf("Unable to create database connection: %s", err)
	}
	if err := dbClient.Ping(dbCtx, readpref.Primary()); err != nil {
		log.Fatalf("Unable to connect to database: %s", err)
	}

	svr.GET("/quotes", getQuotesHandler(dbClient))
	svr.POST("/quotes", addQuoteHandler(dbClient))

	// Setup CORS
	corsConfig := middleware.CORSConfig{
		Skipper:       middleware.DefaultSkipper,
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet},
		ExposeHeaders: []string{paginationPageHeader, paginationSizeHeader, paginationCountHeader},
	}

	svr.Use(middleware.CORSWithConfig(corsConfig))

	svr.Logger.Fatal(svr.Start(":8080"))
}

// Create custom configured handler
func createLogger() *log.Logger {
	logger := log.New("quotes-api")
	logger.SetLevel(log.DEBUG)
	logger.SetHeader("${time_rfc3339} ${level} - [${prefix}] ${short_file}:${line}")
	return logger
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
			paginationSize = defaultPaginationSize
		}
		// Prevent too much results from being returned
		if paginationSize > maxPaginationSize {
			paginationSize = maxPaginationSize
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

func addQuoteHandler(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		// make sure request is authorized
		if c.Request().Header.Get(authorizationHeader) != os.Getenv("AUTHORIZATION_KEY") {
			c.Logger().Warnf("Missing authorization key")
			return c.NoContent(http.StatusUnauthorized)
		}

		// Decode body
		var quote QuoteDto
		if err := c.Bind(&quote); err != nil {
			c.Logger().Warnf("Error while decoding json body: %s", err)
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		// Acquire database collection + context
		collection := client.Database("quotes").Collection("quotes")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		// Insert quote
		res, err := collection.InsertOne(ctx, QuoteEntity{
			Id:     primitive.NewObjectID(),
			Text:   quote.Text,
			Source: quote.Source,
		})

		if err != nil {
			c.Logger().Errorf("Error while creating quote: %s", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		quoteId := res.InsertedID.(primitive.ObjectID)

		c.Logger().Infof("New quote %s has been created", quoteId.Hex())

		// Return create quote
		quote.Id = quoteId.Hex()

		return c.JSON(http.StatusCreated, quote)
	}
}
