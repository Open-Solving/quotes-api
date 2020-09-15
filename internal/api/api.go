package api

import (
	"github.com/creekorful/quotes-api/internal/database"
	"github.com/creekorful/quotes-api/internal/service"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"os"
	"strconv"
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

func NewAPI(dsn string) (*echo.Echo, error) {
	e := echo.New()

	// create the service
	svc, err := service.NewService(dsn)
	if err != nil {
		return nil, err
	}

	// register endpoints
	e.GET("/quotes", getQuotesHandler(svc))
	e.POST("/quotes", addQuoteHandler(svc))
	e.PUT("/quotes", setQuotesHandler(svc))

	// setup CORS
	corsConfig := middleware.CORSConfig{
		Skipper:       middleware.DefaultSkipper,
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet},
		ExposeHeaders: []string{paginationPageHeader, paginationSizeHeader, paginationCountHeader},
	}
	e.Use(middleware.CORSWithConfig(corsConfig))

	return e, nil
}

func getQuotesHandler(s *service.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		pagination := readPagination(c)
		results, count, err := s.GetQuotes(pagination)
		if err != nil {
			return err
		}

		writePagination(c, pagination, count)

		return c.JSON(http.StatusOK, results)
	}
}

func addQuoteHandler(s *service.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// make sure request is authorized
		if c.Request().Header.Get(authorizationHeader) != os.Getenv("AUTHORIZATION_KEY") {
			c.Logger().Warnf("Missing authorization key")
			return c.NoContent(http.StatusUnauthorized)
		}

		var quoteDto service.QuoteDto
		if err := c.Bind(&quoteDto); err != nil {
			return err
		}

		if quoteDto.Text == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		quote, err := s.AddQuote(quoteDto)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, quote)
	}
}

func setQuotesHandler(s *service.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// make sure request is authorized
		if c.Request().Header.Get(authorizationHeader) != os.Getenv("AUTHORIZATION_KEY") {
			c.Logger().Warnf("Missing authorization key")
			return c.NoContent(http.StatusUnauthorized)
		}

		var quotesDto []service.QuoteDto
		if err := c.Bind(&quotesDto); err != nil {
			return err
		}

		quotesDto, err := s.SetQuotes(quotesDto)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, quotesDto)
	}
}

func readPagination(c echo.Context) database.Pagination {
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

	return database.Pagination{
		Page: paginationPage,
		Size: paginationSize,
	}
}

func writePagination(c echo.Context, pagination database.Pagination, count int64) {
	c.Response().Header().Set(paginationPageHeader, strconv.Itoa(pagination.Page))
	c.Response().Header().Set(paginationSizeHeader, strconv.Itoa(pagination.Size))
	c.Response().Header().Set(paginationCountHeader, strconv.FormatInt(count, 10))
}
