package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"scraper-carrot/internal/output"
	"scraper-carrot/internal/service"
)

type ScrapeWorker struct {
	scraper   *service.ScrapeService
	apiClient *output.APIClient
	source    string
	searchURL string
	interval  time.Duration
}

func NewScrapeWorker(
	scraper *service.ScrapeService,
	apiClient *output.APIClient,
	source string,
	searchURL string,
	interval time.Duration,
) *ScrapeWorker {
	return &ScrapeWorker{
		scraper:   scraper,
		apiClient: apiClient,
		source:    source,
		searchURL: searchURL,
		interval:  interval,
	}
}

func (w *ScrapeWorker) Start(ctx context.Context) {
	fmt.Printf("🥕 Scraper запущен [source=%s]\n", w.source)
	for {
		w.run(ctx)

		select {
		case <-ctx.Done():
			fmt.Println("🛑 Scraper остановлен")
			return
		case <-time.After(w.interval):
		}
	}
}

func (w *ScrapeWorker) run(ctx context.Context) {
	fmt.Println("--------------------------------------------------")
	fmt.Printf("🔍 [%s] Запрашиваем %s...\n", time.Now().Format("15:04:05"), w.source)

	articles, err := w.scraper.Scrape(ctx, w.searchURL)
	if err != nil {
		log.Printf("⚠️ Ошибка скрейпинга: %v", err)
		return
	}

	fmt.Printf("📦 Объявлений после фильтрации: %d\n", len(articles))

	result, err := w.apiClient.SyncListings(w.source, articles)
	if err != nil {
		log.Printf("❌ Ошибка синхронизации: %v", err)
		return
	}

	fmt.Printf("✅ Sync: новых=%d обновлено=%d закрыто=%d\n",
		result.New, result.Updated, result.Closed)
}
