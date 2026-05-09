package vision

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// VisionResult - структура того, что мы ждем от ИИ
type VisionResult struct {
	IsPokemonCard bool   `json:"is_pokemon_card"`
	Name          string `json:"name"`
	Rarity        string `json:"rarity"`
}

func normalizeJSONContent(raw string) string {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	start := strings.Index(clean, "{")
	end := strings.LastIndex(clean, "}")
	if start >= 0 && end > start {
		return clean[start : end+1]
	}
	return clean
}

// RecognizeCard отправляет фото в OpenRouter
func RecognizeCard(imageURL string) (*VisionResult, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("переменная OPENROUTER_API_KEY не задана")
	}

	// Бесплатные модели с поддержкой vision. Если одна недоступна, пробуем следующую.
	freeModels := []string{
		"google/gemini-2.5-flash-lite",
		"openrouter/free",
		"google/gemma-3-27b-it:free",
		"nvidia/nemotron-nano-12b-v2-vl:free",
	}

	promptText := `You are an expert Pokemon TCG appraiser for the Korean market. Analyze this image.
	
	STRICT RULES:
	1. TARGET: It MUST be a single, clear Pokemon card. If the image shows a sealed booster pack, a box, card sleeves, or just the back of a card, return {"is_pokemon_card": false}.
	2. LANGUAGE: The card MUST be printed in Korean (Hangul). If the text is Japanese (Katakana/Hiragana) or English, return {"is_pokemon_card": false}.
	3. HALLUCINATION CHECK: Do NOT invent names, HP, or descriptions. Extract ONLY the exact Pokemon name.
	4. FORMAT: Return ONLY a valid JSON object. No markdown, no explanations.
	
	Expected JSON if valid: {"is_pokemon_card": true, "name": "피카츄 ex", "rarity": "SAR"}
	Expected JSON if invalid/unsure: {"is_pokemon_card": false, "name": "", "rarity": ""}`

	client := &http.Client{Timeout: 40 * time.Second}
	const maxAttemptsPerModel = 3

	for _, modelName := range freeModels {
		for attempt := 1; attempt <= maxAttemptsPerModel; attempt++ {
			payload := map[string]interface{}{
				"model": modelName,
				"messages": []map[string]interface{}{
					{
						"role": "user",
						"content": []map[string]interface{}{
							{"type": "text", "text": promptText},
							{"type": "image_url", "image_url": map[string]string{"url": imageURL}},
						},
					},
				},
				"response_format": map[string]string{"type": "json_object"},
			}

			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("ошибка подготовки запроса к ИИ: %v", err)
			}

			req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonPayload))
			if err != nil {
				return nil, fmt.Errorf("ошибка создания HTTP-запроса: %v", err)
			}

			req.Header.Set("Authorization", "Bearer "+apiKey)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("⚠️ Сетевая ошибка для %s (попытка %d/%d): %v", modelName, attempt, maxAttemptsPerModel, err)
				if attempt < maxAttemptsPerModel {
					time.Sleep(time.Duration(1<<(attempt-1)) * time.Second)
				}
				continue
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
				log.Printf("⏳ %s вернул %d (попытка %d/%d), повторяем...", modelName, resp.StatusCode, attempt, maxAttemptsPerModel)
				if attempt < maxAttemptsPerModel {
					time.Sleep(time.Duration(1<<(attempt-1)) * time.Second)
				}
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("⏭️ Модель %s недоступна (status=%d): %s", modelName, resp.StatusCode, string(bodyBytes))
				break
			}

			var openRouterResp struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}

			if err := json.Unmarshal(bodyBytes, &openRouterResp); err != nil {
				return nil, fmt.Errorf("ошибка разбора ответа OpenRouter: %v", err)
			}

			if len(openRouterResp.Choices) == 0 || openRouterResp.Choices[0].Message.Content == "" {
				log.Printf("⚠️ Пустой ответ от %s, пробуем следующую модель", modelName)
				break
			}

			var result VisionResult
			cleanContent := normalizeJSONContent(openRouterResp.Choices[0].Message.Content)
			if err := json.Unmarshal([]byte(cleanContent), &result); err != nil {
				log.Printf("⚠️ Невалидный JSON от %s: %s", modelName, openRouterResp.Choices[0].Message.Content)
				break
			}

			return &result, nil
		}
	}

	return nil, fmt.Errorf("не удалось распознать карту: бесплатные vision-модели недоступны или вернули некорректный ответ")
}
