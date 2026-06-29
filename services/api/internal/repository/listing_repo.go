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

func (r *ListingRepo) GetListingsByMasterCardID(masterCardID string) ([]models.MarketListing, error) {
	query := `
		SELECT id,
		       COALESCE(external_id, ''),
		       COALESCE(title, ''),
		       COALESCE(price, 0),
		       COALESCE(image_url, ''),
		       REPLACE(COALESCE(url, ''), 'https://www.daangn.comhttps://www.daangn.com', 'https://www.daangn.com'),
		       COALESCE(master_card_id, ''),
		       COALESCE(status, ''),
		       COALESCE(created_at, NOW()),
		       COALESCE(updated_at, NOW())
		FROM market_listings
		WHERE master_card_id = $1 AND status = 'active'
		ORDER BY price ASC
	`

	rows, err := r.DB.Query(query, masterCardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []models.MarketListing
	for rows.Next() {
		var l models.MarketListing
		var mcID string
		err := rows.Scan(
			&l.ID, &l.DaangnID, &l.Title, &l.Price,
			&l.ImageURL, &l.SourceURL, &mcID, &l.Status,
			&l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			continue
		}
		l.MasterCardID = &mcID
		listings = append(listings, l)
	}

	return listings, nil
}

// CreateListing вставляет объявление; при дубле (ON CONFLICT) дозаполняет пустые поля и возвращает false.
func (r *ListingRepo) CreateListing(req models.CreateListingRequest) (bool, error) {
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	// Дубль: дозаполняем поля, которые могли быть пустыми при первой вставке.
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

	return rowsAffected > 0, nil
}
