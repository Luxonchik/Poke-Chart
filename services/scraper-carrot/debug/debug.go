package main

import (
	"context"
	"fmt"
	"net/http"
	"scraper-carrot/internal/client"
	"scraper-carrot/internal/parser"
)

func main() {
	url := "https://www.daangn.com/kr/buy-sell/s/?in=역삼동-6035&search=포켓몬+EX"

	daangnClient := client.NewDaangnClient(&http.Client{})
	page, err := daangnClient.FetchSearchPage(context.Background(), url)
	if err != nil {
		fmt.Println("Ошибка получения страницы:", err)
		return
	}

	remixCtx, err := parser.ExtractRemixContext(page)
	if err != nil {
		fmt.Println("Ошибка парсинга remix:", err)
		return
	}

	articles, err := daangnClient.FetchArticlesWithPoW(context.Background(), *remixCtx)
	if err != nil {
		fmt.Println("Ошибка PoW запроса:", err)
		return
	}

	fmt.Printf("Получено статей: %d\n", len(articles))
	if len(articles) == 0 {
		return
	}

	first := articles[0]
	fmt.Printf("ID: %s\n", first.ID)
	fmt.Printf("Title: %s\n", first.Title)
	fmt.Printf("CreatedAt(raw): %q\n", first.CreatedAt)
	fmt.Printf("BoostedAt(raw): %q\n", first.BoostedAt)
}