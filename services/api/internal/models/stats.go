package models

import "time"

// MarketStats — сводные цифры для полосы статистики на главной.
type MarketStats struct {
	CardsTracked   int        `json:"cards_tracked"`
	ActiveListings int        `json:"active_listings"`
	AIProcessed    int        `json:"ai_processed"`
	LastScrapedAt  *time.Time `json:"last_scraped_at"`
}
