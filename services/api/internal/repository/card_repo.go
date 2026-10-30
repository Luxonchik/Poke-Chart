package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"pokechart-api/internal/models"
)

type CardRepo struct {
	DB *sql.DB
}

func NewCardRepo(db *sql.DB) *CardRepo {
	return &CardRepo{DB: db}
}

func (r *CardRepo) GetAllCards(limit int, offset int) ([]models.Card, error) {
	query := `
		SELECT id, name, image_url, card_no, rarity, series
		FROM master_cards
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.DB.Query(query, limit, offset)
	if err != nil {
		log.Println("❌ Ошибка при запросе карт:", err)
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card

	for rows.Next() {
		var c models.Card
		err := rows.Scan(&c.ID, &c.Name, &c.ImageURL, &c.CardNo, &c.Rarity, &c.Series)
		if err != nil {
			log.Println("❌ Ошибка при сканировании строки:", err)
			continue
		}
		cards = append(cards, c)
	}

	return cards, nil
}

func (r *CardRepo) GetAllCardsRankedByListings(limit int, offset int) ([]models.CardWithListingCount, error) {
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.image_url, p.card_no, p.rarity, p.series, p.variant_count,
		       COUNT(ml.id) AS listing_count,
		       MIN(ml.price) AS min_price
		FROM (%s) p
		INNER JOIN market_listings ml
			ON ml.master_card_id = ANY(p.ids) AND ml.status = 'active'
		GROUP BY p.id, p.name, p.image_url, p.card_no, p.rarity, p.series, p.variant_count
		ORDER BY listing_count DESC, p.id
		LIMIT $1 OFFSET $2
	`, printingsSubquery(""))

	rows, err := r.DB.Query(query, limit, offset)
	if err != nil {
		log.Println("❌ Ошибка при запросе карт с рейтингом:", err)
		return nil, err
	}
	defer rows.Close()

	return scanPrintings(rows)
}

// GetCatalog returns every printing from master_cards (LEFT JOIN), cards with listings first.
// search filters by name or card number (case-insensitive substring); empty means no filter.
// onSaleOnly keeps only printings with at least one active listing.
func (r *CardRepo) GetCatalog(search string, onSaleOnly bool, limit int, offset int) ([]models.CardWithListingCount, error) {
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.image_url, p.card_no, p.rarity, p.series, p.variant_count,
		       COUNT(ml.id) AS listing_count,
		       MIN(ml.price) AS min_price
		FROM (%s) p
		LEFT JOIN market_listings ml
			ON ml.master_card_id = ANY(p.ids) AND ml.status = 'active'
		GROUP BY p.id, p.name, p.image_url, p.card_no, p.rarity, p.series, p.variant_count
		HAVING NOT $4 OR COUNT(ml.id) > 0
		ORDER BY listing_count DESC, p.id
		LIMIT $2 OFFSET $3
	`, printingsSubquery(`
		WHERE $1 = ''
		   OR name ILIKE '%' || $1 || '%' ESCAPE '\'
		   OR card_no ILIKE '%' || $1 || '%' ESCAPE '\'`))

	rows, err := r.DB.Query(query, escapeLike(strings.TrimSpace(search)), limit, offset, onSaleOnly)
	if err != nil {
		log.Println("❌ Ошибка при запросе каталога:", err)
		return nil, err
	}
	defer rows.Close()

	return scanPrintings(rows)
}

// printingsSubquery collapses master_cards rows that share name, card number and set
// into one "printing". The official list stores regular and mirror (holo) versions as
// separate ids with identical text, so without this they show up as duplicates.
// The lowest id represents the printing; ids holds every variant for joining listings.
func printingsSubquery(where string) string {
	return `
		SELECT (array_agg(id ORDER BY id))[1]        AS id,
		       name,
		       (array_agg(COALESCE(image_url, '') ORDER BY id))[1] AS image_url,
		       COALESCE(card_no, '')                AS card_no,
		       COALESCE(MAX(NULLIF(rarity, '')), '') AS rarity,
		       COALESCE(series, '')                 AS series,
		       array_agg(id)                        AS ids,
		       COUNT(*)                             AS variant_count
		FROM master_cards` + where + `
		GROUP BY name, COALESCE(card_no, ''), COALESCE(series, '')`
}

func scanPrintings(rows *sql.Rows) ([]models.CardWithListingCount, error) {
	var cards []models.CardWithListingCount
	for rows.Next() {
		var c models.CardWithListingCount
		if err := rows.Scan(
			&c.ID, &c.Name, &c.ImageURL, &c.CardNo, &c.Rarity, &c.Series, &c.VariantCount,
			&c.ListingCount, &c.MinPrice,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// escapeLike makes user input safe to use inside a LIKE pattern.
func escapeLike(s string) string {
	return likeEscaper.Replace(s)
}

func normalizeCardNo(cardNo string) string {
	s := strings.TrimSpace(cardNo)
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return ""
	}
	// Keep a single "/" separator.
	s = strings.ReplaceAll(s, "\\", "/")
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return s
	}
	leftRaw := strings.TrimSpace(parts[0])
	rightRaw := strings.TrimSpace(parts[1])

	left, errL := strconv.Atoi(leftRaw)
	right, errR := strconv.Atoi(rightRaw)
	if errL == nil && errR == nil {
		return fmt.Sprintf("%03d/%03d", left, right)
	}
	return leftRaw + "/" + rightRaw
}

func normalizeSeriesKey(series string) string {
	s := strings.ToLower(strings.TrimSpace(series))
	if s == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		" ", "",
		"「", "",
		"」", "",
		"\"", "",
		"'", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
	)
	s = replacer.Replace(s)
	return s
}

// FindCardByNameAndSeries narrows by a "core" series label (contains match).
func (r *CardRepo) FindCardByNameAndSeries(name string, seriesCore string) (*models.Card, error) {
	nameKey := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "")
	seriesKey := normalizeSeriesKey(seriesCore)
	if nameKey == "" || seriesKey == "" {
		return nil, fmt.Errorf("пустые параметры поиска (name/series)")
	}

	query := `
		SELECT id, name, image_url, card_no, rarity, series
		FROM master_cards
		WHERE REPLACE(LOWER(name), ' ', '') LIKE '%' || $1 || '%'
		  AND REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(LOWER(series), ' ', ''), '「', ''), '」', ''), '"', ''), '''', '') LIKE '%' || $2 || '%'
		LIMIT 1
	`

	var card models.Card
	err := r.DB.QueryRow(query, nameKey, seriesKey).Scan(
		&card.ID,
		&card.Name,
		&card.ImageURL,
		&card.CardNo,
		&card.Rarity,
		&card.Series,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// FindCandidateCardsByNameAndSeries returns a small list of candidates for disambiguation.
func (r *CardRepo) FindCandidateCardsByNameAndSeries(name string, seriesCore string, limit int) ([]models.Card, error) {
	nameKey := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "")
	seriesKey := normalizeSeriesKey(seriesCore)
	if nameKey == "" || seriesKey == "" {
		return nil, fmt.Errorf("пустые параметры поиска (name/series)")
	}
	if limit <= 0 {
		limit = 5
	}

	query := `
		SELECT id, name, image_url, card_no, rarity, series
		FROM master_cards
		WHERE REPLACE(LOWER(name), ' ', '') LIKE '%' || $1 || '%'
		  AND REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(LOWER(series), ' ', ''), '「', ''), '」', ''), '"', ''), '''', '') LIKE '%' || $2 || '%'
		LIMIT $3
	`
	rows, err := r.DB.Query(query, nameKey, seriesKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.Name, &c.ImageURL, &c.CardNo, &c.Rarity, &c.Series); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepo) FindCandidateCardsByName(name string, limit int) ([]models.Card, error) {
	nameKey := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "")
	if nameKey == "" {
		return nil, fmt.Errorf("пустое имя")
	}
	if limit <= 0 {
		limit = 5
	}
	query := `
		SELECT id, name, image_url, card_no, rarity, series
		FROM master_cards
		WHERE REPLACE(LOWER(name), ' ', '') LIKE '%' || $1 || '%'
		LIMIT $2
	`
	rows, err := r.DB.Query(query, nameKey, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.Name, &c.ImageURL, &c.CardNo, &c.Rarity, &c.Series); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// FindCardsByCardNo returns up to a few cards with the same card_no.
func (r *CardRepo) FindCardsByCardNo(cardNo string, limit int) ([]models.Card, error) {
	norm := normalizeCardNo(cardNo)
	if norm == "" {
		return nil, fmt.Errorf("пустой card_no")
	}
	if limit <= 0 {
		limit = 5
	}
	query := `
		SELECT id, name, image_url, card_no, rarity, series
		FROM master_cards
		WHERE REPLACE(LOWER(card_no), ' ', '') = REPLACE(LOWER($1), ' ', '')
		LIMIT $2
	`
	rows, err := r.DB.Query(query, norm, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.Name, &c.ImageURL, &c.CardNo, &c.Rarity, &c.Series); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// FindBestByCardNoThenName tries card_no first; if duplicates exist, it narrows by name.
func (r *CardRepo) FindBestByCardNoThenName(cardNo string, name string) (*models.Card, error) {
	candidates, err := r.FindCardsByCardNo(cardNo, 10)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, sql.ErrNoRows
	}
	if len(candidates) == 1 {
		return &candidates[0], nil
	}

	// Multiple cards share the same card_no; narrow by name.
	nameNorm := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "")
	if nameNorm == "" {
		return nil, fmt.Errorf("несколько карт с card_no=%s и пустое имя для уточнения", cardNo)
	}
	for _, c := range candidates {
		cn := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(c.Name)), " ", "")
		if strings.Contains(cn, nameNorm) || strings.Contains(nameNorm, cn) {
			cc := c
			return &cc, nil
		}
	}

	return nil, fmt.Errorf("несколько карт с card_no=%s, но имя '%s' не уточнило выбор", cardNo, name)
}

func (r *CardRepo) FindCardByNameAndRarity(name string, rarity string) (*models.Card, error) {
	query := `
		SELECT id, name, image_url, card_no, rarity, series 
		FROM master_cards 
		WHERE REPLACE(LOWER(name), ' ', '') LIKE '%' || REPLACE(LOWER($1), ' ', '') || '%' 
		AND UPPER(REPLACE(rarity, ' ', '')) = UPPER(REPLACE($2, ' ', ''))
		LIMIT 1
	`

	var card models.Card
	
	err := r.DB.QueryRow(query, name, rarity).Scan(
		&card.ID,
		&card.Name,
		&card.ImageURL,
		&card.CardNo,
		&card.Rarity,
		&card.Series,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("карта '%s' (%s) не найдена в master_cards", name, rarity)
		}
		return nil, err
	}

	return &card, nil
}
