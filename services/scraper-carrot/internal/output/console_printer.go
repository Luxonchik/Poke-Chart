package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"scraper-carrot/internal/models"
)

// func PrintSummary(count int) {
// 	fmt.Printf("Успешно найдено карточек: %d\n", count)
// 	fmt.Println("--------------------------------------------------")
// }

// func PrintArticle(item models.Article) {
// 	priceFloat, _ := strconv.ParseFloat(item.Price, 64)
// 	priceInt := int(priceFloat)

// 	createdAtTime, err := time.Parse(time.RFC3339, item.CreatedAt)
// 	if err != nil {
// 		log.Printf("Ошибка парсинга даты: %v", err)
// 	}

// 	boostedAtTime, err := time.Parse(time.RFC3339, item.BoostedAt)
// 	if err != nil {
// 		log.Printf("Ошибка парсинга даты буста: %v", err)
// 	}

// 	fmt.Printf("Название: %s\n", item.Title)
// 	fmt.Printf("Цена: %d KRW\n", priceInt)
// 	fmt.Printf("Статус: %s\n", item.Status)
// 	fmt.Printf("Картинка: %s\n", item.Thumbnail)
// 	fmt.Printf("Ссылка: %s\n", item.Href)
// 	fmt.Printf("Дата создания: %s\n", createdAtTime.Format("2006-01-02 15:04:05"))
// 	fmt.Printf("Дата буста: %s\n", boostedAtTime.Format("2006-01-02 15:04:05"))
// 	fmt.Println("--------------------------------------------------")
// }


func ensureAbsoluteURL(href string) string {
	if strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "http://") {
		return href
	}
	return "https://www.daangn.com" + href
}

func SendToAPI(item models.Article, priceInt int) {
	// 1. Собираем данные ровно в том формате, который ждет наш API
	payload := map[string]interface{}{
		"daangn_id":  item.ID,
		"title":      item.Title,
		"price":      priceInt, // Отправляем как число
		"image_url":  item.Thumbnail,
		"source_url": ensureAbsoluteURL(item.Href),
		"created_at": item.CreatedAt,
		"boosted_at": item.BoostedAt,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Ошибка маршалинга JSON: %v", err)
		return
	}

	response, err := http.Post("http://localhost:8080/api/listings", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Ошибка отправки запроса: %v", err)
		return
	}
	defer response.Body.Close()

	// 2. Обрабатываем два успешных статуса (200 и 201)
	if response.StatusCode == http.StatusCreated {
		fmt.Printf("✅ [НОВОЕ] Отправлено в очередь: %s\n", item.Title)
	} else if response.StatusCode == http.StatusOK {
		fmt.Printf("⏭️ [ДУБЛЬ] Пропущено сервером: %s\n", item.Title)
	} else {
		// Читаем текст ошибки от сервера, чтобы понимать, что пошло не так
		bodyBytes, _ := io.ReadAll(response.Body)
		log.Printf("⚠️ Ошибка сервера %d: %s", response.StatusCode, string(bodyBytes))
	}
}