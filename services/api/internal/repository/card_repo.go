package repository

import (
	"database/sql"
	"fmt"
	"log"

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
	// Берем с лимитом, чтобы не вывалить сразу 5000 карт в браузер
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

	// Создаем пустой слайс (массив) для карточек
	var cards []models.Card

	// Проходимся по каждой строке из базы
	for rows.Next() {
		var c models.Card
		// Записываем данные из SQL-строки прямо в нашу структуру
		err := rows.Scan(&c.ID, &c.Name, &c.ImageURL, &c.CardNo, &c.Rarity, &c.Series)
		if err != nil {
			log.Println("❌ Ошибка при сканировании строки:", err)
			continue
		}
		cards = append(cards, c)
	}

	return cards, nil
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
