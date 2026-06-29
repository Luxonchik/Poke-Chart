package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"scraper-carrot/internal/client"
	"scraper-carrot/internal/filter"
	"scraper-carrot/internal/output"
	"scraper-carrot/internal/service"
)

func main() {
	url := "https://www.daangn.com/kr/buy-sell/s/?in=역삼동-6035&search=포켓몬+EX"

	daangnClient := client.NewDaangnClient(&http.Client{Timeout: 10 * time.Second})
	cardFilter := filter.NewCardFilter()
	scrapeService := service.NewScrapeService(daangnClient, cardFilter)

	fmt.Println("🥕 Scraper-Carrot запущен. Начинаем мониторинг...")

	for {
		fmt.Println("--------------------------------------------------")
		fmt.Printf("🔍 [%s] Запрашиваем 당근마켓...\n", time.Now().Format("15:04:05"))

		articles, err := scrapeService.Scrape(context.Background(), url)
		if err != nil {
			// log.Printf вместо log.Fatalf — не убиваем процесс на временной ошибке.
			log.Printf("⚠️ Ошибка скрейпинга: %v\n", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		fmt.Printf("📦 Получено сырых объявлений: %d\n", len(articles))
		validCount := 0

		for _, item := range articles {
			if !cardFilter.IsValid(item) {
				continue 
			}
			
			validCount++
			priceFloat, _ := strconv.ParseFloat(item.Price, 64)
			priceInt := int(priceFloat)
			
			output.SendToAPI(item, priceInt)
		}

		fmt.Printf("✅ Валидных покемонов отправлено в API: %d из %d\n", validCount, len(articles))

		// 5 минут — минимальный интервал, чтобы не получить бан по IP от 당근마켓.
		fmt.Println("💤 Сплю 5 минут...")
		time.Sleep(5 * time.Minute)
	}
}