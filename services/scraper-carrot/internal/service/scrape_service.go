package service

import (
	"context"

	"scraper-carrot/internal/client"
	"scraper-carrot/internal/filter"
	"scraper-carrot/internal/models"
	"scraper-carrot/internal/parser"
)

type ScrapeService struct {
	client *client.DaangnClient
	filter *filter.CardFilter
}

func NewScrapeService(c *client.DaangnClient, f *filter.CardFilter) *ScrapeService {
	return &ScrapeService{
		client: c,
		filter: f,
	}
}

func (s *ScrapeService) Scrape(ctx context.Context, url string) ([]models.Article, error) {
	page, err := s.client.FetchSearchPage(ctx, url)
	if err != nil {
		return nil, err
	}

	articles, err := parser.ExtractArticles(page)
	if err != nil {
		return nil, err
	}

	return s.filter.Apply(articles), nil
}
