package models

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
