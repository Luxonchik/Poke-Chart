package models

import "time"

type MarketListing struct {
	ID           int        `json:"id"`
	Source       string     `json:"source"`
	ExternalID   string     `json:"external_id"`
	Title        string     `json:"title"`
	Price        int        `json:"price"`
	ImageURL     string     `json:"image_url"`
	SourceURL    string     `json:"source_url"`
	MasterCardID *string    `json:"master_card_id"`
	Status       string     `json:"status"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	AIErrorMsg   *string    `json:"ai_error_msg,omitempty"`
	AICardNo     *string    `json:"ai_card_no,omitempty"`
	AISeries     *string    `json:"ai_series,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RawListing — данные одного объявления от скрапера (площадко-независимо).
type RawListing struct {
	ExternalID string `json:"external_id"`
	Title      string `json:"title"`
	Price      int    `json:"price"`
	ImageURL   string `json:"image_url"`
	SourceURL  string `json:"source_url"`
	CreatedAt  string `json:"created_at"`
	BoostedAt  string `json:"boosted_at"`
}

// SyncRequest — один батч от скрапера: площадка + все текущие объявления.
type SyncRequest struct {
	Source   string       `json:"source"`
	Listings []RawListing `json:"listings"`
}

// SyncResult — итог синхронизации.
type SyncResult struct {
	New     int `json:"new"`
	Updated int `json:"updated"`
	Closed  int `json:"closed"`
}
