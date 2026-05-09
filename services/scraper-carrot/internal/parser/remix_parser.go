package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"scraper-carrot/internal/models"
)

var remixContextRegexp = regexp.MustCompile(`window\.__remixContext = (\{.*?\});`)

func ExtractArticles(page []byte) ([]models.Article, error) {
	ctx, err := ExtractRemixContext(page)
	if err != nil {
		return nil, err
	}

	return ctx.State.LoaderData.Route.AllPage.Articles, nil
}

func ExtractRemixContext(page []byte) (*models.RemixContext, error) {
	matches := remixContextRegexp.FindSubmatch(page)
	if len(matches) < 2 {
		return nil, errors.New("remix context not found on page")
	}

	var remixData models.RemixContext
	if err := json.Unmarshal(matches[1], &remixData); err != nil {
		return nil, fmt.Errorf("parse remix json: %w", err)
	}

	return &remixData, nil
}
