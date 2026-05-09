package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/lib/pq"
)

func main() {
	// Подключение к БД (Твой актуальный порт 5433)
	connStr := "postgres://admin:admin_password@localhost:5433/pokecart?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Ошибка настройки БД:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("❌ База данных недоступна. Проверь Docker! Ошибка:", err)
	}
	fmt.Println("✅ Успешное подключение к PostgreSQL!")

	const imgBaseURL = "https://cards.image.pokemonkorea.co.kr/data/"
	page := 163
	totalParsed := 4890

	fmt.Println("🚀 Начинаем PRO-сбор коллекции...")

	for {
		fmt.Printf("\n📄 Читаем страницу пагинации %d...\n", page)
		ids, rawImages, count := fetchCardPage(page)

		if count == 0 || len(ids) == 0 {
			fmt.Printf("\n🏁 Сбор полностью завершен! Сохранено в базу: %d карт.\n", totalParsed)
			break
		}

		for i, id := range ids {
			totalParsed++
			fullImageURL := imgBaseURL + rawImages[i]

			// Вызываем надежную функцию с 3 попытками
			name, cardNo, rarity, series := fetchCardDetails(id)

			// Если после всех попыток распарсить не удалось, пропускаем запись
			if name == "Unknown" {
				fmt.Printf("[%d] ⏭️ Пропуск ID %s (ошибка сети)\n", totalParsed, id)
				continue
			}

			// Запись в таблицу
			query := `
				INSERT INTO master_cards (id, name, image_url, card_no, rarity, series)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO UPDATE 
				SET name = EXCLUDED.name, 
				    image_url = EXCLUDED.image_url, 
				    card_no = EXCLUDED.card_no, 
				    rarity = EXCLUDED.rarity, 
				    series = EXCLUDED.series;
			`

			_, err := db.Exec(query, id, name, fullImageURL, cardNo, rarity, series)
			if err != nil {
				fmt.Printf("[%d] ❌ Ошибка БД для ID %s: %v\n", totalParsed, id, err)
			} else {
				fmt.Printf("[%d] ✅ Сохранено: %s | %s\n", totalParsed, name, series)
			}

			// Обязательная пауза, чтобы не забанили IP
			time.Sleep(1 * time.Second)
		}
		page++
	}
}

// Надежный сбор списка ID (с полными заголовками)
func fetchCardPage(page int) ([]string, []string, int) {
	url := "https://pokemoncard.co.kr/v2/ajax2_dev2"
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("action", "search_text_cards")
	w.WriteField("search_text", "스칼렛&바이올렛")
	w.WriteField("search_params", "all")
	w.WriteField("limit", strconv.Itoa(page)) // Передаем как строку
	w.Close()

	req, _ := http.NewRequest("POST", url, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	// Полный User-Agent под Windows критически важен здесь
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://pokemoncard.co.kr")
	req.Header.Set("Referer", "https://pokemoncard.co.kr/cards")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, 0
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)

	var data struct {
		Status bool                   `json:"status"`
		Count  int                    `json:"count"`
		Result map[string]interface{} `json:"result"`
	}
	json.Unmarshal(bodyBytes, &data)

	var ids, images []string
	if data.Result != nil {
		for _, cardData := range data.Result {
			if cardMap, ok := cardData.(map[string]interface{}); ok {
				if cNum, ok := cardMap["CardNum"].(string); ok {
					ids = append(ids, cNum)
					if img, ok := cardMap["feature_image"].(string); ok {
						images = append(images, img)
					} else {
						images = append(images, "")
					}
				}
			}
		}
	}
	return ids, images, data.Count
}

// PRO-версия с Retry-логикой (3 попытки)
func fetchCardDetails(cardID string) (string, string, string, string) {
	url := "https://pokemoncard.co.kr/cards/detail/" + cardID
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		client := &http.Client{Timeout: 10 * time.Second}
		// Явно получаем ошибку (err) и не игнорируем её
		resp, err := client.Do(req)

		// Успешный ответ
		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			doc, docErr := goquery.NewDocumentFromReader(resp.Body)
			
			if docErr == nil {
				name := strings.TrimSpace(doc.Find(".card-hp.title").Text())
				cardNoRaw := doc.Find(".p_num").Contents().Not("span").Text()
				cardNo := strings.TrimSpace(cardNoRaw)
				rarity := strings.TrimSpace(doc.Find(".rarity-mark").Text()) 
				series := strings.TrimSpace(doc.Find(".search_href").First().Text())

				if name != "" {
					return name, cardNo, rarity, series
				}
			}
		}

		// Если произошла ошибка сети или плохой статус - закрываем тело, если оно есть
		if resp != nil {
			resp.Body.Close()
		}

		// Логируем попытку и ждем перед новой
		fmt.Printf("   ⚠️ Заминка с %s (Попытка %d/%d). Пробуем еще раз...\n", cardID, attempt, maxRetries)
		time.Sleep(2 * time.Second)
	}

	// Если все 3 попытки провалились
	return "Unknown", "", "", ""
}