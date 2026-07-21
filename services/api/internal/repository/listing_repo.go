package repository

import (
	"database/sql"

	"github.com/lib/pq"
	"pokechart-api/internal/models"
)

type ListingRepo struct {
	DB *sql.DB
}

func NewListingRepo(db *sql.DB) *ListingRepo {
	return &ListingRepo{DB: db}
}

func (r *ListingRepo) GetListingsByMasterCardID(masterCardID string) ([]models.MarketListing, error) {
	const q = `
		SELECT id,
		       COALESCE(source, 'daangn'),
		       COALESCE(external_id, ''),
		       COALESCE(title, ''),
		       COALESCE(price, 0),
		       COALESCE(image_url, ''),
		       COALESCE(url, ''),
		       COALESCE(master_card_id, ''),
		       COALESCE(status, ''),
		       COALESCE(created_at, NOW()),
		       COALESCE(updated_at, NOW())
		FROM market_listings
		WHERE master_card_id = $1 AND status = 'active'
		ORDER BY price ASC
	`
	rows, err := r.DB.Query(q, masterCardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []models.MarketListing
	for rows.Next() {
		var l models.MarketListing
		var mcID string
		if err := rows.Scan(
			&l.ID, &l.Source, &l.ExternalID, &l.Title, &l.Price,
			&l.ImageURL, &l.SourceURL, &mcID, &l.Status,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.MasterCardID = &mcID
		listings = append(listings, l)
	}
	return listings, rows.Err()
}

// SyncListings делает diff текущего снимка с БД:
//   - новые объявления → INSERT status='pending'
//   - существующие     → UPDATE цены и метаданных (статус не трогаем)
//   - исчезнувшие active → UPDATE status='closed'
func (r *ListingRepo) SyncListings(source string, listings []models.RawListing) (models.SyncResult, error) {
	var result models.SyncResult
	if len(listings) == 0 {
		return result, nil
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	externalIDs := make([]string, 0, len(listings))

	for _, l := range listings {
		externalIDs = append(externalIDs, l.ExternalID)

		var isNew bool
		err := tx.QueryRow(`
			INSERT INTO market_listings
			    (source, external_id, title, price, image_url, url, status, created_at, boosted_at)
			VALUES
			    ($1, $2, $3, $4, $5, $6, 'pending', NULLIF($7,'')::timestamp, NULLIF($8,'')::timestamp)
			ON CONFLICT (source, external_id) DO UPDATE SET
			    price      = EXCLUDED.price,
			    image_url  = COALESCE(NULLIF(EXCLUDED.image_url,''), market_listings.image_url),
			    boosted_at = COALESCE(EXCLUDED.boosted_at, market_listings.boosted_at),
			    updated_at = NOW()
			RETURNING (xmax = 0)
		`, source, l.ExternalID, l.Title, l.Price, l.ImageURL, l.SourceURL,
			l.CreatedAt, l.BoostedAt,
		).Scan(&isNew)
		if err != nil {
			return result, err
		}
		if isNew {
			result.New++
		} else {
			result.Updated++
		}
	}

	// Закрываем active-объявления, которых больше нет в свежем снимке
	res, err := tx.Exec(`
		UPDATE market_listings
		SET status = 'closed', closed_at = NOW(), updated_at = NOW()
		WHERE source = $1
		  AND status = 'active'
		  AND NOT (external_id = ANY($2))
	`, source, pq.Array(externalIDs))
	if err != nil {
		return result, err
	}

	closed, _ := res.RowsAffected()
	result.Closed = int(closed)

	return result, tx.Commit()
}
