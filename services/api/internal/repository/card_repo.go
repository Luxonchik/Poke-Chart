package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"pokechart-api/internal/models"
)

// CardRepo будет хранить подключение к БД
type CardRepo struct {
	DB *sql.DB
}

// NewCardRepo - конструктор для нашего репозитория
func NewCardRepo(db *sql.DB) *CardRepo {
	return &CardRepo{DB: db}
}

// GetAllCards делает SELECT запрос и возвращает массив карточек (с лимитом для начала)
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

// GetAllCardsRankedByListings возвращает только те карточки, у которых есть
// хотя бы одно активное объявление на рынке, отсортированные по количеству (DESC).
func (r *CardRepo) GetAllCardsRankedByListings(limit int, offset int) ([]models.CardWithListingCount, error) {
	query := `
		SELECT mc.id, mc.name, mc.image_url, mc.card_no, mc.rarity, mc.series,
		       COUNT(ml.id) AS listing_count
		FROM master_cards mc
		INNER JOIN market_listings ml
			ON ml.master_card_id = mc.id AND ml.status = 'active'
		GROUP BY mc.id, mc.name, mc.image_url, mc.card_no, mc.rarity, mc.series
		ORDER BY listing_count DESC, mc.id
		LIMIT $1 OFFSET $2
	`

	rows, err := r.DB.Query(query, limit, offset)
	if err != nil {
		log.Println("❌ Ошибка при запросе карт с рейтингом:", err)
		return nil, err
	}
	defer rows.Close()

	var cards []models.CardWithListingCount

	for rows.Next() {
		var c models.CardWithListingCount
		err := rows.Scan(&c.ID, &c.Name, &c.ImageURL, &c.CardNo, &c.Rarity, &c.Series, &c.ListingCount)
		if err != nil {
			log.Println("❌ Ошибка при сканировании строки:", err)
			continue
		}
		cards = append(cards, c)
	}

	return cards, nil
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

// FindCardByNameAndRarity ищет официальную карту по имени и редкости
func (r *CardRepo) FindCardByNameAndRarity(name string, rarity string) (*models.Card, error) {
	// Пуленепробиваемый SQL-запрос:
	// 1. LOWER() - переводит всё в нижний регистр (EX -> ex)
	// 2. REPLACE(..., ' ', '') - удаляет все пробелы ("피카츄 ex" -> "피카츄ex")
	// То же самое мы делаем с редкостью (SAR -> sar), на всякий случай
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
