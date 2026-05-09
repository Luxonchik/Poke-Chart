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

	remixCtx, err := parser.ExtractRemixContext(page)
	if err != nil {
		return nil, err
	}

	articles := remixCtx.State.LoaderData.Route.AllPage.Articles
	if len(articles) == 0 && remixCtx.State.LoaderData.Route.PoW.Challenge != "" {
		articles, err = s.client.FetchArticlesWithPoW(ctx, *remixCtx)
		if err != nil {
			return nil, err
		}
	}

	filtered := make([]models.Article, 0, len(articles))
	for _, article := range articles {
		if s.filter.IsValid(article) {
			filtered = append(filtered, article)
		}
	}

	return filtered, nil
}
