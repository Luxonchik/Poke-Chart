package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scraper-carrot/internal/client"
	"scraper-carrot/internal/filter"
	"scraper-carrot/internal/output"
	"scraper-carrot/internal/service"
	"scraper-carrot/internal/worker"
)

func main() {
	searchURL := getEnv("DAANGN_SEARCH_URL", "https://www.daangn.com/kr/buy-sell/s/?in=역삼동-6035&search=포켓몬+EX")
	apiURL := getEnv("API_URL", "http://localhost:8080")

	daangnClient := client.NewDaangnClient(&http.Client{Timeout: 10 * time.Second})
	scrapeService := service.NewScrapeService(daangnClient, filter.NewCardFilter())
	apiClient := output.NewAPIClient(apiURL)

	// 5 минут — минимальный интервал, чтобы не получить бан от 당근마켓
	scrapeWorker := worker.NewScrapeWorker(scrapeService, apiClient, "daangn", searchURL, 5*time.Minute)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("🥕 Scraper-Carrot запущен. URL: %s\n", searchURL)
	scrapeWorker.Start(ctx)
	fmt.Println("👋 Scraper-Carrot остановлен")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
