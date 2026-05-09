package repository

import (
	"database/sql"
	"pokechart-api/internal/models"
)

type ListingRepo struct {
	DB *sql.DB
}

func NewListingRepo(db *sql.DB) *ListingRepo {
	return &ListingRepo{DB: db}
}

// CreateListing пытается сохранить объявление. Если оно уже есть - просто игнорирует.
func (r *ListingRepo) CreateListing(req models.CreateListingRequest) (bool, error) {
	// Тот самый SQL запрос, который решает проблему дублей
	query := `
		INSERT INTO market_listings (
			external_id,
			title,
			price,
			image_url,
			url,
			status,
			created_at,
			boosted_at
		)
		VALUES ($1, $2, $3, $4, $5, 'pending', NULLIF($6, '')::timestamp, NULLIF($7, '')::timestamp)
		ON CONFLICT (external_id) DO NOTHING
	`

	result, err := r.DB.Exec(
		query,
		req.DaangnID,
		req.Title,
		req.Price,
		req.ImageURL,
		req.SourceURL,
		req.CreatedAt,
		req.BoostedAt,
	)
	if err != nil {
		return false, err
	}

	// Проверяем, вставилась ли новая строка
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	// Если запись уже была, дозаполняем недостающие поля.
	if rowsAffected == 0 {
		updateQuery := `
			UPDATE market_listings
			SET
				image_url = COALESCE(NULLIF(image_url, ''), NULLIF($2, '')),
				created_at = COALESCE(created_at, NULLIF($3, '')::timestamp),
				boosted_at = COALESCE(boosted_at, NULLIF($4, '')::timestamp)
			WHERE external_id = $1
		`
		if _, err := r.DB.Exec(updateQuery, req.DaangnID, req.ImageURL, req.CreatedAt, req.BoostedAt); err != nil {
			return false, err
		}
	}

	// Вернет true, если это новое объявление, и false, если дубликат
	return rowsAffected > 0, nil
}
