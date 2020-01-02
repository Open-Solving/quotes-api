package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
)

type QuoteDto struct {
	Id     string `json:"id"`
	Text   string `json:"text"`
	Source string `json:"source"`
}

func main() {
	svr := echo.New()

	svr.GET("/quotes", getQuotesHandler())

	svr.Use(middleware.CORS())

	log.Fatal(svr.Start(":8080"))
}

func getQuotesHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, []QuoteDto{{Id: "1", Text: "Test #1", Source: "Wikipedia"}, {Id: "2", Text: "Text #2", Source: "Nop"}})
	}
}
