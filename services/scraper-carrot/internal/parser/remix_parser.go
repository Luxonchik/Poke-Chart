package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"scraper-carrot/internal/models"
)

var remixContextRegexp = regexp.MustCompile(`window\.__remixContext = (\{.*?\});`)

// ExtractArticles returns the listings rendered on a daangn.com search page.
// It fails loudly when the page layout changes instead of returning zero articles.
func ExtractArticles(page []byte) ([]models.Article, error) {
	matches := remixContextRegexp.FindSubmatch(page)
	if len(matches) < 2 {
		return nil, errors.New("remix context not found on page")
	}

	var remixData models.RemixContext
	if err := json.Unmarshal(matches[1], &remixData); err != nil {
		return nil, fmt.Errorf("parse remix json: %w", err)
	}

	route, ok := remixData.State.LoaderData[models.SearchRouteKey]
	if !ok {
		return nil, fmt.Errorf("unexpected page structure: loader %q not found", models.SearchRouteKey)
	}
	return route.Articles, nil
}
