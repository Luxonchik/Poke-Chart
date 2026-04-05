package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 1. Описываем только тот путь, который нам нужен (всё остальное Go проигнорирует)
type RemixContext struct {
	State struct {
		LoaderData struct {
			// Обрати внимание, как мы маппим ключ со спецсимволами через тег json
			Route struct {
				AllPage struct {
					Articles []Article `json:"fleamarketArticles"`
				} `json:"allPage"`
			} `json:"routes/kr.buy-sell.s"`
		} `json:"loaderData"`
	} `json:"state"`
}

// 2. Наша структура карточки
type Article struct {
	ID        string `json:"id"`
	Href      string `json:"href"`
	Price     string `json:"price"` // Приходит как "100000.0"
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Status    string `json:"status"`
	Content  string `json:"content"`
	CreatedAt string `json:"createdAt"`
	BoostedAt string `json:"boostedAt"`
}

func isValidCard(item Article) bool {

	priceFloat, _ := strconv.ParseFloat(item.Price, 64)

	if priceFloat < 3000 {
		return false
	}


	fakePrices := []string{"123456.0", "12345.0"}
	for _, fake := range fakePrices {
		if item.Price == fake {
			return false
		}
	}

	stopWords := []string{
		// Мусор
		"피규어", "인형", "띠부씰", "케이스", "빈박스", "빈 박스", "닌텐도", "칩", "가챠", "figure", 
		// Опт
		"일괄", "처분", "대량", "모음", "다량", "랜덤", "통째로", "정리",
		// НОВЫЕ СЛОВА: Меню-посты
		"각종", // "всякие разные / все виды" (как на твоем скрине)
		"등등", // "и так далее / и прочие" (как на твоем скрине)
		"개별", // "поштучно / индивидуально" (когда пишут 개별판매 - продаю по одной из списка)
		"각각", // "каждая" (пишут 각각 가격 다름 - цены на каждую разные)
	}

	lowerTitle := strings.ToLower(item.Title)
	lowerContent := strings.ToLower(item.Content)


	for _, word := range stopWords {
		if strings.Contains(lowerTitle, word) || strings.Contains(lowerContent, word) {
			return false
		}
	}


	bulkRegex := regexp.MustCompile(`[2-9]\d+장|1\d{2,}장`) 
	if bulkRegex.MatchString(lowerTitle) || bulkRegex.MatchString(lowerContent) {
		return false
	}

	return true
}

func main() {
	// Подставляем твой Yeoksam-dong (Каннам) и ищем покемонов
	url := "https://www.daangn.com/kr/buy-sell/s/?in=역삼동-6035&search=포켓몬+카드"

	// Создаем HTTP клиент и запрос (добавляем User-Agent, чтобы не казаться ботом)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка при запросе: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа: %v", err)
	}

	// Вырезаем JSON с помощью регулярки
	re := regexp.MustCompile(`window\.__remixContext = (\{.*?\});`)
	matches := re.FindSubmatch(body)

	if len(matches) < 2 {
		log.Fatal("Не удалось найти JSON на странице. Возможно, капча или изменилась структура.")
	}

	jsonString := matches[1]

	// Создаем пустую структуру и "распаковываем" в нее наш JSON
	var remixData RemixContext
	err = json.Unmarshal(jsonString, &remixData)
	if err != nil {
		log.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	// Достаем массив карточек
	articles := remixData.State.LoaderData.Route.AllPage.Articles

	fmt.Printf("Успешно найдено карточек: %d\n", len(articles))
	fmt.Println("--------------------------------------------------")

	// Проходим по массиву и выводим результаты
	for _, item := range articles {

		if !isValidCard(item) {
			// continue означает "прервать текущий шаг и перейти к следующему item"
			continue 
		}



		// Конвертируем цену "10000.0" в нормальный integer (воны)
		priceFloat, _ := strconv.ParseFloat(item.Price, 64)
		priceInt := int(priceFloat)

		createdAtTime, err := time.Parse(time.RFC3339, item.CreatedAt)

		if err != nil {
			log.Printf("Ошибка парсинга даты: %v", err)
		}

		boostedAtTime, err := time.Parse(time.RFC3339, item.BoostedAt)

		if err != nil {
			log.Printf("Ошибка парсинга даты буста: %v", err)
		}
		
	

		fmt.Printf("Название: %s\n", item.Title)
		fmt.Printf("Цена: %d KRW\n", priceInt)
		fmt.Printf("Статус: %s\n", item.Status)
		fmt.Printf("Ссылка: %s\n", item.Href)
		fmt.Printf("Дата создания: %s\n", createdAtTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Дата буста: %s\n", boostedAtTime.Format("2006-01-02 15:04:05"))
		fmt.Println("--------------------------------------------------")
	}
}