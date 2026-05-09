package models

import "time"

type MarketListing struct {
	ID           int       `json:"id"`
	DaangnID     string    `json:"daangn_id"`
	Title        string    `json:"title"`
	Price        int       `json:"price"`
	ImageURL     string    `json:"image_url"`
	SourceURL    string    `json:"source_url"`
	MasterCardID *int      `json:"master_card_id"` // Указатель, так как может быть nil (NULL в базе)
	Status       string    `json:"status"`
	AIErrorMsg   *string   `json:"ai_error_msg,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateListingRequest - то, что будет присылать  скрапер
type CreateListingRequest struct {
	DaangnID  string `json:"daangn_id"`
	Title     string `json:"title"`
	Price     int    `json:"price"`
	ImageURL  string `json:"image_url"`
	SourceURL string `json:"source_url"`
	CreatedAt string `json:"created_at"`
	BoostedAt string `json:"boosted_at"`
}
