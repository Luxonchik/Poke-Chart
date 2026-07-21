package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"scraper-carrot/internal/models"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

// SyncResult — ответ API после синхронизации.
type SyncResult struct {
	New     int `json:"new"`
	Updated int `json:"updated"`
	Closed  int `json:"closed"`
}

type rawListing struct {
	ExternalID string `json:"external_id"`
	Title      string `json:"title"`
	Price      int    `json:"price"`
	ImageURL   string `json:"image_url"`
	SourceURL  string `json:"source_url"`
	CreatedAt  string `json:"created_at"`
	BoostedAt  string `json:"boosted_at"`
}

type syncPayload struct {
	Source   string       `json:"source"`
	Listings []rawListing `json:"listings"`
}

// SyncListings отправляет весь текущий снимок объявлений одним запросом.
// API сам разберёт что новое, что обновилось, что закрылось.
func (c *APIClient) SyncListings(source string, articles []models.Article) (SyncResult, error) {
	listings := make([]rawListing, 0, len(articles))
	for _, a := range articles {
		price, err := parsePrice(a.Price)
		if err != nil {
			log.Printf("⚠️ Не удалось распарсить цену '%s' для '%s': %v", a.Price, a.Title, err)
		}
		listings = append(listings, rawListing{
			ExternalID: a.ID,
			Title:      a.Title,
			Price:      price,
			ImageURL:   a.Thumbnail,
			SourceURL:  ensureAbsoluteURL(a.Href),
			CreatedAt:  a.CreatedAt,
			BoostedAt:  a.BoostedAt,
		})
	}

	body, err := json.Marshal(syncPayload{Source: source, Listings: listings})
	if err != nil {
		return SyncResult{}, fmt.Errorf("marshal: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/sync", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return SyncResult{}, fmt.Errorf("post /api/sync: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return SyncResult{}, fmt.Errorf("api вернул %d: %s", resp.StatusCode, string(raw))
	}

	var result SyncResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return SyncResult{}, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

func parsePrice(raw string) (int, error) {
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}
	return int(f), nil
}

func ensureAbsoluteURL(href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	return "https://www.daangn.com" + href
}
