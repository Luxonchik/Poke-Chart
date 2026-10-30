package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"scraper-carrot/internal/client"
	"scraper-carrot/internal/filter"
	"scraper-carrot/internal/parser"
)

// Fetches the search page once and reports what the scraper would send.
// Run from services/scraper-carrot: go run ./debug
func main() {
	url := "https://www.daangn.com/kr/buy-sell/s/?in=역삼동-6035&search=포켓몬+EX"

	daangnClient := client.NewDaangnClient(&http.Client{Timeout: 10 * time.Second})
	page, err := daangnClient.FetchSearchPage(context.Background(), url)
	if err != nil {
		fmt.Println("Ошибка получения страницы:", err)
		return
	}

	articles, err := parser.ExtractArticles(page)
	if err != nil {
		fmt.Println("Ошибка парсинга:", err)
		return
	}

	valid := filter.NewCardFilter().Apply(articles)
	fmt.Printf("Объявлений на странице: %d, после фильтра: %d\n", len(articles), len(valid))

	for i, a := range valid {
		if i == 5 {
			break
		}
		fmt.Printf("- %s | %s | %s\n", a.Price, a.Title, a.Thumbnail)
	}
}
