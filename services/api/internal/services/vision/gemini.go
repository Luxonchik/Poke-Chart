package vision

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"pokechart-api/internal/models"
)

// defaultGeminiModel is used when GEMINI_MODEL is not set. Google retires model
// versions regularly; switching models should only need a .env change.
const defaultGeminiModel = "gemini-3.8-flash"

func geminiModel() string {
	if m := os.Getenv("GEMINI_MODEL"); m != "" {
		return m
	}
	return defaultGeminiModel
}

// ErrModelBusy means Gemini rejected the request because of load or rate limits (429/503).
// The request itself is fine and should be retried later.
var ErrModelBusy = errors.New("Gemini временно перегружен")

func statusError(code int, body []byte) error {
	if code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable {
		return fmt.Errorf("%w (статус %d)", ErrModelBusy, code)
	}
	return fmt.Errorf("Gemini вернул статус %d: %s", code, string(body))
}

type VisionResult struct {
	IsPokemonCard bool   `json:"is_pokemon_card"`
	Name          string `json:"name"`
	Rarity        string `json:"rarity"`
	CardNo        string `json:"card_no"`
	Series        string `json:"series"`
}

// Gemini API request/response structs (package-level so they can be reused).
type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	Temperature      float64 `json:"temperature,omitempty"`
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
}

type geminiGenerateContentRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
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

var cardNoLike = regexp.MustCompile(`(?i)^\s*\d{1,3}\s*/\s*\d{1,3}\s*$`)

func sanitizeCardNo(cardNo string) string {
	s := strings.TrimSpace(cardNo)
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return ""
	}
	// If it doesn't look like NNN/NNN, treat it as hallucination.
	if !cardNoLike.MatchString(s) {
		return ""
	}
	return s
}

var exToken = regexp.MustCompile(`(?i)\bex\b`)
var rarityTokens = regexp.MustCompile(`(?i)\b(SAR|UR|SR|AR|RRR|RR|HR|CHR|CSR|ACE|K|R|U|C)\b`)

func sanitizeNameAndRarity(name string, rarity string) (string, string) {
	n := strings.TrimSpace(name)
	r := strings.TrimSpace(rarity)

	// Ensure "ex" is always lowercase in the name (EX/Ex -> ex).
	if n != "" {
		n = exToken.ReplaceAllStringFunc(n, func(_ string) string { return "ex" })
	}

	// If model accidentally appended rarity to the name, remove known rarity tokens from name
	// and keep them in rarity field if rarity is empty.
	if n != "" && rarityTokens.MatchString(n) {
		if r == "" {
			// pick first token as rarity
			m := rarityTokens.FindStringSubmatch(n)
			if len(m) > 0 {
				r = strings.ToUpper(m[0])
			}
		}
		n = strings.TrimSpace(rarityTokens.ReplaceAllString(n, ""))
		// Clean doubled spaces after removals
		n = strings.Join(strings.Fields(n), " ")
	}

	return n, r
}

func sanitizeSeriesCore(series string) string {
	s := strings.TrimSpace(series)
	if s == "" {
		return ""
	}
	// Keep it short and "core"-like to avoid overfitting to long labels.
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) > 40 {
		// If model returned a long label, take only the first chunk before "세트" or first 2 words.
		if i := strings.Index(s, "세트"); i > 0 {
			s = strings.TrimSpace(s[:i])
		} else {
			parts := strings.Fields(s)
			if len(parts) >= 2 {
				s = parts[0] + " " + parts[1]
			} else if len(parts) == 1 {
				s = parts[0]
			}
		}
	}
	return s
}

type resolveByImagesResult struct {
	Match        bool    `json:"match"`
	MasterCardID string  `json:"master_card_id"`
	Confidence   float64 `json:"confidence"`
}

func downloadAsInlineData(client *http.Client, imageURL string, maxBytes int64) (*geminiInlineData, error) {
	imgReq, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос к изображению: %w", err)
	}
	imgResp, err := client.Do(imgReq)
	if err != nil {
		return nil, fmt.Errorf("не удалось скачать изображение: %w", err)
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode < 200 || imgResp.StatusCode >= 300 {
		body, _ := io.ReadAll(imgResp.Body)
		return nil, fmt.Errorf("не удалось скачать изображение: статус %d: %s", imgResp.StatusCode, string(body))
	}

	mimeType := imgResp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	var reader io.Reader = imgResp.Body
	if maxBytes > 0 {
		reader = io.LimitReader(imgResp.Body, maxBytes)
	}
	imgBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать изображение: %w", err)
	}
	if maxBytes > 0 && int64(len(imgBytes)) >= maxBytes {
		return nil, fmt.Errorf("изображение слишком большое (>%d bytes), ограничь кандидатов или сожми", maxBytes)
	}

	return &geminiInlineData{
		MimeType: mimeType,
		Data:     base64.StdEncoding.EncodeToString(imgBytes),
	}, nil
}

// ResolveMasterCardByImages uses Gemini to pick the best master card among candidates by comparing images.
// It returns nil if Gemini can't confidently choose.
func ResolveMasterCardByImages(listingImageURL string, candidates []models.Card) (*models.Card, error) {
	if listingImageURL == "" {
		return nil, fmt.Errorf("пустой listingImageURL")
	}
	if len(candidates) < 2 {
		return nil, fmt.Errorf("нужно минимум 2 кандидата для сравнения")
	}
	if len(candidates) > 3 {
		candidates = candidates[:3] // cap: контроль размера токенов и стоимости
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("переменная GEMINI_API_KEY не задана")
	}
	modelName := geminiModel()

	client := &http.Client{Timeout: 60 * time.Second}

	const maxImageBytes = 3_500_000
	listingInline, err := downloadAsInlineData(client, listingImageURL, maxImageBytes)
	if err != nil {
		return nil, err
	}

	type candidateDescriptor struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		CardNo string `json:"card_no"`
		Rarity string `json:"rarity"`
		Series string `json:"series"`
	}

	descs := make([]candidateDescriptor, 0, len(candidates))
	for _, c := range candidates {
		descs = append(descs, candidateDescriptor{
			ID:     c.ID,
			Name:   c.Name,
			CardNo: c.CardNo,
			Rarity: c.Rarity,
			Series: c.Series,
		})
	}
	descsJSON, _ := json.Marshal(descs)

	parts := []geminiPart{
		{
			Text: "TASK: You will be given 1 marketplace photo (often cropped/glare) and 2-3 official reference card images.\n" +
				"Choose which reference image is the SAME card as the marketplace photo.\n" +
				"Return ONLY JSON: {\"match\":true|false, \"master_card_id\":\"<id or empty>\", \"confidence\":0..1}.\n" +
				"Rules: If you are not sure, set match=false, master_card_id=\"\", confidence<=0.4. Do not guess.\n" +
				"Candidate metadata (IDs are authoritative): " + string(descsJSON),
		},
		{Text: "MARKETPLACE_PHOTO:"},
		{InlineData: listingInline},
	}

	for i, c := range candidates {
		inline, err := downloadAsInlineData(client, c.ImageURL, maxImageBytes)
		if err != nil {
			return nil, err
		}
		parts = append(parts,
			geminiPart{Text: fmt.Sprintf("REFERENCE_%d id=%s name=%s card_no=%s rarity=%s series=%s", i+1, c.ID, c.Name, c.CardNo, c.Rarity, c.Series)},
			geminiPart{InlineData: inline},
		)
	}

	payload := geminiGenerateContentRequest{
		Contents: []geminiContent{
			{Role: "user", Parts: parts},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.0,
			ResponseMimeType: "application/json",
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать запрос Gemini: %w", err)
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", modelName)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос к Gemini: %w", err)
	}
	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, statusError(resp.StatusCode, bodyBytes)
	}

	var gr geminiGenerateContentResponse
	if err := json.Unmarshal(bodyBytes, &gr); err != nil {
		return nil, fmt.Errorf("не удалось распарсить ответ Gemini: %w; body=%s", err, string(bodyBytes))
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("пустой ответ Gemini: %s", string(bodyBytes))
	}

	rawText := gr.Candidates[0].Content.Parts[0].Text
	cleanContent := normalizeJSONContent(rawText)

	var rr resolveByImagesResult
	if err := json.Unmarshal([]byte(cleanContent), &rr); err != nil {
		return nil, fmt.Errorf("Gemini вернул невалидный JSON: %w; text=%q; body=%s", err, rawText, string(bodyBytes))
	}
	if !rr.Match || rr.MasterCardID == "" || rr.Confidence < 0.5 {
		return nil, nil
	}

	// Safety: only accept ids from the candidates we provided.
	for _, c := range candidates {
		if c.ID == rr.MasterCardID {
			cc := c
			return &cc, nil
		}
	}
	return nil, nil
}

func RecognizeCard(imageURL string) (*VisionResult, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("переменная GEMINI_API_KEY не задана")
	}

	modelName := geminiModel()

	promptText := `You are an expert Pokemon TCG appraiser for the Korean market. Analyze this image.
	
	STRICT RULES:
	1. TARGET: It MUST be a single, clear Pokemon card. If the image shows a sealed booster pack, a box, card sleeves, or just the back of a card, return {"is_pokemon_card": false}.
	2. LANGUAGE: The card MUST be printed in Korean (Hangul). If the text is Japanese (Katakana/Hiragana) or English, return {"is_pokemon_card": false}.
	3. HALLUCINATION CHECK: Do NOT invent names, HP, or descriptions. Extract ONLY the exact Pokemon name.
	4. NAME FORMAT: Put ONLY the Pokemon name (Korean) in "name". If the card is an "ex" card, append " ex" to the name and ALWAYS write it as lowercase "ex". Do NOT put rarity tokens (SAR/UR/SR/AR/etc.) into "name".
	5. RARITY FIELD: Put rarity (e.g., SAR/UR/SR/AR/RR) ONLY in "rarity". If rarity is not visible, set it to "".
	6. SERIES CORE: If the card belongs to a clear product line/series and it is visible/obvious from the card, set "series" to the core label (e.g., "스칼렛&바이올렛"). If you are not sure, set "series" to "".
	7. CARD NUMBER RULE: Only set "card_no" if the number is clearly readable on the card. If it is not clearly readable, set "card_no" to "" (empty string). Do NOT guess.
	8. FORMAT: Return ONLY a valid JSON object. No markdown, no explanations.
	
	Expected JSON if valid: {"is_pokemon_card": true, "name": "피카츄 ex", "rarity": "SAR", "card_no": "017/086", "series": "스칼렛&바이올렛"}
	Expected JSON if invalid/unsure: {"is_pokemon_card": false, "name": "", "rarity": "", "card_no": "", "series": ""}`

	client := &http.Client{Timeout: 40 * time.Second}
	inline, err := downloadAsInlineData(client, imageURL, 3_500_000)
	if err != nil {
		return nil, err
	}

	payload := geminiGenerateContentRequest{
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: promptText},
					{
						InlineData: inline,
					},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.0,
			ResponseMimeType: "application/json",
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("не удалось сериализовать запрос Gemini: %w", err)
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", modelName)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос к Gemini: %w", err)
	}
	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, statusError(resp.StatusCode, bodyBytes)
	}

	var gr geminiGenerateContentResponse
	if err := json.Unmarshal(bodyBytes, &gr); err != nil {
		return nil, fmt.Errorf("не удалось распарсить ответ Gemini: %w; body=%s", err, string(bodyBytes))
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("пустой ответ Gemini: %s", string(bodyBytes))
	}

	rawText := gr.Candidates[0].Content.Parts[0].Text
	cleanContent := normalizeJSONContent(rawText)

	var result VisionResult
	if err := json.Unmarshal([]byte(cleanContent), &result); err != nil {
		return nil, fmt.Errorf("Gemini вернул невалидный JSON: %w; text=%q; body=%s", err, rawText, string(bodyBytes))
	}

	// Protect downstream matching from hallucinated card numbers.
	result.CardNo = sanitizeCardNo(result.CardNo)
	result.Name, result.Rarity = sanitizeNameAndRarity(result.Name, result.Rarity)
	result.Series = sanitizeSeriesCore(result.Series)

	return &result, nil
}
