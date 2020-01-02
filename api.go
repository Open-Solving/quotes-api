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
	"time"
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

	svr.Use(middleware.CORS())

	log.Fatal(svr.Start(":8080"))
}

func getQuotesHandler(client *mongo.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		collection := client.Database("quotes").Collection("quotes")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		var quotes []QuoteEntity

		cur, err := collection.Find(ctx, bson.M{})
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
