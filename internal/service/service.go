package service

import (
	"github.com/creekorful/quotes-api/internal/database"
	"github.com/labstack/echo"
	"net/http"
)

type QuoteDto struct {
	Id     string `json:"id"`
	Text   string `json:"text"`
	Source string `json:"source"`
}

type Service struct {
	conn database.Database
}

func NewService(dsn string) (*Service, error) {
	conn, err := database.GetDatabase(dsn)
	if err != nil {
		return nil, err
	}

	return &Service{
		conn: conn,
	}, nil
}

func (s *Service) GetQuotes(pagination database.Pagination) ([]QuoteDto, int64, error) {
	quotes, err := s.conn.GetQuotes(pagination)
	if err != nil {
		return nil, -1, err
	}

	total, err := s.conn.CountQuotes("")
	if err != nil {
		return nil, -1, err
	}

	var quotesDto []QuoteDto
	for _, quote := range quotes {
		quotesDto = append(quotesDto, QuoteDto{
			Id:     quote.Id.Hex(),
			Text:   quote.Text,
			Source: quote.Source,
		})
	}

	return quotesDto, total, nil
}

func (s *Service) AddQuote(quoteDto QuoteDto) (QuoteDto, error) {
	// Make sure quote doesn't already exist
	count, err := s.conn.CountQuotes(quoteDto.Text)
	if err != nil {
		return QuoteDto{}, err
	}

	if count > 0 {
		return QuoteDto{}, echo.NewHTTPError(http.StatusConflict, "quote already exist")
	}

	q, err := s.conn.AddQuote(database.QuoteEntity{
		Text:   quoteDto.Text,
		Source: quoteDto.Source,
	})
	if err != nil {
		return QuoteDto{}, err
	}

	return QuoteDto{
		Id:     q.Id.Hex(),
		Text:   q.Text,
		Source: q.Source,
	}, nil
}

func (s *Service) SetQuotes(quotesDto []QuoteDto) ([]QuoteDto, error) {
	var quotes []database.QuoteEntity

	for _, quoteDto := range quotesDto {
		quotes = append(quotes, database.QuoteEntity{
			Text:   quoteDto.Text,
			Source: quoteDto.Source,
		})
	}

	quotes, err := s.conn.SetQuotes(quotes)
	if err != nil {
		return nil, err
	}

	var ret []QuoteDto
	for _, quote := range quotes {
		ret = append(ret, QuoteDto{
			Id:     quote.Id.Hex(),
			Text:   quote.Text,
			Source: quote.Source,
		})
	}

	return ret, nil
}
